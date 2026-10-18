package presenter

import (
	"bytes"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"

	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/usecase"
)

func testPeriod() domain.Period {
	return domain.Period{
		Start: time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 10, 14, 3, 0, 0, 0, time.UTC),
	}
}

func TestRenderSummaryChart_ProducesValidPNG(t *testing.T) {
	rows := []usecase.CategoryTotal{
		{Category: "Mercado", TotalCents: 7500, Count: 2},
		{Category: "Uber", TotalCents: 1500, Count: 1},
		{Category: othersLabel, TotalCents: 500, Count: 1},
	}

	data, err := renderSummaryChart("Gastos de 01/10 a 13/10", "Total R$ 95,00 · 4 gastos", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("output is not a valid PNG: %v", err)
	}

	wantHeight := chartHeaderH + len(rows)*chartRowH + chartFooterH
	if img.Bounds().Dx() != chartWidth || img.Bounds().Dy() != wantHeight {
		t.Errorf("got %dx%d, want %dx%d", img.Bounds().Dx(), img.Bounds().Dy(), chartWidth, wantHeight)
	}
}

func TestRenderSummaryChart_LongCategoryName(t *testing.T) {
	rows := []usecase.CategoryTotal{
		{Category: strings.Repeat("Supermercado ", 10), TotalCents: 1000, Count: 1},
		{Category: "Uber", TotalCents: 500, Count: 1},
	}

	if _, err := renderSummaryChart("Gastos", "Total", rows); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderSummaryChart_Preview(t *testing.T) {
	out := os.Getenv("CHART_OUT")
	if out == "" {
		t.Skip("set CHART_OUT=/path/chart.png to write a preview image")
	}

	rows := []usecase.CategoryTotal{
		{Category: "Mercado", TotalCents: 52090, Count: 6},
		{Category: "Uber", TotalCents: 18700, Count: 5},
		{Category: "Restaurante", TotalCents: 15400, Count: 3},
		{Category: "Farmácia", TotalCents: 8990, Count: 2},
		{Category: "Assinaturas e serviços de streaming", TotalCents: 4590, Count: 1},
		{Category: othersLabel, TotalCents: 1200, Count: 2},
	}

	data, err := renderSummaryChart("Gastos de 01/10 a 13/10", "Total R$ 969,70 · 19 gastos", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		t.Fatalf("write preview: %v", err)
	}
}

func TestGroupCategories_MergesTail(t *testing.T) {
	var categories []usecase.CategoryTotal
	for i := 0; i < 7; i++ {
		categories = append(categories, usecase.CategoryTotal{
			Category:   string(rune('A' + i)),
			TotalCents: int64(1000 - i*100),
			Count:      1,
		})
	}

	got := groupCategories(categories)

	if len(got) != maxCategories+1 {
		t.Fatalf("expected %d rows, got %d", maxCategories+1, len(got))
	}
	others := got[len(got)-1]
	if others.Category != othersLabel || others.TotalCents != 900 || others.Count != 2 {
		t.Errorf("unexpected others row: %+v", others)
	}
}

func TestQuerySummary_ImageOnlyWithMultipleCategories(t *testing.T) {
	tests := []struct {
		name      string
		summary   usecase.ExpensesSummary
		wantImage bool
	}{
		{
			name:    "sem gastos",
			summary: usecase.ExpensesSummary{Period: testPeriod()},
		},
		{
			name: "uma categoria",
			summary: usecase.ExpensesSummary{
				Period: testPeriod(), TotalCents: 5000, Count: 2,
				Categories: []usecase.CategoryTotal{{Category: "Mercado", TotalCents: 5000, Count: 2}},
			},
		},
		{
			name: "duas categorias",
			summary: usecase.ExpensesSummary{
				Period: testPeriod(), TotalCents: 6500, Count: 3,
				Categories: []usecase.CategoryTotal{
					{Category: "Mercado", TotalCents: 5000, Count: 2},
					{Category: "Uber", TotalCents: 1500, Count: 1},
				},
			},
			wantImage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QuerySummary(tt.summary)

			if (got.Image != nil) != tt.wantImage {
				t.Errorf("image present = %v, want %v", got.Image != nil, tt.wantImage)
			}
			if got.Text == "" {
				t.Error("expected a text (message or caption)")
			}
		})
	}
}
