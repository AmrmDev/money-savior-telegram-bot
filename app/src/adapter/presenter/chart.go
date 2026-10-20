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
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"money-savior-telegram-bot/app/src/core/usecase"
)

const (
	chartWidth    = 1000
	chartMargin   = 56
	chartGridStep = 40
	chartCornerAA = 4

	chartEyebrowY  = 76
	chartTotalY    = 164
	chartSubtitleY = 210
	chartDividerY  = 242
	chartRowsTop   = 254
	chartRowH      = 80
	chartBarHeight = 14

	chartLabelX    = chartMargin + 28
	chartBarX      = 310
	chartValueW    = 190
	chartPercentW  = 84
	chartColumnGap = 16

	chartCardGap = 28
	chartCardH   = 120

	chartFooterGap = 68
	chartBottomPad = 48
)

var (
	colorBackground = color.RGBA{R: 0x0B, G: 0x0B, B: 0x24, A: 0xFF}
	colorPanel      = color.RGBA{R: 0x12, G: 0x12, B: 0x2E, A: 0xFF}
	colorBorder     = color.RGBA{R: 0x2A, G: 0x2A, B: 0x4E, A: 0xFF}
	colorDivider    = color.RGBA{R: 0x1F, G: 0x1F, B: 0x3F, A: 0xFF}
	colorTrack      = color.RGBA{R: 0x1B, G: 0x1B, B: 0x3A, A: 0xFF}
	colorText       = color.RGBA{R: 0xF5, G: 0xF5, B: 0xFF, A: 0xFF}
	colorSoftText   = color.RGBA{R: 0xC9, G: 0xC9, B: 0xE0, A: 0xFF}
	colorMuted      = color.RGBA{R: 0x8F, G: 0x8F, B: 0xB3, A: 0xFF}
	colorAccent     = color.RGBA{R: 0xFF, G: 0x8A, B: 0x3D, A: 0xFF}
	colorBar        = color.RGBA{R: 0x3D, G: 0xDC, B: 0x84, A: 0xFF}
	colorOthers     = color.RGBA{R: 0x64, G: 0x6B, B: 0x8F, A: 0xFF}
)

// chartContent holds the ready-to-draw strings, so this file never formats domain data.
type chartContent struct {
	Eyebrow  string
	Total    string
	Subtitle string
	Rows     []usecase.CategoryTotal
	Peak     *chartPeak
}

type chartPeak struct {
	Label string
	Value string
	Note  string
}

type chartFaces struct {
	total     font.Face
	subtitle  font.Face
	label     font.Face
	value     font.Face
	percent   font.Face
	eyebrow   font.Face
	peakLabel font.Face
	peakValue font.Face
	peakNote  font.Face
	footer    font.Face
}

func newChartFaces() (*chartFaces, error) {
	var faces chartFaces

	specs := []struct {
		target *font.Face
		ttf    []byte
		size   float64
	}{
		{&faces.total, fontSansMedium, 76},
		{&faces.subtitle, fontSansRegular, 26},
		{&faces.label, fontSansRegular, 28},
		{&faces.value, fontMono, 26},
		{&faces.percent, fontMono, 22},
		{&faces.eyebrow, fontMono, 20},
		{&faces.peakLabel, fontSansMedium, 34},
		{&faces.peakValue, fontMono, 36},
		{&faces.peakNote, fontMono, 20},
		{&faces.footer, fontMono, 18},
	}

	for _, spec := range specs {
		face, err := newFace(spec.ttf, spec.size)
		if err != nil {
			faces.Close()
			return nil, err
		}
		*spec.target = face
	}

	return &faces, nil
}

func (f *chartFaces) Close() {
	all := []font.Face{
		f.total, f.subtitle, f.label, f.value, f.percent,
		f.eyebrow, f.peakLabel, f.peakValue, f.peakNote, f.footer,
	}
	for _, face := range all {
		if face != nil {
			face.Close()
		}
	}
}

func chartHeight(rowCount int, withPeak bool) int {
	contentEnd := chartRowsTop + rowCount*chartRowH
	if withPeak {
		contentEnd += chartCardGap + chartCardH
	}
	return contentEnd + chartFooterGap + chartBottomPad
}

func renderSummaryChart(content chartContent) ([]byte, error) {
	faces, err := newChartFaces()
	if err != nil {
		return nil, err
	}
	defer faces.Close()

	height := chartHeight(len(content.Rows), content.Peak != nil)
	rowsEnd := chartRowsTop + len(content.Rows)*chartRowH

	img := image.NewRGBA(image.Rect(0, 0, chartWidth, height))
	fillRect(img, img.Bounds(), colorBackground)
	drawGrid(img)

	drawHeader(img, faces, content)
	drawRows(img, faces, content.Rows)
	if content.Peak != nil {
		drawPeakCard(img, faces, *content.Peak, rowsEnd+chartCardGap)
	}
	drawSpacedText(img, faces.footer, colorMuted, chartMargin, height-chartBottomPad, "MONEY SAVIOR", 4)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode chart png: %w", err)
	}
	return buf.Bytes(), nil
}

func drawGrid(img *image.RGBA) {
	grid := mixColor(colorBackground, color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, 0.04)
	bounds := img.Bounds()

	for x := 0; x < bounds.Max.X; x += chartGridStep {
		fillRect(img, image.Rect(x, 0, x+1, bounds.Max.Y), grid)
	}
	for y := 0; y < bounds.Max.Y; y += chartGridStep {
		fillRect(img, image.Rect(0, y, bounds.Max.X, y+1), grid)
	}
}

func drawHeader(img *image.RGBA, faces *chartFaces, content chartContent) {
	contentW := chartWidth - 2*chartMargin

	fillRect(img, image.Rect(chartMargin, chartEyebrowY-14, chartMargin+12, chartEyebrowY-2), colorAccent)
	drawSpacedText(img, faces.eyebrow, colorAccent, chartMargin+28, chartEyebrowY, content.Eyebrow, 2)

	drawText(img, faces.total, colorText, chartMargin, chartTotalY, fitText(faces.total, content.Total, contentW))
	drawText(img, faces.subtitle, colorMuted, chartMargin, chartSubtitleY, fitText(faces.subtitle, content.Subtitle, contentW))

	fillRect(img, image.Rect(chartMargin, chartDividerY, chartWidth-chartMargin, chartDividerY+1), colorBorder)
}

func drawRows(img *image.RGBA, faces *chartFaces, rows []usecase.CategoryTotal) {
	var totalCents, maxCents int64
	for _, row := range rows {
		totalCents += row.TotalCents
		if row.TotalCents > maxCents {
			maxCents = row.TotalCents
		}
	}

	percentRight := chartWidth - chartMargin
	valueRight := percentRight - chartPercentW
	barW := valueRight - chartValueW - chartColumnGap - chartBarX
	radius := chartBarHeight / 2

	for i, row := range rows {
		rowTop := chartRowsTop + i*chartRowH
		center := rowTop + chartRowH/2
		barTop := center - chartBarHeight/2
		isLeader := i == 0

		nameColor, valueColor := colorSoftText, colorSoftText
		if isLeader {
			nameColor, valueColor = colorText, colorText
			fillRect(img, image.Rect(chartMargin, center-26, chartMargin+4, center+26), colorAccent)
		}

		fillRoundedRect(img, image.Rect(chartBarX, barTop, chartBarX+barW, barTop+chartBarHeight), radius, colorTrack)

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
		fillRoundedRect(img, image.Rect(chartBarX, barTop, chartBarX+fillW, barTop+chartBarHeight), radius, barColor)

		nameWidth := chartBarX - chartLabelX - chartColumnGap
		drawText(img, faces.label, nameColor, chartLabelX, centeredBaseline(faces.label, center), fitText(faces.label, row.Category, nameWidth))
		drawTextRight(img, faces.value, valueColor, valueRight, centeredBaseline(faces.value, center), formatMoney(row.TotalCents))
		drawTextRight(img, faces.percent, colorMuted, percentRight, centeredBaseline(faces.percent, center), percentLabel(row.TotalCents, totalCents))

		separator := colorDivider
		if i == len(rows)-1 {
			separator = colorBorder
		}
		fillRect(img, image.Rect(chartMargin, rowTop+chartRowH-1, chartWidth-chartMargin, rowTop+chartRowH), separator)
	}
}

func drawPeakCard(img *image.RGBA, faces *chartFaces, peak chartPeak, top int) {
	left, right := chartMargin, chartWidth-chartMargin

	fillRect(img, image.Rect(left, top, right, top+chartCardH), colorBorder)
	fillRect(img, image.Rect(left+1, top+1, right-1, top+chartCardH-1), colorPanel)
	fillRect(img, image.Rect(left, top, left+4, top+chartCardH), colorAccent)

	textX := left + 32
	valueRight := right - 32

	drawSpacedText(img, faces.eyebrow, colorAccent, textX, top+42, "DIA MAIS PESADO", 2)
	drawText(img, faces.peakLabel, colorText, textX, top+92, fitText(faces.peakLabel, peak.Label, valueRight-textX-300))
	drawTextRight(img, faces.peakNote, colorMuted, valueRight, top+42, peak.Note)
	drawTextRight(img, faces.peakValue, colorAccent, valueRight, top+92, peak.Value)
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

func mixColor(base, over color.RGBA, amount float64) color.RGBA {
	mix := func(b, o uint8) uint8 {
		return uint8(float64(b)*(1-amount) + float64(o)*amount + 0.5)
	}
	return color.RGBA{R: mix(base.R, over.R), G: mix(base.G, over.G), B: mix(base.B, over.B), A: 0xFF}
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

func centeredBaseline(face font.Face, center int) int {
	metrics := face.Metrics()
	return center + ((metrics.Ascent - metrics.Descent) / 2).Round()
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

func drawSpacedText(img *image.RGBA, face font.Face, c color.Color, x, baseline int, text string, spacing int) {
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(x, baseline),
	}
	for _, r := range text {
		drawer.DrawString(string(r))
		drawer.Dot.X += fixed.I(spacing)
	}
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