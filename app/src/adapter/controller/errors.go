package controller

import (
	"context"
	"errors"
	"fmt"

	"money-savior-telegram-bot/app/src/adapter/presenter"
	"money-savior-telegram-bot/app/src/core/domain/exception"
)

// ErrInvalidInput marks a request rejected because of what the user sent.
// The user was already answered, so this is not a failure of the bot.
var ErrInvalidInput = errors.New("invalid input")

// reject answers the user and reports the input as invalid.
// The reason goes to the logs, so it must be a fixed string and never carry user data.
func reject(ctx context.Context, sender Sender, chatID int64, reply presenter.Reply, reason string) error {
	if err := sender.Send(ctx, chatID, reply); err != nil {
		return err
	}
	return fmt.Errorf("%w: %s", ErrInvalidInput, reason)
}

// respondError answers the user after a use case error and returns what should be logged:
// invalid input when the user is at fault, the original error otherwise.
func respondError(ctx context.Context, sender Sender, chatID int64, reply presenter.Reply, err error) error {
	if isInvalidInput(err) {
		return reject(ctx, sender, chatID, reply, err.Error())
	}

	if sendErr := sender.Send(ctx, chatID, reply); sendErr != nil {
		return errors.Join(err, sendErr)
	}
	return err
}

func isInvalidInput(err error) bool {
	return errors.Is(err, exception.ErrInvalidAmount) ||
		errors.Is(err, exception.ErrInvalidCategory) ||
		errors.Is(err, exception.ErrInvalidPeriod)
}
