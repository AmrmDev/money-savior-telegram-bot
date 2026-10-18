package controller

import (
	"testing"

	"money-savior-telegram-bot/app/src/core/domain"
)

func TestParsePeriodSpec_Valid(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want domain.PeriodSpec
	}{
		{"sem argumento é o mês atual", nil, domain.PeriodSpec{Kind: domain.PeriodThisMonth}},
		{"hoje", []string{"hoje"}, domain.PeriodSpec{Kind: domain.PeriodToday}},
		{"ontem", []string{"ontem"}, domain.PeriodSpec{Kind: domain.PeriodYesterday}},
		{"semana", []string{"semana"}, domain.PeriodSpec{Kind: domain.PeriodLast7Days}},
		{"mes", []string{"mes"}, domain.PeriodSpec{Kind: domain.PeriodThisMonth}},
		{"mês com acento", []string{"mês"}, domain.PeriodSpec{Kind: domain.PeriodThisMonth}},
		{"maiúsculas", []string{"HOJE"}, domain.PeriodSpec{Kind: domain.PeriodToday}},
		{"mes passado", []string{"mes", "passado"}, domain.PeriodSpec{Kind: domain.PeriodLastMonth}},
		{"mês passado com acento", []string{"Mês", "Passado"}, domain.PeriodSpec{Kind: domain.PeriodLastMonth}},
		{
			"intervalo",
			[]string{"01/10", "15/10"},
			domain.PeriodSpec{Kind: domain.PeriodCustom, From: domain.Date{Month: 10, Day: 1}, To: domain.Date{Month: 10, Day: 15}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePeriodSpec(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParsePeriodSpec_Invalid(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"palavra desconhecida", []string{"amanha"}},
		{"data sozinha", []string{"01/10"}},
		{"segunda data inválida", []string{"01/10", "abc"}},
		{"data com ano", []string{"01/10/2026", "15/10/2026"}},
		{"mes retrasado", []string{"mes", "retrasado"}},
		{"argumentos demais", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parsePeriodSpec(tt.args); err == nil {
				t.Error("expected an error")
			}
		})
	}
}
