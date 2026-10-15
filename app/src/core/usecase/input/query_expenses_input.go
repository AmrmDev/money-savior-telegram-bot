package input

import "money-savior-telegram-bot/app/src/core/domain"

type QueryExpenses struct {
	UserID int64
	Period domain.PeriodSpec
}