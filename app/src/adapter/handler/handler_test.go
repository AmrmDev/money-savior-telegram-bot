package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"money-savior-telegram-bot/app/src/adapter/controller"
	"money-savior-telegram-bot/app/src/adapter/presenter"
	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/domain/exception"
	"money-savior-telegram-bot/app/src/core/usecase"
	"money-savior-telegram-bot/app/src/core/usecase/input"
)

type fakeSender struct {
	replies []presenter.Reply
}

func (f *fakeSender) Send(_ context.Context, _ int64, reply presenter.Reply) error {
	f.replies = append(f.replies, reply)
	return nil
}

type fakeRegister struct {
	expense domain.Expense
	err     error
}

func (f *fakeRegister) Execute(context.Context, input.RegisterExpense) (domain.Expense, error) {
	return f.expense, f.err
}

type fakeQuery struct{}

func (f *fakeQuery) Execute(context.Context, input.QueryExpenses) (usecase.ExpensesSummary, error) {
	return usecase.ExpensesSummary{}, nil
}

func newTestHandler(sender *fakeSender, register *fakeRegister, logs *bytes.Buffer) *Handler {
	return NewHandler(
		controller.NewStartController(sender),
		controller.NewHelpController(sender),
		controller.NewRegisterExpenseController(register, sender),
		controller.NewQueryExpensesController(&fakeQuery{}, sender),
		slog.New(slog.NewJSONHandler(logs, nil)),
	)
}

func commandUpdate(updateID int, text string) tgbotapi.Update {
	commandLength := strings.Index(text+" ", " ")

	return tgbotapi.Update{
		UpdateID: updateID,
		Message: &tgbotapi.Message{
			Text:     text,
			Chat:     &tgbotapi.Chat{ID: 10},
			From:     &tgbotapi.User{ID: 20, UserName: "tester"},
			Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: commandLength}},
		},
	}
}

func lastEntry(t *testing.T, logs *bytes.Buffer) map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(logs.String()), "\n")

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("log line is not valid JSON: %v", err)
	}
	return entry
}

func TestHandle_LogsOutcome(t *testing.T) {
	registered := domain.Expense{ID: "123MSB456PA", AmountCents: 5090, Category: "Mercado", Method: "Pix"}

	tests := []struct {
		name        string
		text        string
		registerErr error
		wantLevel   string
		wantOutcome string
		wantCommand string
	}{
		{"start", "/start", nil, "INFO", "ok", "start"},
		{"gasto registrado", "/gastei 50,90 mercado pix", nil, "INFO", "ok", "gastei"},
		{"comando desconhecido cai na ajuda", "/xyz", nil, "INFO", "ok", "help"},
		{"formato do valor inválido", "/gastei abc mercado", nil, "WARN", "invalid_input", "gastei"},
		{"argumentos faltando", "/gastei 50", nil, "WARN", "invalid_input", "gastei"},
		{"regra do domain", "/gastei 0 mercado", exception.ErrInvalidAmount, "WARN", "invalid_input", "gastei"},
		{"falha técnica", "/gastei 50,90 mercado pix", errors.New("dynamo down"), "ERROR", "error", "gastei"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			sender := &fakeSender{}
			h := newTestHandler(sender, &fakeRegister{expense: registered, err: tt.registerErr}, &logs)

			h.Handle(context.Background(), commandUpdate(7, tt.text))

			entry := lastEntry(t, &logs)
			if entry["level"] != tt.wantLevel || entry["outcome"] != tt.wantOutcome || entry["command"] != tt.wantCommand {
				t.Errorf("unexpected entry: %v", entry)
			}
			if entry["update_id"] != float64(7) || entry["user_id"] != float64(20) {
				t.Errorf("missing identifiers: %v", entry)
			}
			if _, ok := entry["duration_ms"]; !ok {
				t.Errorf("missing duration_ms: %v", entry)
			}
			if len(sender.replies) != 1 {
				t.Errorf("the user must always get exactly one reply, got %d", len(sender.replies))
			}
		})
	}
}

func TestHandle_NeverLogsUserData(t *testing.T) {
	var logs bytes.Buffer
	sender := &fakeSender{}
	register := &fakeRegister{expense: domain.Expense{ID: "123MSB456PA", AmountCents: 5090, Category: "Mercado"}}
	h := newTestHandler(sender, register, &logs)

	h.Handle(context.Background(), commandUpdate(1, "/gastei 50,90 mercado pix"))
	h.Handle(context.Background(), commandUpdate(2, "/gastei abc supermercado"))
	h.Handle(context.Background(), commandUpdate(3, "/minhasenha123"))

	output := strings.ToLower(logs.String())
	for _, secret := range []string{"50,90", "mercado", "pix", "abc", "minhasenha", "tester", "5090", "123msb456pa"} {
		if strings.Contains(output, secret) {
			t.Errorf("logs must not contain %q:\n%s", secret, logs.String())
		}
	}
}

func TestHandle_IgnoresUpdateWithoutMessage(t *testing.T) {
	var logs bytes.Buffer
	sender := &fakeSender{}
	h := newTestHandler(sender, &fakeRegister{}, &logs)

	h.Handle(context.Background(), tgbotapi.Update{UpdateID: 1})

	if len(sender.replies) != 0 {
		t.Errorf("expected no reply, got %d", len(sender.replies))
	}
	if logs.Len() != 0 {
		t.Errorf("an ignored update must not log at the default level: %s", logs.String())
	}
}