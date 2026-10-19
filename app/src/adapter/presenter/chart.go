package presenter

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"money-savior-telegram-bot/app/src/core/usecase"
)

const (
	chartWidth     = 1000
	chartMargin    = 40
	chartHeaderH   = 140
	chartRowH      = 64
	chartFooterH   = 30
	chartBarHeight = 30
	chartLabelW    = 220
	chartValueW    = 170
	chartPercentW  = 90
	chartColumnGap = 16
	chartCornerAA  = 4
)

type chartTheme struct {
	background color.RGBA
	title      color.RGBA
	muted      color.RGBA
	track      color.RGBA
	bar        color.RGBA
	others     color.RGBA
}

var lightTheme = chartTheme{
	background: color.RGBA{R: 0xF9, G: 0xFA, B: 0xFB, A: 0xFF},
	title:      color.RGBA{R: 0x11, G: 0x18, B: 0x27, A: 0xFF},
	muted:      color.RGBA{R: 0x6B, G: 0x72, B: 0x80, A: 0xFF},
	track:      color.RGBA{R: 0xE5, G: 0xE7, B: 0xEB, A: 0xFF},
	bar:        color.RGBA{R: 0x10, G: 0xB9, B: 0x81, A: 0xFF},
	others:     color.RGBA{R: 0x9C, G: 0xA3, B: 0xAF, A: 0xFF},
}

var darkTheme = chartTheme{
	background: color.RGBA{R: 0x1E, G: 0x29, B: 0x3B, A: 0xFF},
	title:      color.RGBA{R: 0xF1, G: 0xF5, B: 0xF9, A: 0xFF},
	muted:      color.RGBA{R: 0x94, G: 0xA3, B: 0xB8, A: 0xFF},
	track:      color.RGBA{R: 0x33, G: 0x41, B: 0x55, A: 0xFF},
	bar:        color.RGBA{R: 0x34, G: 0xD3, B: 0x99, A: 0xFF},
	others:     color.RGBA{R: 0x64, G: 0x74, B: 0x8B, A: 0xFF},
}

var defaultTheme = darkTheme

func renderSummaryChart(title, subtitle string, rows []usecase.CategoryTotal) ([]byte, error) {
	return renderSummaryChartWithTheme(defaultTheme, title, subtitle, rows)
}

func renderSummaryChartWithTheme(theme chartTheme, title, subtitle string, rows []usecase.CategoryTotal) ([]byte, error) {
	titleFace, err := newFace(gobold.TTF, 38)
	if err != nil {
		return nil, err
	}
	defer titleFace.Close()

	textFace, err := newFace(goregular.TTF, 26)
	if err != nil {
		return nil, err
	}
	defer textFace.Close()

	height := chartHeaderH + len(rows)*chartRowH + chartFooterH
	img := image.NewRGBA(image.Rect(0, 0, chartWidth, height))
	fillRect(img, img.Bounds(), theme.background)

	contentW := chartWidth - 2*chartMargin
	drawText(img, titleFace, theme.title, chartMargin, 70, fitText(titleFace, title, contentW))
	drawText(img, textFace, theme.muted, chartMargin, 112, fitText(textFace, subtitle, contentW))

	var totalCents, maxCents int64
	for _, row := range rows {
		totalCents += row.TotalCents
		if row.TotalCents > maxCents {
			maxCents = row.TotalCents
		}
	}

	metrics := textFace.Metrics()
	centerOffset := ((metrics.Ascent - metrics.Descent) / 2).Round()

	percentRight := chartWidth - chartMargin
	valueRight := percentRight - chartPercentW
	barX := chartMargin + chartLabelW
	barW := valueRight - chartValueW - chartColumnGap - barX
	radius := chartBarHeight / 2

	for i, row := range rows {
		rowTop := chartHeaderH + i*chartRowH
		barTop := rowTop + (chartRowH-chartBarHeight)/2
		baseline := barTop + chartBarHeight/2 + centerOffset

		fillRoundedRect(img, image.Rect(barX, barTop, barX+barW, barTop+chartBarHeight), radius, theme.track)

		fillW := 0
		if maxCents > 0 {
			fillW = int(int64(barW) * row.TotalCents / maxCents)
		}
		if row.TotalCents > 0 && fillW < 4 {
			fillW = 4
		}

		barColor := theme.bar
		if row.Category == othersLabel {
			barColor = theme.others
		}
		fillRoundedRect(img, image.Rect(barX, barTop, barX+fillW, barTop+chartBarHeight), radius, barColor)

		drawText(img, textFace, theme.title, chartMargin, baseline, fitText(textFace, row.Category, chartLabelW-chartColumnGap))
		drawTextRight(img, textFace, theme.title, valueRight, baseline, formatMoney(row.TotalCents))
		drawTextRight(img, textFace, theme.muted, percentRight, baseline, percentLabel(row.TotalCents, totalCents))
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode chart png: %w", err)
	}
	return buf.Bytes(), nil
}

func percentLabel(cents, total int64) string {
	if total <= 0 {
		return ""
	}

	percent := int(math.Round(float64(cents) * 100 / float64(total)))
	if percent == 0 && cents > 0 {
		return "<1%"
	}
	return fmt.Sprintf("%d%%", percent)
}

func newFace(ttf []byte, size float64) (font.Face, error) {
	parsed, err := opentype.Parse(ttf)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}

	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("create font face: %w", err)
	}
	return face, nil
}

func fillRect(img *image.RGBA, rect image.Rectangle, c color.Color) {
	draw.Draw(img, rect, image.NewUniform(c), image.Point{}, draw.Src)
}

// fillRoundedRect paints rect with rounded corners. The radius is clamped to half of the
// shortest side, so very short bars shrink their curve instead of breaking.
func fillRoundedRect(img *image.RGBA, rect image.Rectangle, radius int, c color.RGBA) {
	if rect.Empty() {
		return
	}
	if limit := min(rect.Dx(), rect.Dy()) / 2; radius > limit {
		radius = limit
	}

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if coverage := cornerCoverage(rect, radius, x, y); coverage > 0 {
				blendPixel(img, x, y, c, coverage)
			}
		}
	}
}

// cornerCoverage returns how much of the pixel is inside the rounded rectangle (0 to 1).
// Only pixels in the four corner zones need sampling, the rest is fully covered.
func cornerCoverage(rect image.Rectangle, radius, x, y int) float64 {
	inLeft := x < rect.Min.X+radius
	inRight := x >= rect.Max.X-radius
	inTop := y < rect.Min.Y+radius
	inBottom := y >= rect.Max.Y-radius

	if !(inLeft || inRight) || !(inTop || inBottom) {
		return 1
	}

	centerX := float64(rect.Min.X + radius)
	if inRight {
		centerX = float64(rect.Max.X - radius)
	}
	centerY := float64(rect.Min.Y + radius)
	if inBottom {
		centerY = float64(rect.Max.Y - radius)
	}

	inside := 0
	for sy := 0; sy < chartCornerAA; sy++ {
		for sx := 0; sx < chartCornerAA; sx++ {
			px := float64(x) + (float64(sx)+0.5)/chartCornerAA
			py := float64(y) + (float64(sy)+0.5)/chartCornerAA
			dx, dy := px-centerX, py-centerY
			if dx*dx+dy*dy <= float64(radius*radius) {
				inside++
			}
		}
	}
	return float64(inside) / (chartCornerAA * chartCornerAA)
}

func blendPixel(img *image.RGBA, x, y int, c color.RGBA, coverage float64) {
	dst := img.RGBAAt(x, y)
	mix := func(src, dst uint8) uint8 {
		return uint8(float64(src)*coverage + float64(dst)*(1-coverage) + 0.5)
	}
	img.SetRGBA(x, y, color.RGBA{R: mix(c.R, dst.R), G: mix(c.G, dst.G), B: mix(c.B, dst.B), A: 0xFF})
}

func drawText(img *image.RGBA, face font.Face, c color.Color, x, baseline int, text string) {
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(x, baseline),
	}
	drawer.DrawString(text)
}

func drawTextRight(img *image.RGBA, face font.Face, c color.Color, rightX, baseline int, text string) {
	width := font.MeasureString(face, text).Ceil()
	drawText(img, face, c, rightX-width, baseline, text)
}

func fitText(face font.Face, text string, maxWidth int) string {
	if font.MeasureString(face, text).Ceil() <= maxWidth {
		return text
	}

	runes := []rune(text)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := string(runes) + "…"
		if font.MeasureString(face, candidate).Ceil() <= maxWidth {
			return candidate
		}
	}
	return "…"
}