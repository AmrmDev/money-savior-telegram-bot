package port

import "strings"

// IntentKind is what the user wants, as understood from a voice message.
type IntentKind string

const (
	IntentExpense IntentKind = "gasto"
	IntentQuery   IntentKind = "consulta"
	IntentUnknown IntentKind = "desconhecido"
)

// ParseIntentKind turns the text returned by the AI into a known kind. Anything that is
// not on the list becomes IntentUnknown: the AI output is untrusted.
func ParseIntentKind(text string) IntentKind {
	switch IntentKind(strings.ToLower(strings.TrimSpace(text))) {
	case IntentExpense:
		return IntentExpense
	case IntentQuery:
		return IntentQuery
	default:
		return IntentUnknown
	}
}

// Intent is the raw result of interpreting a sentence. Every field is untrusted text:
// nothing here was validated, and the adapter must run it through the domain rules.
type Intent struct {
	Kind IntentKind

	// Expense fields. Amount is text such as "45,73", never a number.
	Amount   string
	Category string
	Method   string

	// Query fields. PeriodKeyword is one of hoje, ontem, semana, mes or mes_passado;
	// when the user names two days, From and To hold them as dd/mm.
	PeriodKeyword string
	From          string
	To            string
}