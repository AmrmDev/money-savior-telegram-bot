package controller

import (
	"context"

	"money-savior-telegram-bot/app/src/adapter/presenter"
)

type StartController struct {
	sender Sender
}

func NewStartController(sender Sender) *StartController {
	return &StartController{sender: sender}
}

func (c *StartController) Handle(ctx context.Context, message Message) error {
	return c.sender.Send(ctx, message.ChatID, presenter.Welcome(message.Username))
}
