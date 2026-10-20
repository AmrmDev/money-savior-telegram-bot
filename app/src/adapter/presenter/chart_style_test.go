package presenter

import (
	"image"
	"image/color"
	"testing"
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

func TestMixColor(t *testing.T) {
	black := color.RGBA{A: 0xFF}
	white := color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	if got := mixColor(black, white, 0); got != black {
		t.Errorf("amount 0 should keep the base color, got %v", got)
	}
	if got := mixColor(black, white, 1); got != white {
		t.Errorf("amount 1 should return the overlay color, got %v", got)
	}
	if got := mixColor(black, white, 0.5); got.R != 128 || got.G != 128 || got.B != 128 {
		t.Errorf("amount 0.5 should land in the middle, got %v", got)
	}
}

func TestDrawGrid_PaintsOnlyTheLines(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	fillRect(img, img.Bounds(), colorBackground)

	drawGrid(img)

	if got := img.RGBAAt(chartGridStep, 10); got == colorBackground {
		t.Error("a vertical grid line should be painted")
	}
	if got := img.RGBAAt(10, chartGridStep); got == colorBackground {
		t.Error("a horizontal grid line should be painted")
	}
	if got := img.RGBAAt(10, 10); got != colorBackground {
		t.Errorf("pixels between the lines must stay untouched, got %v", got)
	}
}