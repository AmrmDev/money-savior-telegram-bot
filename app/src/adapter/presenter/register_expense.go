package presenter

import (
	"errors"
	"fmt"

	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/domain/exception"
)

func InvalidExpenseFormat() Reply {
	return Reply{Text: "⚠️ Formato incorreto — use:\n/gastei <valor> <categoria> [método]\n\nExemplo:\n/gastei 21,90 uber pix"}
}

func InvalidAmount() Reply {
	return Reply{Text: "❌ Valor inválido."}
}

func ExpenseRegistered(e domain.Expense) Reply {
	method := e.Method
	if method == "" {
		method = "Não informado"
	}

	return Reply{Text: fmt.Sprintf(
		"✅ Gasto registrado com sucesso!\n\n🆔 %s\n💰 %s\n📝 %s\n💳 %s",
		e.ID, formatMoney(e.AmountCents), e.Category, method,
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