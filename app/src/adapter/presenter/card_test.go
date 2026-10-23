package presenter

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

func TestRenderCard_ProducesValidPNG(t *testing.T) {
	cards := map[string]cardContent{"invite": inviteCard, "guide": guideCard}

	for name, content := range cards {
		t.Run(name, func(t *testing.T) {
			data, err := renderCard(content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("output is not a valid PNG: %v", err)
			}

			layout := layoutCard(content)
			if img.Bounds().Dx() != chartWidth || img.Bounds().Dy() != layout.height {
				t.Errorf("got %dx%d, want %dx%d", img.Bounds().Dx(), img.Bounds().Dy(), chartWidth, layout.height)
			}
			if got := pixelAt(img, 10, 10); got != colorBackground {
				t.Errorf("background = %v, want %v", got, colorBackground)
			}
			if got := pixelAt(img, chartMargin+1, chartEyebrowY-8); got != colorAccent {
				t.Errorf("expected the accent marker next to the eyebrow, got %v", got)
			}
		})
	}
}

func TestLayoutCard_GrowsWithItsContent(t *testing.T) {
	base := cardContent{Title: "T", Steps: []cardStep{{Title: "a", Detail: "b"}}}

	withSubtitle := base
	withSubtitle.Subtitle = "subtitle"

	withHint := base
	withHint.Hint = "hint"

	moreSteps := base
	moreSteps.Steps = append([]cardStep{}, base.Steps...)
	moreSteps.Steps = append(moreSteps.Steps, cardStep{Title: "c", Detail: "d"})

	baseHeight := layoutCard(base).height

	for name, content := range map[string]cardContent{
		"subtitle":   withSubtitle,
		"hint":       withHint,
		"extra step": moreSteps,
	} {
		if got := layoutCard(content).height; got <= baseHeight {
			t.Errorf("%s: height %d should be greater than %d", name, got, baseHeight)
		}
	}
}

func TestRenderCard_LongTextDoesNotFail(t *testing.T) {
	long := strings.Repeat("palavra comprida ", 20)

	content := cardContent{
		Eyebrow:    "EYEBROW",
		Title:      long,
		Subtitle:   long,
		Steps:      []cardStep{{Title: long, Detail: long}, {Title: long, Detail: long, Mono: true}},
		Hint:       strings.ToUpper(long),
		Note:       long,
		FooterNote: strings.ToUpper(long),
	}

	if _, err := renderCard(content); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCachedCard_RendersOnce(t *testing.T) {
	card := &cachedCard{content: guideCard}

	first := card.png()
	second := card.png()

	if len(first) == 0 {
		t.Fatal("expected a rendered image")
	}
	if &first[0] != &second[0] {
		t.Error("the second call should return the cached image")
	}
}

func assertGlyphs(t *testing.T, ttf []byte, text string) {
	t.Helper()

	parsed, err := opentype.Parse(ttf)
	if err != nil {
		t.Fatalf("font does not parse: %v", err)
	}

	var buf sfnt.Buffer
	for _, r := range text {
		index, err := parsed.GlyphIndex(&buf, r)
		if err != nil {
			t.Fatalf("glyph lookup failed for %q: %v", r, err)
		}
		if index == 0 {
			t.Errorf("the font has no glyph for %q (U+%04X)", r, r)
		}
	}
}

func TestCardTexts_AreCoveredByTheFonts(t *testing.T) {
	cards := map[string]cardContent{"invite": inviteCard, "guide": guideCard}

	for name, content := range cards {
		t.Run(name, func(t *testing.T) {
			medium := content.Title
			regular := content.Subtitle + content.Note
			mono := content.Eyebrow + content.Hint + content.FooterNote + "MONEY SAVIOR" + "0123456789"

			for _, step := range content.Steps {
				medium += step.Title
				if step.Mono {
					mono += step.Detail
				} else {
					regular += step.Detail
				}
			}

			assertGlyphs(t, fontSansMedium, medium)
			assertGlyphs(t, fontSansRegular, regular)
			assertGlyphs(t, fontMono, mono)
		})
	}
}

func TestRenderCard_Preview(t *testing.T) {
	dir := os.Getenv("CHART_DIR")
	if dir == "" {
		t.Skip("set CHART_DIR=/some/folder to write card-invite.png and card-guide.png")
	}

	files := map[string]cardContent{"card-invite.png": inviteCard, "card-guide.png": guideCard}

	for name, content := range files {
		data, err := renderCard(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatalf("write preview: %v", err)
		}
	}
}