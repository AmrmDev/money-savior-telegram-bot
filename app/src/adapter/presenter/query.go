package presenter

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/domain/exception"
	"money-savior-telegram-bot/app/src/core/usecase"
)

const (
	maxCategories = 5
	othersLabel   = "Outros"
)

var weekdayNames = [7]string{"domingo", "segunda", "terça", "quarta", "quinta", "sexta", "sábado"}

func InvalidQueryFormat() Reply {
	return Reply{Text: "⚠️ Período não reconhecido. Use:\n\n/consulta hoje\n/consulta ontem\n/consulta semana\n/consulta mes\n/consulta mes passado\n/consulta 01/10 15/10"}
}

func QueryExpensesError(err error) Reply {
	switch {
	case errors.Is(err, exception.ErrInvalidPeriod):
		return Reply{Text: "❌ Período inválido. Confira as datas."}
	default:
		return Reply{Text: "❌ Erro ao consultar gastos."}
	}
}

func QuerySummary(s usecase.ExpensesSummary) Reply {
	label := periodLabel(s.Period)

	if s.Count == 0 {
		return Reply{Text: fmt.Sprintf("Nenhum gasto registrado (%s).", label)}
	}

	rows := groupCategories(s.Categories)
	peak := describePeakDay(s)

	if len(rows) > 1 {
		content := chartContent{
			Eyebrow:  "GASTOS · " + strings.ToUpper(label),
			Total:    formatMoney(s.TotalCents),
			Subtitle: countLabel(s.Count),
			Rows:     rows,
		}
		if peak != nil {
			content.Peak = peak.card()
		}

		if image, err := renderSummaryChart(content); err == nil {
			return Reply{Text: chartCaption(label, peak), Image: image}
		}
	}

	return textSummary(label, s, rows, peak)
}

type peakDay struct {
	weekday string
	date    string
	value   string
	count   string
}

// describePeakDay returns nil when there is nothing worth highlighting:
// no expenses, or a period of a single day (the peak would be that day).
func describePeakDay(s usecase.ExpensesSummary) *peakDay {
	if s.PeakDay == nil || isSingleDay(s.Period) {
		return nil
	}

	date := s.PeakDay.Date
	return &peakDay{
		weekday: weekdayNames[date.Weekday()],
		date:    fmt.Sprintf("%02d/%02d", date.Day, int(date.Month)),
		value:   formatMoney(s.PeakDay.TotalCents),
		count:   countLabel(s.PeakDay.Count),
	}
}

func (p peakDay) sentence() string {
	return fmt.Sprintf("Seu dia mais pesado foi %s (%s): %s", p.weekday, p.date, p.value)
}

func (p peakDay) card() *chartPeak {
	return &chartPeak{
		Label: capitalize(p.weekday) + " · " + p.date,
		Value: p.value,
		Note:  p.count,
	}
}

func chartCaption(label string, peak *peakDay) string {
	caption := fmt.Sprintf("📊 Gastos de %s", label)
	if peak != nil {
		caption += "\n\n" + peak.sentence()
	}
	return caption
}

func textSummary(label string, s usecase.ExpensesSummary, rows []usecase.CategoryTotal, peak *peakDay) Reply {
	var b strings.Builder
	fmt.Fprintf(&b, "📊 Gastos de %s\n\n", label)
	fmt.Fprintf(&b, "💰 Total: %s (%s)\n\n", formatMoney(s.TotalCents), countLabel(s.Count))
	for _, c := range rows {
		fmt.Fprintf(&b, "🏷️ %s — %s\n", c.Category, formatMoney(c.TotalCents))
	}
	if peak != nil {
		b.WriteString("\n" + peak.sentence())
	}

	return Reply{Text: strings.TrimRight(b.String(), "\n")}
}

func isSingleDay(p domain.Period) bool {
	first, last := p.Days()
	return first.Format("2006-01-02") == last.Format("2006-01-02")
}

func periodLabel(p domain.Period) string {
	first, last := p.Days()
	if isSingleDay(p) {
		return first.Format("02/01")
	}
	return first.Format("02/01") + " a " + last.Format("02/01")
}

func countLabel(count int) string {
	if count == 1 {
		return "1 gasto"
	}
	return fmt.Sprintf("%d gastos", count)
}

func capitalize(text string) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func groupCategories(categories []usecase.CategoryTotal) []usecase.CategoryTotal {
	if len(categories) <= maxCategories {
		return categories
	}

	grouped := append([]usecase.CategoryTotal{}, categories[:maxCategories]...)
	others := usecase.CategoryTotal{Category: othersLabel}
	for _, c := range categories[maxCategories:] {
		others.TotalCents += c.TotalCents
		others.Count += c.Count
	}

	return append(grouped, others)
}