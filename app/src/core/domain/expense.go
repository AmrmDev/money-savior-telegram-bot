package domain

import (
	"time"

	"money-savior-telegram-bot/app/src/core/domain/exception"
)

type Expense struct {
	ID          string
	UserID      int64
	AmountCents int64
	Category    string
	Method      string
	CreatedAt   time.Time
}

func NewExpense(
	id string,
	userID int64,
	amountCents int64,
	category string,
	method string,
	createdAt time.Time,
) (Expense, error) {
	if amountCents <= 0 {
		return Expense{}, exception.ErrInvalidAmount
	}

	category = formatTitle(category)
	if category == "" {
		return Expense{}, exception.ErrInvalidCategory
	}

	return Expense{
		ID:          id,
		UserID:      userID,
		AmountCents: amountCents,
		Category:    category,
		Method:      normalizeMethod(method),
		CreatedAt:   createdAt,
	}, nil
}