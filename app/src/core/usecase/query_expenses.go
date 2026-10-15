package usecase

import (
	"context"
	"fmt"
	"sort"
	"time"

	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/port"
	"money-savior-telegram-bot/app/src/core/usecase/input"
)

type CategoryTotal struct {
	Category   string
	TotalCents int64
	Count      int
}

type ExpensesSummary struct {
	Period     domain.Period
	TotalCents int64
	Count      int
	Categories []CategoryTotal
}

type QueryExpenses struct {
	reader port.ExpenseReader
	now    func() time.Time
}

func NewQueryExpenses(reader port.ExpenseReader) *QueryExpenses {
	return &QueryExpenses{reader: reader, now: time.Now}
}

func (u *QueryExpenses) Execute(ctx context.Context, in input.QueryExpenses) (ExpensesSummary, error) {
	period, err := domain.ResolvePeriod(in.Period, u.now())
	if err != nil {
		return ExpensesSummary{}, err
	}

	expenses, err := u.reader.FindByPeriod(ctx, in.UserID, period)
	if err != nil {
		return ExpensesSummary{}, fmt.Errorf("find expenses by period: %w", err)
	}

	return summarize(period, expenses), nil
}

func summarize(period domain.Period, expenses []domain.Expense) ExpensesSummary {
	summary := ExpensesSummary{Period: period, Count: len(expenses)}

	byCategory := make(map[string]*CategoryTotal)
	for _, e := range expenses {
		summary.TotalCents += e.AmountCents

		total, ok := byCategory[e.Category]
		if !ok {
			total = &CategoryTotal{Category: e.Category}
			byCategory[e.Category] = total
		}
		total.TotalCents += e.AmountCents
		total.Count++
	}

	for _, total := range byCategory {
		summary.Categories = append(summary.Categories, *total)
	}
	sort.Slice(summary.Categories, func(i, j int) bool {
		a, b := summary.Categories[i], summary.Categories[j]
		if a.TotalCents != b.TotalCents {
			return a.TotalCents > b.TotalCents
		}
		return a.Category < b.Category
	})

	return summary
}
