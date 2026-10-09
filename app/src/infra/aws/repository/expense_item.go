package repository

import (
	"time"

	"money-savior-telegram-bot/app/src/core/domain"
)

type expenseItem struct {
	UserID      int64     `dynamodbav:"user_id"`
	ExpenseID   string    `dynamodbav:"expense_id"`
	AmountCents int64     `dynamodbav:"amount_cents"`
	Category    string    `dynamodbav:"category"`
	Method      string    `dynamodbav:"method"`
	CreatedAt   time.Time `dynamodbav:"created_at"`
}

func toItem(e domain.Expense) expenseItem {
	return expenseItem{
		UserID:      e.UserID,
		ExpenseID:   e.ID,
		AmountCents: e.AmountCents,
		Category:    e.Category,
		Method:      e.Method,
		CreatedAt:   e.CreatedAt,
	}
}

func (i expenseItem) toDomain() domain.Expense {
	return domain.Expense{
		ID:          i.ExpenseID,
		UserID:      i.UserID,
		AmountCents: i.AmountCents,
		Category:    i.Category,
		Method:      i.Method,
		CreatedAt:   i.CreatedAt,
	}
}