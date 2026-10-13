package domain

import (
	"time"

	"money-savior-telegram-bot/app/src/core/domain/exception"
)

var brasiliaZone = time.FixedZone("BRT", -3*60*60)

type PeriodKind int

const (
	PeriodToday PeriodKind = iota + 1
	PeriodYesterday
	PeriodLast7Days
	PeriodThisMonth
	PeriodLastMonth
	PeriodCustom
)

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

type PeriodSpec struct {
	Kind PeriodKind
	From Date
	To   Date
}

type Period struct {
	Start time.Time
	End   time.Time
}

func ResolvePeriod(spec PeriodSpec, now time.Time) (Period, error) {
	local := now.In(brasiliaZone)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, brasiliaZone)
	tomorrow := today.AddDate(0, 0, 1)
	monthStart := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, brasiliaZone)

	switch spec.Kind {
	case PeriodToday:
		return newPeriod(today, tomorrow), nil
	case PeriodYesterday:
		return newPeriod(today.AddDate(0, 0, -1), today), nil
	case PeriodLast7Days:
		return newPeriod(today.AddDate(0, 0, -6), tomorrow), nil
	case PeriodThisMonth:
		return newPeriod(monthStart, tomorrow), nil
	case PeriodLastMonth:
		return newPeriod(monthStart.AddDate(0, -1, 0), monthStart), nil
	case PeriodCustom:
		from, okFrom := spec.From.startOfDay()
		to, okTo := spec.To.startOfDay()
		if !okFrom || !okTo || to.Before(from) {
			return Period{}, exception.ErrInvalidPeriod
		}
		return newPeriod(from, to.AddDate(0, 0, 1)), nil
	default:
		return Period{}, exception.ErrInvalidPeriod
	}
}

func newPeriod(start, end time.Time) Period {
	return Period{Start: start.UTC(), End: end.UTC()}
}

func (d Date) startOfDay() (time.Time, bool) {
	t := time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, brasiliaZone)
	if t.Year() != d.Year || t.Month() != d.Month || t.Day() != d.Day {
		return time.Time{}, false
	}
	return t, true
}