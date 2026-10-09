package domain

import "time"

type Expense struct {
	ID        string
	UserID    int64
	Amount    float64
	Category  string
	Method    string
	CreatedAt time.Time
}