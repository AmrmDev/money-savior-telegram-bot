package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/domain/exception"
	"money-savior-telegram-bot/app/src/core/port"
	"money-savior-telegram-bot/app/src/core/usecase/input"
)

const maxSaveAttempts = 3

type RegisterExpense struct {
	repo port.ExpenseRepository
	ids  port.IDGenerator
}

func NewRegisterExpense(repo port.ExpenseRepository, ids port.IDGenerator) *RegisterExpense {
	return &RegisterExpense{repo: repo, ids: ids}
}

func (u *RegisterExpense) Execute(ctx context.Context, in input.RegisterExpense) (domain.Expense, error) {
	expense, err := domain.NewExpense(
		u.ids.NewExpenseID(),
		in.UserID,
		in.AmountCents,
		in.Category,
		in.Method,
		time.Now().UTC(),
	)
	if err != nil {
		return domain.Expense{}, err
	}

	for attempt := 1; ; attempt++ {
		err = u.repo.Save(ctx, expense)
		if !errors.Is(err, exception.ErrExpenseIDConflict) || attempt == maxSaveAttempts {
			break
		}
		expense.ID = u.ids.NewExpenseID()
	}

	if err != nil {
		return domain.Expense{}, fmt.Errorf("Save expense: %w", err)
	}
	return expense, nil
}