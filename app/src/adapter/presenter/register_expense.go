package presenter

import (
	"errors"
	"fmt"

	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/domain/exception"
)

const receiptDivider = "━━━━━━━━━━━━━━"

func InvalidExpenseFormat() Reply {
	return Reply{Text: "⚠️ Formato incorreto — use:\n/gastei &lt;valor&gt; &lt;categoria&gt; [método]\n\nExemplo:\n/gastei 21,90 uber pix"}
}

func InvalidAmount() Reply {
	return Reply{Text: "❌ Valor inválido."}
}

func ExpenseRegistered(e domain.Expense) Reply {
	return Reply{Text: fmt.Sprintf(
		"Opa, gasto registrado! ✅\n\n📝 %s\n💰 %s",
		e.Category, formatMoney(e.AmountCents),
	)}
}

func RegisterExpenseError(err error) Reply {
	switch {
	case errors.Is(err, exception.ErrInvalidAmount):
		return InvalidAmount()
	case errors.Is(err, exception.ErrInvalidCategory):
		return Reply{Text: "❌ Categoria inválida."}
	default:
		return Reply{Text: "❌ Erro ao salvar gasto."}
	}
}

func formatMoney(cents int64) string {
	return fmt.Sprintf("R$ %d,%02d", cents/100, cents%100)
}
