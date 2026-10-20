package presenter

import (
	"embed"
	"fmt"
)

// Chart fonts, bundled into the binary so the bot has no file dependencies at runtime.
// Both families are licensed under the SIL Open Font License (see the fonts folder).
//
//go:embed fonts/*.ttf
var fontFiles embed.FS

var (
	fontSansRegular = mustReadFont("SpaceGrotesk-Regular.ttf")
	fontSansMedium  = mustReadFont("SpaceGrotesk-Medium.ttf")
	fontMono        = mustReadFont("JetBrainsMono-Regular.ttf")
)

// mustReadFont fails at startup when an embedded font is missing. That is a broken build,
// and failing loudly beats answering every query without the chart.
func mustReadFont(name string) []byte {
	data, err := fontFiles.ReadFile("fonts/" + name)
	if err != nil {
		panic(fmt.Sprintf("embedded font %s is missing: %v", name, err))
	}
	return data
}
