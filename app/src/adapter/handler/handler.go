package handler

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"money-savior-telegram-bot/app/src/adapter/controller"
)

type Handler struct {
	startController           *controller.StartController
	helpController            *controller.HelpController
	registerExpenseController *controller.RegisterExpenseController
	queryExpensesController   *controller.QueryExpensesController
	voiceController           *controller.VoiceController
	confirmController         *controller.ConfirmExpenseController
	allowedUserID             int64
	logger                    *slog.Logger
}

// NewHandler builds the update router. When allowedUserID is not zero, updates from any
// other user are ignored in silence.
func NewHandler(
	start *controller.StartController,
	help *controller.HelpController,
	registerExpense *controller.RegisterExpenseController,
	queryExpenses *controller.QueryExpensesController,
	voice *controller.VoiceController,
	confirm *controller.ConfirmExpenseController,
	allowedUserID int64,
	logger *slog.Logger,
) *Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return &Handler{
		startController:           start,
		helpController:            help,
		registerExpenseController: registerExpense,
		queryExpensesController:   queryExpenses,
		voiceController:           voice,
		confirmController:         confirm,
		allowedUserID:             allowedUserID,
		logger:                    logger,
	}
}

// Handle processes one update. Failures are logged and never returned: on a webhook,
// a non-200 answer makes Telegram resend the update and could register an expense twice.
func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		h.handleCallback(ctx, update)
		return
	}

	if update.Message == nil || update.Message.From == nil {
		h.logger.DebugContext(ctx, "update ignored", slog.Int("update_id", update.UpdateID))
		return
	}

	if !h.allowed(update.Message.From.ID) {
		h.logger.DebugContext(ctx, "update from a user that is not allowed",
			slog.Int("update_id", update.UpdateID),
			slog.Int64("user_id", update.Message.From.ID),
		)
		return
	}

	message := toMessage(update.Message)

	startedAt := time.Now()
	command, err := h.dispatch(ctx, message)
	h.logResult(ctx, update.UpdateID, message.UserID, command, time.Since(startedAt), err)
}

// handleCallback processes a pressed button. The user check comes first, as for messages.
// The button data is never logged: it carries the amount and the category.
func (h *Handler) handleCallback(ctx context.Context, update tgbotapi.Update) {
	query := update.CallbackQuery

	if query.From == nil {
		h.logger.DebugContext(ctx, "update ignored", slog.Int("update_id", update.UpdateID))
		return
	}

	if !h.allowed(query.From.ID) {
		h.logger.DebugContext(ctx, "update from a user that is not allowed",
			slog.Int("update_id", update.UpdateID),
			slog.Int64("user_id", query.From.ID),
		)
		return
	}

	callback := controller.Callback{ID: query.ID, UserID: query.From.ID, Data: query.Data}
	if query.Message != nil && query.Message.Chat != nil {
		callback.ChatID = query.Message.Chat.ID
		callback.MessageID = query.Message.MessageID
	}

	startedAt := time.Now()
	err := h.confirmController.Handle(ctx, callback)
	h.logResult(ctx, update.UpdateID, callback.UserID, "callback", time.Since(startedAt), err)
}

func (h *Handler) allowed(userID int64) bool {
	return h.allowedUserID == 0 || userID == h.allowedUserID
}

func (h *Handler) dispatch(ctx context.Context, message controller.Message) (string, error) {
	if message.VoiceFileID != "" {
		return "voz", h.voiceController.Handle(ctx, message)
	}

	switch message.Command {
	case "start":
		return "start", h.startController.Handle(ctx, message)
	case "gastei":
		return "gastei", h.registerExpenseController.Handle(ctx, message)
	case "consulta":
		return "consulta", h.queryExpensesController.Handle(ctx, message)
	default:
		return "help", h.helpController.Handle(ctx, message)
	}
}

func (h *Handler) logResult(ctx context.Context, updateID int, userID int64, command string, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.Int("update_id", updateID),
		slog.Int64("user_id", userID),
		slog.String("command", command),
		slog.Int64("duration_ms", duration.Milliseconds()),
	}

	switch {
	case err == nil:
		attrs = append(attrs, slog.String("outcome", "ok"))
		h.logger.LogAttrs(ctx, slog.LevelInfo, "update handled", attrs...)
	case errors.Is(err, controller.ErrInvalidInput):
		attrs = append(attrs, slog.String("outcome", "invalid_input"), slog.String("reason", err.Error()))
		h.logger.LogAttrs(ctx, slog.LevelWarn, "update rejected", attrs...)
	default:
		attrs = append(attrs, slog.String("outcome", "error"), slog.String("error", err.Error()))
		h.logger.LogAttrs(ctx, slog.LevelError, "update failed", attrs...)
	}
}

func toMessage(m *tgbotapi.Message) controller.Message {
	message := controller.Message{
		ChatID:   m.Chat.ID,
		UserID:   m.From.ID,
		Username: m.From.UserName,
		Command:  m.Command(),
		Args:     strings.Fields(m.CommandArguments()),
	}

	if m.Voice != nil {
		message.VoiceFileID = m.Voice.FileID
		message.VoiceSeconds = m.Voice.Duration
	}
	return message
}
