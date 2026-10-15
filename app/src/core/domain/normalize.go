package domain

import (
	"strings"
	"unicode"
)

func formatTitle(text string) string {
	words := strings.Fields(strings.ToLower(text))
	for i, word := range words {
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func normalizeMethod(method string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "debito", "débito":
		return "Débito"
	case "credito", "crédito":
		return "Crédito"
	case "pix":
		return "Pix"
	case "dinheiro":
		return "Dinheiro"
	default:
		return formatTitle(method)
	}
}
