package domain

import "time"

type Expense struct {
	ID        string
	UserID    int64
	AmountCents    int64
	Category  string
	Method    string
	CreatedAt time.Time
}