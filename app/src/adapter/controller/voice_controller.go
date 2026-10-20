package controller
import (
	"context"
	"fmt"
	"strings"
	"time"

	"money-savior-telegram-bot/app/src/adapter/presenter"
	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/port"
)

const (
	maxVoiceSeconds = 30
	voiceTimeout    = 40 * time.Second
)

type VoiceController struct {
	downloader  AudioDownloader
	transcriber port.Transcriber
	parser      port.IntentParser
	queries     MessageHandler
	sender      Sender
}

func NewVoiceController(
	downloader AudioDownloader,
	transcriber port.Transcriber,
	parser port.IntentParser,
	queries MessageHandler,
	sender Sender,
) *VoiceController {
	return &VoiceController{
		downloader:  downloader,
		transcriber: transcriber,
		parser:      parser,
		queries:     queries,
		sender:      sender,
	}
}

func (c *VoiceController) Handle(ctx context.Context, message Message) error {
	if message.VoiceFileID == "" {
		return reject(ctx, c.sender, message.ChatID, presenter.VoiceNotUnderstood(), "missing voice file")
	}
	if message.VoiceSeconds > maxVoiceSeconds {
		return reject(ctx, c.sender, message.ChatID, presenter.VoiceTooLong(maxVoiceSeconds), "voice too long")
	}

	execCtx, cancel := context.WithTimeout(ctx, voiceTimeout)
	defer cancel()

	audio, err := c.downloader.Download(execCtx, message.VoiceFileID)
	if err != nil {
		return respondError(ctx, c.sender, message.ChatID, presenter.VoiceError(), fmt.Errorf("download voice: %w", err))
	}

	text, err := c.transcriber.Transcribe(execCtx, audio)
	if err != nil {
		return respondError(ctx, c.sender, message.ChatID, presenter.VoiceError(), fmt.Errorf("transcribe voice: %w", err))
	}
	if strings.TrimSpace(text) == "" {
		return reject(ctx, c.sender, message.ChatID, presenter.VoiceNotUnderstood(), "empty transcript")
	}

	intent, err := c.parser.Parse(execCtx, text)
	if err != nil {
		return respondError(ctx, c.sender, message.ChatID, presenter.VoiceError(), fmt.Errorf("interpret voice: %w", err))
	}

	switch intent.Kind {
	case port.IntentQuery:
		return c.answerQuery(ctx, message, intent)
	case port.IntentExpense:
		return c.previewExpense(ctx, message, intent)
	default:
		return reject(ctx, c.sender, message.ChatID, presenter.VoiceNotUnderstood(), "unknown intent")
	}
}

func (c *VoiceController) answerQuery(ctx context.Context, message Message, intent port.Intent) error {
	args := periodArgs(intent)
	if _, err := parsePeriodSpec(args); err != nil {
		return reject(ctx, c.sender, message.ChatID, presenter.VoiceNotUnderstood(), "invalid period from voice")
	}

	return c.queries.Handle(ctx, Message{
		ChatID: message.ChatID,
		UserID: message.UserID,
		Args:   args,
	})
}

func (c *VoiceController) previewExpense(ctx context.Context, message Message, intent port.Intent) error {
	amountCents, err := parseAmountCents(intent.Amount)
	if err != nil {
		return reject(ctx, c.sender, message.ChatID, presenter.VoiceNotUnderstood(), "invalid amount from voice")
	}

	expense, err := domain.NewExpense("", message.UserID, amountCents, intent.Category, intent.Method, time.Time{})
	if err != nil {
		return reject(ctx, c.sender, message.ChatID, presenter.VoiceNotUnderstood(), "invalid expense from voice")
	}

	data, err := encodeConfirmCallback(expense.AmountCents, expense.Category, expense.Method)
	if err != nil {
		return reject(ctx, c.sender, message.ChatID, presenter.VoiceExpenseUseCommand(), "expense does not fit the confirmation button")
	}

	return c.sender.Send(ctx, message.ChatID, presenter.ExpensePreview(expense, data, cancelCallbackData))
}

func periodArgs(intent port.Intent) []string {
	keyword := strings.TrimSpace(strings.ReplaceAll(intent.PeriodKeyword, "_", " "))

	switch {
	case keyword != "":
		return strings.Fields(keyword)
	case intent.From != "" || intent.To != "":
		return []string{intent.From, intent.To}
	default:
		return nil
	}
}
