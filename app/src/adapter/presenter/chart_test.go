package presenter

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
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

func sampleRows() []usecase.CategoryTotal {
	return []usecase.CategoryTotal{
		{Category: "Mercado", TotalCents: 52090, Count: 6},
		{Category: "Uber", TotalCents: 18700, Count: 5},
		{Category: "Restaurante", TotalCents: 15400, Count: 3},
		{Category: "Farmácia", TotalCents: 8990, Count: 2},
		{Category: "Assinaturas e serviços de streaming", TotalCents: 4590, Count: 1},
		{Category: othersLabel, TotalCents: 1200, Count: 2},
	}
}

func sampleContent(withPeak bool) chartContent {
	content := chartContent{
		Eyebrow:  "GASTOS · 01/10 A 18/10",
		Total:    "R$ 969,70",
		Subtitle: "19 gastos",
		Rows:     sampleRows(),
	}
	if withPeak {
		content.Peak = &chartPeak{Label: "Segunda · 12/10", Value: "R$ 372,80", Note: "4 gastos"}
	}
	return content
}

func renderAndDecode(t *testing.T, content chartContent) image.Image {
	t.Helper()

	data, err := renderSummaryChart(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("output is not a valid PNG: %v", err)
	}
	return img
}

func pixelAt(img image.Image, x, y int) color.RGBA {
	return color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
}

func TestRenderSummaryChart_ProducesValidPNG(t *testing.T) {
	for _, withPeak := range []bool{true, false} {
		name := "without peak"
		if withPeak {
			name = "with peak"
		}

		t.Run(name, func(t *testing.T) {
			content := sampleContent(withPeak)

			img := renderAndDecode(t, content)

			wantHeight := chartHeight(len(content.Rows), withPeak)
			if img.Bounds().Dx() != chartWidth || img.Bounds().Dy() != wantHeight {
				t.Errorf("got %dx%d, want %dx%d", img.Bounds().Dx(), img.Bounds().Dy(), chartWidth, wantHeight)
			}
			if got := pixelAt(img, 10, 10); got != colorBackground {
				t.Errorf("background = %v, want %v", got, colorBackground)
			}
		})
	}
}

func TestRenderSummaryChart_DrawsPeakCardOnlyWhenGiven(t *testing.T) {
	cardTop := chartRowsTop + len(sampleRows())*chartRowH + chartCardGap

	withPeak := renderAndDecode(t, sampleContent(true))
	if got := pixelAt(withPeak, chartMargin+1, cardTop+10); got != colorAccent {
		t.Errorf("expected the card accent bar, got %v", got)
	}

	withoutPeak := renderAndDecode(t, sampleContent(false))
	if got := pixelAt(withoutPeak, chartMargin+1, cardTop+10); got == colorAccent {
		t.Error("the card must not be drawn without a peak day")
	}
}

func TestRenderSummaryChart_MarksTheLeaderRow(t *testing.T) {
	img := renderAndDecode(t, sampleContent(false))

	leaderCenter := chartRowsTop + chartRowH/2
	secondCenter := chartRowsTop + chartRowH + chartRowH/2

	if got := pixelAt(img, chartMargin+1, leaderCenter); got != colorAccent {
		t.Errorf("expected the accent marker on the leader row, got %v", got)
	}
	if got := pixelAt(img, chartMargin+1, secondCenter); got == colorAccent {
		t.Error("only the leader row should carry the accent marker")
	}
}

func TestRenderSummaryChart_LongCategoryName(t *testing.T) {
	content := sampleContent(true)
	content.Rows = []usecase.CategoryTotal{
		{Category: strings.Repeat("Supermercado ", 10), TotalCents: 1000, Count: 1},
		{Category: "Uber", TotalCents: 500, Count: 1},
	}
	content.Peak.Label = strings.Repeat("Terça-feira ", 10)

	if _, err := renderSummaryChart(content); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderSummaryChart_Preview(t *testing.T) {
	dir := os.Getenv("CHART_DIR")
	if dir == "" {
		t.Skip("set CHART_DIR=/some/folder to write chart-peak.png and chart-plain.png")
	}

	files := map[string]bool{"chart-peak.png": true, "chart-plain.png": false}

	for name, withPeak := range files {
		data, err := renderSummaryChart(sampleContent(withPeak))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatalf("write preview: %v", err)
		}
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

func peakSummary(period domain.Period, categories ...usecase.CategoryTotal) usecase.ExpensesSummary {
	return usecase.ExpensesSummary{
		Period:     period,
		TotalCents: 6500,
		Count:      3,
		Categories: categories,
		PeakDay:    &usecase.DayTotal{Date: domain.Date{Year: 2026, Month: 10, Day: 12}, TotalCents: 8000, Count: 2},
	}
}

func TestQuerySummary_MentionsThePeakDayInTheCaption(t *testing.T) {
	summary := peakSummary(testPeriod(),
		usecase.CategoryTotal{Category: "Mercado", TotalCents: 5000, Count: 2},
		usecase.CategoryTotal{Category: "Uber", TotalCents: 1500, Count: 1},
	)

	got := QuerySummary(summary)

	if got.Image == nil {
		t.Fatal("expected an image")
	}
	want := "Seu dia mais pesado foi segunda (12/10): R$ 80,00"
	if !strings.Contains(got.Text, want) {
		t.Errorf("caption %q should contain %q", got.Text, want)
	}
}

func TestQuerySummary_MentionsThePeakDayInTheTextReply(t *testing.T) {
	summary := peakSummary(testPeriod(),
		usecase.CategoryTotal{Category: "Mercado", TotalCents: 6500, Count: 3},
	)

	got := QuerySummary(summary)

	if got.Image != nil {
		t.Fatal("a single category should not produce an image")
	}
	if !strings.Contains(got.Text, "Seu dia mais pesado foi segunda (12/10): R$ 80,00") {
		t.Errorf("text reply should mention the peak day: %q", got.Text)
	}
}

func TestQuerySummary_SingleDayPeriodHasNoPeakDay(t *testing.T) {
	oneDay := domain.Period{
		Start: time.Date(2026, 10, 13, 3, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 10, 14, 3, 0, 0, 0, time.UTC),
	}
	summary := peakSummary(oneDay,
		usecase.CategoryTotal{Category: "Mercado", TotalCents: 5000, Count: 2},
		usecase.CategoryTotal{Category: "Uber", TotalCents: 1500, Count: 1},
	)

	got := QuerySummary(summary)

	if strings.Contains(got.Text, "dia mais pesado") {
		t.Errorf("a single day period must not mention the peak day: %q", got.Text)
	}
}