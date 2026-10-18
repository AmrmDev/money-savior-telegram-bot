package presenter

import (
	"errors"
	"fmt"
	"strings"

	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/domain/exception"
	"money-savior-telegram-bot/app/src/core/usecase"
)

const (
	maxCategories = 5
	othersLabel   = "Outros"
)

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

	if len(rows) > 1 {
		title := fmt.Sprintf("Gastos de %s", label)
		subtitle := fmt.Sprintf("Total %s · %s", formatMoney(s.TotalCents), countLabel(s.Count))

		if image, err := renderSummaryChart(title, subtitle, rows); err == nil {
			return Reply{Text: fmt.Sprintf("📊 %s — %s", title, subtitle), Image: image}
		}
	}

	return textSummary(label, s, rows)
}

func textSummary(label string, s usecase.ExpensesSummary, rows []usecase.CategoryTotal) Reply {
	var b strings.Builder
	fmt.Fprintf(&b, "📊 Gastos de %s\n\n", label)
	fmt.Fprintf(&b, "💰 Total: %s (%s)\n\n", formatMoney(s.TotalCents), countLabel(s.Count))
	for _, c := range rows {
		fmt.Fprintf(&b, "🏷️ %s — %s\n", c.Category, formatMoney(c.TotalCents))
	}

	return Reply{Text: strings.TrimRight(b.String(), "\n")}
}

func periodLabel(p domain.Period) string {
	first, last := p.Days()
	if first.Format("2006-01-02") == last.Format("2006-01-02") {
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
