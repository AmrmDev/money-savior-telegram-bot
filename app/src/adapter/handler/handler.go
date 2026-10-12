package handler

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"money-savior-telegram-bot/app/src/adapter/controller"
)

type Handler struct {
	startController *controller.StartController
	helpController  *controller.HelpController
}

func NewHandler(
	start *controller.StartController,
	help *controller.HelpController,
) *Handler {
	return &Handler{
		startController: start,
		helpController:  help,
	}
}

func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) error {
	if update.Message == nil || update.Message.From == nil {
		return nil
	}

	message := toMessage(update.Message)

	switch message.Command {
	case "start":
		return h.startController.Handle(ctx, message)
	default:
		return h.helpController.Handle(ctx, message)
	}
}

func toMessage(m *tgbotapi.Message) controller.Message {
	return controller.Message{
		ChatID:   m.Chat.ID,
		UserID:   m.From.ID,
		Username: m.From.UserName,
		Command:  m.Command(),
		Args:     strings.Fields(m.CommandArguments()),
	}
}
