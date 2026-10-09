package controller

import (
	"context"

	"money-savior-telegram-bot/app/src/adapter/presenter"
)

type Sender interface {
	Send(ctx context.Context, chatID int64, reply presenter.Reply) error
}