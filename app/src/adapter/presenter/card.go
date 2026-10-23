package presenter

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"golang.org/x/image/font"
)

const (
	cardTitleY    = 156
	cardSubtitleY = 204
	cardStepRowH  = 112
	cardTextX     = chartMargin + 78
)

type cardStep struct {
	Title  string
	Detail string
	Mono   bool // the detail is drawn in the mono font (commands and examples)
}

type cardContent struct {
	Eyebrow    string
	Title      string
	Subtitle   string
	Steps      []cardStep
	Hint       string // orange line under the steps, on the left
	Note       string // muted line under the steps, on the right
	FooterNote string // mono text at the right of the footer
}

type cardLayout struct {
	dividerY       int
	stepsTop       int
	hintBaseline   int
	footerBaseline int
	height         int
}

func layoutCard(content cardContent) cardLayout {
	headerBottom := cardTitleY
	if content.Subtitle != "" {
		headerBottom = cardSubtitleY
	}

	layout := cardLayout{dividerY: headerBottom + 36}
	layout.stepsTop = layout.dividerY

	cursor := layout.stepsTop + len(content.Steps)*cardStepRowH
	if content.Hint != "" || content.Note != "" {
		layout.hintBaseline = cursor + 48
		cursor = layout.hintBaseline
	}

	layout.footerBaseline = cursor + 56
	layout.height = layout.footerBaseline + 44
	return layout
}

type cardFaces struct {
	eyebrow    font.Face
	title      font.Face
	subtitle   font.Face
	number     font.Face
	stepTitle  font.Face
	detailSans font.Face
	detailMono font.Face
	hint       font.Face
	note       font.Face
	footer     font.Face
	footerNote font.Face
}

func newCardFaces() (*cardFaces, error) {
	var faces cardFaces

	specs := []struct {
		target *font.Face
		ttf    []byte
		size   float64
	}{
		{&faces.eyebrow, fontMono, 20},
		{&faces.title, fontSansMedium, 58},
		{&faces.subtitle, fontSansRegular, 27},
		{&faces.number, fontMono, 22},
		{&faces.stepTitle, fontSansMedium, 25},
		{&faces.detailSans, fontSansRegular, 22},
		{&faces.detailMono, fontMono, 21},
		{&faces.hint, fontMono, 18},
		{&faces.note, fontSansRegular, 21},
		{&faces.footer, fontMono, 18},
		{&faces.footerNote, fontMono, 16},
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

func (f *cardFaces) Close() {
	all := []font.Face{
		f.eyebrow, f.title, f.subtitle, f.number, f.stepTitle, f.detailSans,
		f.detailMono, f.hint, f.note, f.footer, f.footerNote,
	}
	for _, face := range all {
		if face != nil {
			face.Close()
		}
	}
}

func renderCard(content cardContent) ([]byte, error) {
	faces, err := newCardFaces()
	if err != nil {
		return nil, err
	}
	defer faces.Close()

	layout := layoutCard(content)
	contentW := chartWidth - 2*chartMargin

	img := image.NewRGBA(image.Rect(0, 0, chartWidth, layout.height))
	fillRect(img, img.Bounds(), colorBackground)
	drawGrid(img)

	fillRect(img, image.Rect(chartMargin, chartEyebrowY-14, chartMargin+12, chartEyebrowY-2), colorAccent)
	drawSpacedText(img, faces.eyebrow, colorAccent, chartMargin+28, chartEyebrowY, content.Eyebrow, 2)

	drawText(img, faces.title, colorText, chartMargin, cardTitleY, fitText(faces.title, content.Title, contentW))
	if content.Subtitle != "" {
		drawText(img, faces.subtitle, colorSoftText, chartMargin, cardSubtitleY, fitText(faces.subtitle, content.Subtitle, contentW))
	}

	fillRect(img, image.Rect(chartMargin, layout.dividerY, chartWidth-chartMargin, layout.dividerY+1), colorBorder)
	drawSteps(img, faces, content.Steps, layout)
	drawHintRow(img, faces, content, layout)

	drawSpacedText(img, faces.footer, colorMuted, chartMargin, layout.footerBaseline, "MONEY SAVIOR", 4)
	if content.FooterNote != "" {
		drawSpacedTextRight(img, faces.footerNote, colorMuted, chartWidth-chartMargin, layout.footerBaseline, content.FooterNote, 1)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode card png: %w", err)
	}
	return buf.Bytes(), nil
}

func drawSteps(img *image.RGBA, faces *cardFaces, steps []cardStep, layout cardLayout) {
	textWidth := chartWidth - chartMargin - cardTextX

	for i, step := range steps {
		rowTop := layout.stepsTop + i*cardStepRowH

		drawText(img, faces.number, colorAccent, chartMargin, rowTop+66, fmt.Sprintf("%02d", i+1))
		drawText(img, faces.stepTitle, colorText, cardTextX, rowTop+50, fitText(faces.stepTitle, step.Title, textWidth))

		detailFace, detailColor := faces.detailSans, colorMuted
		if step.Mono {
			detailFace, detailColor = faces.detailMono, colorSoftText
		}
		drawText(img, detailFace, detailColor, cardTextX, rowTop+88, fitText(detailFace, step.Detail, textWidth))

		separator := colorDivider
		if i == len(steps)-1 {
			separator = colorBorder
		}
		fillRect(img, image.Rect(chartMargin, rowTop+cardStepRowH-1, chartWidth-chartMargin, rowTop+cardStepRowH), separator)
	}
}

func drawHintRow(img *image.RGBA, faces *cardFaces, content cardContent, layout cardLayout) {
	if content.Hint == "" && content.Note == "" {
		return
	}

	contentW := chartWidth - 2*chartMargin
	hintW := 0
	if content.Hint != "" {
		drawSpacedText(img, faces.hint, colorAccent, chartMargin, layout.hintBaseline, content.Hint, 2)
		hintW = measureSpaced(faces.hint, content.Hint, 2)
	}

	if content.Note != "" {
		room := contentW - hintW - 2*chartColumnGap
		drawTextRight(img, faces.note, colorMuted, chartWidth-chartMargin, layout.hintBaseline, fitText(faces.note, content.Note, room))
	}
}

func measureSpaced(face font.Face, text string, spacing int) int {
	return font.MeasureString(face, text).Ceil() + spacing*len([]rune(text))
}

func drawSpacedTextRight(img *image.RGBA, face font.Face, c color.Color, rightX, baseline int, text string, spacing int) {
	drawSpacedText(img, face, c, rightX-measureSpaced(face, text, spacing), baseline, text, spacing)
}