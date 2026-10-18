package repository

import (
	"testing"
	"time"

	"money-savior-telegram-bot/app/src/core/domain"
)

func TestFormatCreatedAt_IsFixedWidthUTC(t *testing.T) {
	brasilia := time.FixedZone("BRT", -3*60*60)

	tests := []struct {
		name string
		in   time.Time
		want string
	}{
		{"sem fração", time.Date(2026, 10, 13, 10, 39, 25, 0, time.UTC), "2026-10-13T10:39:25.000Z"},
		{"com fração", time.Date(2026, 10, 13, 10, 39, 25, 489_000_000, time.UTC), "2026-10-13T10:39:25.489Z"},
		{"converte para UTC", time.Date(2026, 10, 13, 7, 39, 25, 0, brasilia), "2026-10-13T10:39:25.000Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatCreatedAt(tt.in); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestFormatCreatedAt_StringOrderMatchesTimeOrder(t *testing.T) {
	earlier := time.Date(2026, 10, 13, 10, 39, 25, 0, time.UTC)
	later := earlier.Add(500 * time.Millisecond)

	if formatCreatedAt(earlier) >= formatCreatedAt(later) {
		t.Error("expected earlier timestamp to sort before the later one")
	}
}

func TestExpenseItem_RoundTrip(t *testing.T) {
	original := domain.Expense{
		ID:          "399MSB403PA",
		UserID:      1,
		AmountCents: 5090,
		Category:    "Mercado",
		Method:      "Pix",
		CreatedAt:   time.Date(2026, 10, 13, 10, 39, 25, 489_000_000, time.UTC),
	}

	got, err := toItem(original).toDomain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != original.ID || got.UserID != original.UserID ||
		got.AmountCents != original.AmountCents || got.Category != original.Category ||
		got.Method != original.Method || !got.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, original)
	}
}

func TestToDomain_AcceptsLegacyNanosecondTimestamp(t *testing.T) {
	item := expenseItem{ExpenseID: "399MSB403PA", CreatedAt: "2026-10-13T10:39:25.4890692Z"}

	if _, err := item.toDomain(); err != nil {
		t.Errorf("legacy timestamp should still be readable: %v", err)
	}
}
