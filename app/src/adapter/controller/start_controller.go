package controller

import (
    "context"

    "money-savior-telegram-bot/app/src/adapter/presenter"
)

type StartController struct {
    sender Sender
}

func NewStartController(sender Sender) *StartController {
    return &StartController{
        sender: sender,
    }
}

func (c *StartController) Handle(
    ctx context.Context,
    message Message,
) error {
    reply := presenter.Reply{
        ChatID: message.ChatID,
        Text:   welcomeMessage(message.Username),
    }

    return c.sender.Send(ctx, message.ChatID, reply)
}

func welcomeMessage(username string) string {
    if username == "" {
        return "Olá! 👋 Bem-vindo ao Money Savior!"
    }

    return "Olá, @" + username + "! 👋 Bem-vindo ao Money Savior!"
}