package controller

import (
	"context"

	"money-savior-telegram-bot/app/src/adapter/presenter"
)

type CallbackResponder interface {
	Edit(ctx context.Context, chatID int64, messageID int, reply presenter.Reply) error

	AnswerCallback(ctx context.Context, callbackID string) error
}
