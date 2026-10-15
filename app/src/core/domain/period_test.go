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
