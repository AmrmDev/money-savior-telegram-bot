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

var _ controller.Sender = (*TelegramSender)(nil)

func NewTelegramSender(bot *tgbotapi.BotAPI) *TelegramSender {
	return &TelegramSender{bot: bot}
}

func (s *TelegramSender) Send(_ context.Context, chatID int64, reply presenter.Reply) error {
	msg := tgbotapi.NewMessage(chatID, reply.Text)
	msg.ParseMode = tgbotapi.ModeHTML

	if _, err := s.bot.Send(tgbotapi.NewMessage(chatID, reply.Text)); err != nil {
		return fmt.Errorf("send telegram message: %w", err)
	}
	return nil
}
