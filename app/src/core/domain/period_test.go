package domain

import (
	"errors"
	"testing"
	"time"

	"money-savior-telegram-bot/app/src/core/domain/exception"
)

func utc(y int, m time.Month, d, h int) time.Time {
	return time.Date(y, m, d, h, 0, 0, 0, time.UTC)
}

func TestResolvePeriod(t *testing.T) {
	morning := utc(2026, 10, 13, 10)
	lateNight := utc(2026, 10, 13, 1)
	january := utc(2026, 1, 15, 12)

	tests := []struct {
		name  string
		now   time.Time
		spec  PeriodSpec
		start time.Time
		end   time.Time
	}{
		{"hoje", morning, PeriodSpec{Kind: PeriodToday}, utc(2026, 10, 13, 3), utc(2026, 10, 14, 3)},
		{"madrugada UTC ainda é o dia anterior em Brasília", lateNight, PeriodSpec{Kind: PeriodToday}, utc(2026, 10, 12, 3), utc(2026, 10, 13, 3)},
		{"ontem", morning, PeriodSpec{Kind: PeriodYesterday}, utc(2026, 10, 12, 3), utc(2026, 10, 13, 3)},
		{"últimos 7 dias", morning, PeriodSpec{Kind: PeriodLast7Days}, utc(2026, 10, 7, 3), utc(2026, 10, 14, 3)},
		{"mês atual", morning, PeriodSpec{Kind: PeriodThisMonth}, utc(2026, 10, 1, 3), utc(2026, 10, 14, 3)},
		{"mês passado", morning, PeriodSpec{Kind: PeriodLastMonth}, utc(2026, 9, 1, 3), utc(2026, 10, 1, 3)},
		{"mês passado em janeiro", january, PeriodSpec{Kind: PeriodLastMonth}, utc(2025, 12, 1, 3), utc(2026, 1, 1, 3)},
		{"intervalo", morning, PeriodSpec{Kind: PeriodCustom, From: Date{2026, 10, 1}, To: Date{2026, 10, 15}}, utc(2026, 10, 1, 3), utc(2026, 10, 16, 3)},
		{"intervalo sem ano usa o ano atual", morning, PeriodSpec{Kind: PeriodCustom, From: Date{0, 10, 1}, To: Date{0, 10, 15}}, utc(2026, 10, 1, 3), utc(2026, 10, 16, 3)},
		{"o ano é o de Brasília", utc(2027, 1, 1, 1), PeriodSpec{Kind: PeriodCustom, From: Date{0, 12, 30}, To: Date{0, 12, 31}}, utc(2026, 12, 30, 3), utc(2027, 1, 1, 3)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolvePeriod(tt.spec, tt.now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Start.Equal(tt.start) || !got.End.Equal(tt.end) {
				t.Errorf("got [%v, %v), want [%v, %v)", got.Start, got.End, tt.start, tt.end)
			}
		})
	}
}

func TestResolvePeriod_Invalid(t *testing.T) {
	now := utc(2026, 10, 13, 10)

	tests := []struct {
		name string
		spec PeriodSpec
	}{
		{"sem tipo", PeriodSpec{}},
		{"data inexistente", PeriodSpec{Kind: PeriodCustom, From: Date{2026, 2, 31}, To: Date{2026, 3, 5}}},
		{"fim antes do início", PeriodSpec{Kind: PeriodCustom, From: Date{2026, 10, 15}, To: Date{2026, 10, 1}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolvePeriod(tt.spec, now)
			if !errors.Is(err, exception.ErrInvalidPeriod) {
				t.Errorf("expected ErrInvalidPeriod, got %v", err)
			}
		})
	}

}
func TestPeriod_Days(t *testing.T) {
	period := Period{Start: utc(2026, 10, 1, 3), End: utc(2026, 10, 14, 3)}

	first, last := period.Days()

	if first.Format("2006-01-02") != "2026-10-01" || last.Format("2006-01-02") != "2026-10-13" {
		t.Errorf("unexpected days: %s to %s", first, last)
	}
}

func TestDateOf_UsesBrasiliaCalendar(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want Date
	}{
		{"02:00 UTC is still the previous day", utc(2026, 10, 13, 2), Date{2026, 10, 12}},
		{"03:00 UTC is midnight in Brasília", utc(2026, 10, 13, 3), Date{2026, 10, 13}},
		{"crosses the year boundary", utc(2027, 1, 1, 2), Date{2026, 12, 31}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DateOf(tt.in); got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDate_Weekday(t *testing.T) {
	if got := (Date{2026, 10, 19}).Weekday(); got != time.Monday {
		t.Errorf("19/10/2026 should be a Monday, got %v", got)
	}
	if got := (Date{2026, 10, 18}).Weekday(); got != time.Sunday {
		t.Errorf("18/10/2026 should be a Sunday, got %v", got)
	}
}

func TestDate_Before(t *testing.T) {
	tests := []struct {
		name string
		a, b Date
		want bool
	}{
		{"same day", Date{2026, 10, 12}, Date{2026, 10, 12}, false},
		{"earlier day", Date{2026, 10, 11}, Date{2026, 10, 12}, true},
		{"later day", Date{2026, 10, 13}, Date{2026, 10, 12}, false},
		{"earlier month", Date{2026, 9, 30}, Date{2026, 10, 1}, true},
		{"earlier year", Date{2025, 12, 31}, Date{2026, 1, 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Before(tt.b); got != tt.want {
				t.Errorf("%+v.Before(%+v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
