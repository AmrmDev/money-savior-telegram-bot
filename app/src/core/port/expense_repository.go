package port

import (
	"context"

	"money-savior-telegram-bot/app/src/core/domain"
)

type ExpenseRepository interface {
	Save(ctx context.Context, expense domain.Expense) error
}
