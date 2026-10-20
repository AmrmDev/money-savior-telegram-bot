package presenter

import (
	"strings"
	"testing"

	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

// Every character the chart can draw: money, percentages, dates, accents, the ellipsis
// used when a name is cut, and the letters of category names.
const chartCharset = "R$ 0123456789,.%<>:/·…-ÁÀÂÃÉÊÍÓÔÕÚÇáàâãéêíóôõúç" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func TestEmbeddedFonts_ParseAndCoverTheChartCharset(t *testing.T) {
	fonts := []struct {
		name string
		data []byte
		want string
	}{
		{"Space Grotesk Regular", fontSansRegular, "Space Grotesk"},
		{"Space Grotesk Medium", fontSansMedium, "Space Grotesk"},
		{"JetBrains Mono Regular", fontMono, "JetBrains Mono"},
	}

	for _, f := range fonts {
		t.Run(f.name, func(t *testing.T) {
			parsed, err := opentype.Parse(f.data)
			if err != nil {
				t.Fatalf("font does not parse: %v", err)
			}

			var buf sfnt.Buffer

			fullName, err := parsed.Name(&buf, sfnt.NameIDFull)
			if err != nil {
				t.Fatalf("cannot read the font name: %v", err)
			}
			if !strings.Contains(fullName, f.want) {
				t.Errorf("unexpected font %q, expected a %s file", fullName, f.want)
			}

			for _, r := range chartCharset {
				index, err := parsed.GlyphIndex(&buf, r)
				if err != nil {
					t.Fatalf("glyph lookup failed for %q: %v", r, err)
				}
				if index == 0 {
					t.Errorf("%s has no glyph for %q", fullName, r)
				}
			}
		})
	}
}

func TestEmbeddedFonts_MediumIsTheMediumWeight(t *testing.T) {
	parsed, err := opentype.Parse(fontSansMedium)
	if err != nil {
		t.Fatalf("font does not parse: %v", err)
	}

	var buf sfnt.Buffer
	fullName, err := parsed.Name(&buf, sfnt.NameIDFull)
	if err != nil {
		t.Fatalf("cannot read the font name: %v", err)
	}

	if !strings.Contains(fullName, "Medium") {
		t.Errorf("got %q: this looks like the wrong file (a variable font, perhaps). Use the static Medium file", fullName)
	}
}
