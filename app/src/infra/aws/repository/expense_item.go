package repository

import (
	"fmt"
	"time"

	"money-savior-telegram-bot/app/src/core/domain"
)

//fixed-width UTC timestamp: string order matches chronological order
//(RFC3339Nano trims trailing zeros, which breaks lexicographic ordering)
const createdAtLayout = "2006-01-02T15:04:05.000Z"

type expenseItem struct {
	UserID      int64  `dynamodbav:"user_id"`
	ExpenseID   string `dynamodbav:"expense_id"`
	AmountCents int64  `dynamodbav:"amount_cents"`
	Category    string `dynamodbav:"category"`
	Method      string `dynamodbav:"method"`
	CreatedAt   string `dynamodbav:"created_at"`
}

func formatCreatedAt(t time.Time) string {
	return t.UTC().Format(createdAtLayout)
}

func toItem(e domain.Expense) expenseItem {
	return expenseItem{
		UserID:      e.UserID,
		ExpenseID:   e.ID,
		AmountCents: e.AmountCents,
		Category:    e.Category,
		Method:      e.Method,
		CreatedAt:   formatCreatedAt(e.CreatedAt),
	}
}

func (i expenseItem) toDomain() (domain.Expense, error) {
	createdAt, err := time.Parse(time.RFC3339Nano, i.CreatedAt)
	if err != nil {
		return domain.Expense{}, fmt.Errorf("parse created_at %q: %w", i.CreatedAt, err)
	}

	return domain.Expense{
		ID:          i.ExpenseID,
		UserID:      i.UserID,
		AmountCents: i.AmountCents,
		Category:    i.Category,
		Method:      i.Method,
		CreatedAt:   createdAt,
	}, nil
}