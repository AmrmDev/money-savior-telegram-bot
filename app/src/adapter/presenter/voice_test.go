package presenter

import (
	"strings"
	"testing"

	"money-savior-telegram-bot/app/src/core/domain"
)

func TestExpensePreview(t *testing.T) {
	reply := ExpensePreview(domain.Expense{Category: "Água", AmountCents: 300, Method: "Pix"}, "g|300|Água|Pix", "x")

	if !strings.Contains(reply.Text, "Água, R$ 3,00 (Pix)") || !strings.Contains(reply.Text, "Registrar?") {
		t.Errorf("unexpected text: %q", reply.Text)
	}
	if len(reply.Buttons) != 2 || reply.Buttons[0].Data != "g|300|Água|Pix" || reply.Buttons[1].Data != "x" {
		t.Errorf("unexpected buttons: %+v", reply.Buttons)
	}
}

func TestExpensePreview_WithoutMethod(t *testing.T) {
	reply := ExpensePreview(domain.Expense{Category: "Mercado", AmountCents: 4573}, "g|4573|Mercado|", "x")

	if !strings.Contains(reply.Text, "Mercado, R$ 45,73.") || strings.Contains(reply.Text, "(") {
		t.Errorf("unexpected text: %q", reply.Text)
	}
}
