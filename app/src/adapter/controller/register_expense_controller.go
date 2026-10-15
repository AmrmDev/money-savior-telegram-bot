package controller

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"money-savior-telegram-bot/app/src/adapter/presenter"
	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/usecase/input"
)

const registerExpenseTimeout = 5 * time.Second

var errInvalidAmountFormat = errors.New("invalid amount format")

type RegisterExpenseUseCase interface {
	Execute(ctx context.Context, input input.RegisterExpense) (domain.Expense, error)
}

type RegisterExpenseController struct {
	usecase RegisterExpenseUseCase
	sender  Sender
}

func NewRegisterExpenseController(usecase RegisterExpenseUseCase, sender Sender) *RegisterExpenseController {
	return &RegisterExpenseController{usecase: usecase, sender: sender}
}

func (c *RegisterExpenseController) Handle(ctx context.Context, message Message) error {
	if len(message.Args) < 2 {
		return c.sender.Send(ctx, message.ChatID, presenter.InvalidExpenseFormat())
	}

	amountCents, err := parseAmountCents(message.Args[0])
	if err != nil {
		return c.sender.Send(ctx, message.ChatID, presenter.InvalidAmount())
	}

	in := input.RegisterExpense{
		UserID:      message.UserID,
		AmountCents: amountCents,
		Category:    message.Args[1],
	}
	if len(message.Args) > 2 {
		in.Method = message.Args[2]
	}

	execCtx, cancel := context.WithTimeout(ctx, registerExpenseTimeout)
	defer cancel()

	expense, err := c.usecase.Execute(execCtx, in)
	if err != nil {
		if sendErr := c.sender.Send(ctx, message.ChatID, presenter.RegisterExpenseError(err)); sendErr != nil {
			return sendErr
		}
		return err
	}

	return c.sender.Send(ctx, message.ChatID, presenter.ExpenseRegistered(expense))
}

func parseAmountCents(text string) (int64, error) {
	text = strings.Replace(text, ",", ".", 1)

	whole, fraction, hasFraction := strings.Cut(text, ".")
	if !isDigits(whole) {
		return 0, errInvalidAmountFormat
	}
	if hasFraction && (len(fraction) > 2 || !isDigits(fraction)) {
		return 0, errInvalidAmountFormat
	}

	for len(fraction) < 2 {
		fraction += "0"
	}

	cents, err := strconv.ParseInt(whole+fraction, 10, 64)
	if err != nil {
		return 0, errInvalidAmountFormat
	}
	return cents, nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
