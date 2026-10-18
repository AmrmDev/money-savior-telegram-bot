package presenter

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"money-savior-telegram-bot/app/src/core/usecase"
)

const (
	chartWidth     = 900
	chartMargin    = 40
	chartHeaderH   = 140
	chartRowH      = 64
	chartFooterH   = 30
	chartBarHeight = 30
	chartLabelW    = 220
	chartValueW    = 210
)

var (
	colorBackground = color.RGBA{R: 0xF9, G: 0xFA, B: 0xFB, A: 0xFF}
	colorTitle      = color.RGBA{R: 0x11, G: 0x18, B: 0x27, A: 0xFF}
	colorMuted      = color.RGBA{R: 0x6B, G: 0x72, B: 0x80, A: 0xFF}
	colorTrack      = color.RGBA{R: 0xE5, G: 0xE7, B: 0xEB, A: 0xFF}
	colorBar        = color.RGBA{R: 0x10, G: 0xB9, B: 0x81, A: 0xFF}
	colorOthers     = color.RGBA{R: 0x9C, G: 0xA3, B: 0xAF, A: 0xFF}
)

func renderSummaryChart(title, subtitle string, rows []usecase.CategoryTotal) ([]byte, error) {
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
	fillRect(img, img.Bounds(), colorBackground)

	contentW := chartWidth - 2*chartMargin
	drawText(img, titleFace, colorTitle, chartMargin, 70, fitText(titleFace, title, contentW))
	drawText(img, textFace, colorMuted, chartMargin, 112, fitText(textFace, subtitle, contentW))

	var maxCents int64
	for _, row := range rows {
		if row.TotalCents > maxCents {
			maxCents = row.TotalCents
		}
	}

	metrics := textFace.Metrics()
	centerOffset := ((metrics.Ascent - metrics.Descent) / 2).Round()
	barX := chartMargin + chartLabelW
	barW := chartWidth - 2*chartMargin - chartLabelW - chartValueW

	for i, row := range rows {
		rowTop := chartHeaderH + i*chartRowH
		barTop := rowTop + (chartRowH-chartBarHeight)/2
		baseline := barTop + chartBarHeight/2 + centerOffset

		fillRect(img, image.Rect(barX, barTop, barX+barW, barTop+chartBarHeight), colorTrack)

		fillW := 0
		if maxCents > 0 {
			fillW = int(int64(barW) * row.TotalCents / maxCents)
		}
		if row.TotalCents > 0 && fillW < 4 {
			fillW = 4
		}

		barColor := colorBar
		if row.Category == othersLabel {
			barColor = colorOthers
		}
		fillRect(img, image.Rect(barX, barTop, barX+fillW, barTop+chartBarHeight), barColor)

		drawText(img, textFace, colorTitle, chartMargin, baseline, fitText(textFace, row.Category, chartLabelW-16))
		drawText(img, textFace, colorMuted, barX+barW+16, baseline, formatMoney(row.TotalCents))
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode chart png: %w", err)
	}
	return buf.Bytes(), nil
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

func drawText(img *image.RGBA, face font.Face, c color.Color, x, baseline int, text string) {
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(x, baseline),
	}
	drawer.DrawString(text)
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
