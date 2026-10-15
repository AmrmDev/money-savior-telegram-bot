package port

import (
	"context"

	"money-savior-telegram-bot/app/src/core/domain"
)

type ExpenseReader interface {
	FindByPeriod(ctx context.Context, userID int64, period domain.Period) ([]domain.Expense, error)
}
