package presenter

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"money-savior-telegram-bot/app/src/core/usecase"
)

func TestPercentLabel(t *testing.T) {
	tests := []struct {
		name  string
		cents int64
		total int64
		want  string
	}{
		{"maior parte", 6340, 12030, "53%"},
		{"arredonda para baixo", 2190, 12030, "18%"},
		{"tudo", 12030, 12030, "100%"},
		{"menos de 1%", 40, 12030, "<1%"},
		{"zero", 0, 12030, "0%"},
		{"total zerado", 1000, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := percentLabel(tt.cents, tt.total); got != tt.want {
				t.Errorf("percentLabel(%d, %d) = %q, want %q", tt.cents, tt.total, got, tt.want)
			}
		})
	}
}

func TestFillRoundedRect_RoundsTheCorners(t *testing.T) {
	background := color.RGBA{A: 0xFF}
	foreground := color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	fillRect(img, img.Bounds(), background)

	fillRoundedRect(img, image.Rect(0, 0, 40, 20), 10, foreground)

	if got := img.RGBAAt(0, 0); got != background {
		t.Errorf("corner pixel should stay untouched, got %v", got)
	}
	if got := img.RGBAAt(20, 10); got != foreground {
		t.Errorf("center pixel should be filled, got %v", got)
	}
	if got := img.RGBAAt(20, 0); got != foreground {
		t.Errorf("middle of the top edge should be filled, got %v", got)
	}

	smooth := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			if v := img.RGBAAt(x, y).R; v > 0 && v < 0xFF {
				smooth = true
			}
		}
	}
	if !smooth {
		t.Error("expected anti-aliased pixels along the curve")
	}
}

func TestFillRoundedRect_ShortBarKeepsItsBody(t *testing.T) {
	background := color.RGBA{A: 0xFF}
	foreground := color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	img := image.NewRGBA(image.Rect(0, 0, 10, 30))
	fillRect(img, img.Bounds(), background)

	fillRoundedRect(img, image.Rect(0, 0, 4, 30), 15, foreground)

	if got := img.RGBAAt(2, 15); got == background {
		t.Error("a very short bar should still be visible")
	}
	if got := img.RGBAAt(8, 15); got != background {
		t.Errorf("pixels outside the bar must stay untouched, got %v", got)
	}
}

func TestFillRoundedRect_EmptyRectIsIgnored(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	fillRoundedRect(img, image.Rect(5, 5, 5, 5), 4, color.RGBA{R: 0xFF, A: 0xFF})
}

func TestRenderSummaryChart_UsesThemeBackground(t *testing.T) {
	rows := []usecase.CategoryTotal{
		{Category: "Mercado", TotalCents: 7500, Count: 2},
		{Category: "Uber", TotalCents: 1500, Count: 1},
	}

	themes := map[string]chartTheme{"dark": darkTheme, "light": lightTheme}

	for name, theme := range themes {
		t.Run(name, func(t *testing.T) {
			data, err := renderSummaryChartWithTheme(theme, "Gastos", "Total", rows)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("output is not a valid PNG: %v", err)
			}

			got := color.RGBAModel.Convert(img.At(0, 0)).(color.RGBA)
			if got != theme.background {
				t.Errorf("background = %v, want %v", got, theme.background)
			}
		})
	}
}

func TestRenderSummaryChart_ThemePreview(t *testing.T) {
	dir := os.Getenv("CHART_DIR")
	if dir == "" {
		t.Skip("set CHART_DIR=/some/folder to write chart-dark.png and chart-light.png")
	}

	rows := []usecase.CategoryTotal{
		{Category: "Mercado", TotalCents: 52090, Count: 6},
		{Category: "Uber", TotalCents: 18700, Count: 5},
		{Category: "Restaurante", TotalCents: 15400, Count: 3},
		{Category: "Farmácia", TotalCents: 8990, Count: 2},
		{Category: "Assinaturas e serviços de streaming", TotalCents: 4590, Count: 1},
		{Category: othersLabel, TotalCents: 1200, Count: 2},
	}

	themes := map[string]chartTheme{"dark": darkTheme, "light": lightTheme}

	for name, theme := range themes {
		data, err := renderSummaryChartWithTheme(theme, "Gastos de 01/10 a 13/10", "Total R$ 969,70 · 19 gastos", rows)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "chart-"+name+".png"), data, 0o644); err != nil {
			t.Fatalf("write preview: %v", err)
		}
	}
}
