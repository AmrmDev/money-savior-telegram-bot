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
		return s.sendPhoto(chatID, reply)
	}

	message := tgbotapi.NewMessage(chatID, reply.Text)
	if len(reply.Buttons) > 0 {
		message.ReplyMarkup = inlineKeyboard(reply.Buttons)
	}

	if _, err := s.bot.Send(message); err != nil {
		return fmt.Errorf("send telegram message: %w", err)
	}
	return nil
}

func (s *TelegramSender) sendPhoto(chatID int64, reply presenter.Reply) error {
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{Name: "summary.png", Bytes: reply.Image})
	photo.Caption = reply.Text
	if len(reply.Buttons) > 0 {
		photo.ReplyMarkup = inlineKeyboard(reply.Buttons)
	}

	if _, err := s.bot.Send(photo); err != nil {
		return fmt.Errorf("send telegram photo: %w", err)
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
