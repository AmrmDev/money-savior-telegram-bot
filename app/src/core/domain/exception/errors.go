package exception

import "errors"

var (
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidCategory   = errors.New("invalid category")
	ErrExpenseNotFound   = errors.New("expense not found")
	ErrExpenseIDConflict = errors.New("expense id already exists")
	ErrInvalidPeriod     = errors.New("invalid period")
)
