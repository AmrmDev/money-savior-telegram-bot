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

type DayTotal struct {
	Date       domain.Date
	TotalCents int64
	Count      int
}

type ExpensesSummary struct {
	Period     domain.Period
	TotalCents int64
	Count      int
	Categories []CategoryTotal
	// PeakDay is the day with the highest spending, nil when there are no expenses.
	PeakDay *DayTotal
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
	byDay := make(map[domain.Date]*DayTotal)

	for _, e := range expenses {
		summary.TotalCents += e.AmountCents

		category, ok := byCategory[e.Category]
		if !ok {
			category = &CategoryTotal{Category: e.Category}
			byCategory[e.Category] = category
		}
		category.TotalCents += e.AmountCents
		category.Count++

		date := domain.DateOf(e.CreatedAt)
		day, ok := byDay[date]
		if !ok {
			day = &DayTotal{Date: date}
			byDay[date] = day
		}
		day.TotalCents += e.AmountCents
		day.Count++
	}

	for _, category := range byCategory {
		summary.Categories = append(summary.Categories, *category)
	}
	sort.Slice(summary.Categories, func(i, j int) bool {
		a, b := summary.Categories[i], summary.Categories[j]
		if a.TotalCents != b.TotalCents {
			return a.TotalCents > b.TotalCents
		}
		return a.Category < b.Category
	})

	summary.PeakDay = peakDay(byDay)

	return summary
}

// peakDay picks the day with the highest total. On a tie the earlier date wins,
// so the result never depends on map iteration order.
func peakDay(days map[domain.Date]*DayTotal) *DayTotal {
	var peak *DayTotal
	for _, day := range days {
		if peak == nil ||
			day.TotalCents > peak.TotalCents ||
			(day.TotalCents == peak.TotalCents && day.Date.Before(peak.Date)) {
			peak = day
		}
	}
	return peak
}
