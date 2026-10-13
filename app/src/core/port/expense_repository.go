package port

import (
	"context"

	"money-savior-telegram-bot/app/src/core/domain"
)

type ExpenseRepository interface {
	Save(ctx context.Context, expense domain.Expense) error
	FindByID(ctx context.Context, userID int64, expenseID string) (*domain.Expense, error)
	FindAllByUser(ctx context.Context, userID int64) ([]domain.Expense, error)
	DeleteByID(ctx context.Context, userID int64, expenseID string) error
	DeleteAllByUser(ctx context.Context, userID int64) error
}