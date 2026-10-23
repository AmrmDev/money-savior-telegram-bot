package http

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"money-savior-telegram-bot/app/src/adapter/controller"
	"money-savior-telegram-bot/app/src/adapter/presenter"
)

type TelegramSender struct {
	bot *tgbotapi.BotAPI
}

var (
	_ controller.Sender            = (*TelegramSender)(nil)
	_ controller.CallbackResponder = (*TelegramSender)(nil)
)

func NewTelegramSender(bot *tgbotapi.BotAPI) *TelegramSender {
	return &TelegramSender{bot: bot}
}

func (s *TelegramSender) Send(_ context.Context, chatID int64, reply presenter.Reply) error {
	if len(reply.Image) > 0 {
		if _, err := s.bot.Send(buildPhoto(chatID, reply)); err != nil {
			return fmt.Errorf("send telegram photo: %w", err)
		}
		return nil
	}

	if _, err := s.bot.Send(buildMessage(chatID, reply)); err != nil {
		return fmt.Errorf("send telegram message: %w", err)
	}
	return nil
}

func buildMessage(chatID int64, reply presenter.Reply) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, reply.Text)
	msg.ReplyMarkup = replyMarkup(reply)
	return msg
}

func buildPhoto(chatID int64, reply presenter.Reply) tgbotapi.PhotoConfig {
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{Name: "image.png", Bytes: reply.Image})
	photo.Caption = reply.Text
	photo.ReplyMarkup = replyMarkup(reply)
	return photo
}

func replyMarkup(reply presenter.Reply) any {
	if len(reply.Buttons) > 0 {
		row := make([]tgbotapi.InlineKeyboardButton, 0, len(reply.Buttons))
		for _, button := range reply.Buttons {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(button.Label, button.Data))
		}
		return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(row...))
	}

	if reply.ReplyPlaceHolder != "" {
		return tgbotapi.ForceReply{ForceReply: true, InputFieldPlaceholder: reply.ReplyPlaceHolder}
	}

	return nil
}

func (s *TelegramSender) Edit(_ context.Context, chatID int64, messageID int, reply presenter.Reply) error {
	if _, err := s.bot.Send(tgbotapi.NewEditMessageText(chatID, messageID, reply.Text)); err != nil {
		return fmt.Errorf("edit telegram message: %w", err)
	}
	return nil
}

func (s *TelegramSender) AnswerCallback(_ context.Context, callbackID string) error {
	if _, err := s.bot.Request(tgbotapi.NewCallback(callbackID, "")); err != nil {
		return fmt.Errorf("answer telegram callback: %w", err)
	}
	return nil
}

func inlineKeyboard(buttons []presenter.Button) tgbotapi.InlineKeyboardMarkup {
	row := make([]tgbotapi.InlineKeyboardButton, len(buttons))
	for i, button := range buttons {
		row[i] = tgbotapi.NewInlineKeyboardButtonData(button.Label, button.Data)
	}
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(row...))
}
