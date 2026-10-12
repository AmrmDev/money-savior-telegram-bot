package controller

import (
	"context"

	"money-savior-telegram-bot/app/src/adapter/presenter"
)

type HelpController struct {
	sender Sender
}

func NewHelpController(sender Sender) *HelpController {
	return &HelpController{sender: sender}
}

func (c *HelpController) Handle(ctx context.Context, message Message) error {
	return c.sender.Send(ctx, message.ChatID, presenter.Help())
}
