package port

import "testing"

func TestParseIntentKind(t *testing.T) {
	tests := []struct {
		text string
		want IntentKind
	}{
		{"gasto", IntentExpense},
		{"consulta", IntentQuery},
		{"  Consulta ", IntentQuery},
		{"GASTO", IntentExpense},
		{"desconhecido", IntentUnknown},
		{"", IntentUnknown},
		{"deletar", IntentUnknown},
		{"gasto; drop table", IntentUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			if got := ParseIntentKind(tt.text); got != tt.want {
				t.Errorf("ParseIntentKind(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}