package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/domain/exception"
	"money-savior-telegram-bot/app/src/core/usecase/input"
)

type fakeReader struct {
	expenses  []domain.Expense
	err       error
	called    bool
	gotUser   int64
	gotPeriod domain.Period
}

func (f *fakeReader) FindByPeriod(_ context.Context, userID int64, period domain.Period) ([]domain.Expense, error) {
	f.called = true
	f.gotUser = userID
	f.gotPeriod = period
	return f.expenses, f.err
}

func newQueryUseCase(reader *fakeReader) *QueryExpenses {
	uc := NewQueryExpenses(reader)
	uc.now = func() time.Time { return time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC) }
	return uc
}

func TestQueryExpenses_SummarizesByCategory(t *testing.T) {
	reader := &fakeReader{expenses: []domain.Expense{
		{Category: "Mercado", AmountCents: 5000},
		{Category: "Uber", AmountCents: 1500},
		{Category: "Mercado", AmountCents: 2500},
	}}

	got, err := newQueryUseCase(reader).Execute(context.Background(), input.QueryExpenses{
		UserID: 1,
		Period: domain.PeriodSpec{Kind: domain.PeriodToday},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalCents != 9000 || got.Count != 3 {
		t.Errorf("unexpected totals: %+v", got)
	}
	if len(got.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(got.Categories))
	}
	if got.Categories[0] != (CategoryTotal{Category: "Mercado", TotalCents: 7500, Count: 2}) {
		t.Errorf("unexpected first category: %+v", got.Categories[0])
	}
	if got.Categories[1] != (CategoryTotal{Category: "Uber", TotalCents: 1500, Count: 1}) {
		t.Errorf("unexpected second category: %+v", got.Categories[1])
	}
}

func TestQueryExpenses_NoExpenses(t *testing.T) {
	reader := &fakeReader{}

	got, err := newQueryUseCase(reader).Execute(context.Background(), input.QueryExpenses{
		UserID: 1,
		Period: domain.PeriodSpec{Kind: domain.PeriodToday},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalCents != 0 || got.Count != 0 || len(got.Categories) != 0 {
		t.Errorf("expected empty summary, got %+v", got)
	}
}

func TestQueryExpenses_PassesResolvedPeriodToReader(t *testing.T) {
	reader := &fakeReader{}

	_, err := newQueryUseCase(reader).Execute(context.Background(), input.QueryExpenses{
		UserID: 42,
		Period: domain.PeriodSpec{Kind: domain.PeriodYesterday},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reader.gotUser != 42 {
		t.Errorf("expected user 42, got %d", reader.gotUser)
	}
	wantStart := time.Date(2026, 10, 12, 3, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 10, 13, 3, 0, 0, 0, time.UTC)
	if !reader.gotPeriod.Start.Equal(wantStart) || !reader.gotPeriod.End.Equal(wantEnd) {
		t.Errorf("unexpected period: %+v", reader.gotPeriod)
	}
}

func TestQueryExpenses_InvalidPeriodDoesNotQuery(t *testing.T) {
	reader := &fakeReader{}

	_, err := newQueryUseCase(reader).Execute(context.Background(), input.QueryExpenses{
		UserID: 1,
		Period: domain.PeriodSpec{},
	})

	if !errors.Is(err, exception.ErrInvalidPeriod) {
		t.Errorf("expected ErrInvalidPeriod, got %v", err)
	}
	if reader.called {
		t.Error("reader should not be called with an invalid period")
	}
}

func TestQueryExpenses_ReaderErrorIsWrapped(t *testing.T) {
	boom := errors.New("boom")
	reader := &fakeReader{err: boom}

	_, err := newQueryUseCase(reader).Execute(context.Background(), input.QueryExpenses{
		UserID: 1,
		Period: domain.PeriodSpec{Kind: domain.PeriodToday},
	})

	if !errors.Is(err, boom) {
		t.Errorf("expected wrapped boom, got %v", err)
	}
}
