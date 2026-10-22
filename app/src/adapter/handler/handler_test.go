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
	"money-savior-telegram-bot/app/src/core/port"
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
		newVoice(sender, &fakeDownloader{}),
		newConfirm(&countingRegister{}, &fakeResponder{}),
		0,
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

// newRestrictedHandler builds a handler that only answers to userID 20 (the user
// used by commandUpdate) and logs from DEBUG up, so blocked updates show in the buffer.
func newRestrictedHandler(sender *fakeSender, register *fakeRegister, logs *bytes.Buffer) *Handler {
	return NewHandler(
		controller.NewStartController(sender),
		controller.NewHelpController(sender),
		controller.NewRegisterExpenseController(register, sender),
		controller.NewQueryExpensesController(&fakeQuery{}, sender),
		newVoice(sender, &fakeDownloader{}),
		newConfirm(&countingRegister{}, &fakeResponder{}),
		20,
		slog.New(slog.NewJSONHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	)
}

func TestHandle_AllowedUserIsServed(t *testing.T) {
	var logs bytes.Buffer
	sender := &fakeSender{}
	register := &fakeRegister{expense: domain.Expense{ID: "123MSB456PA", AmountCents: 5090, Category: "Mercado"}}
	h := newRestrictedHandler(sender, register, &logs)

	h.Handle(context.Background(), commandUpdate(1, "/gastei 50,90 mercado"))

	if len(sender.replies) != 1 {
		t.Fatalf("expected one reply, got %d", len(sender.replies))
	}
	if entry := lastEntry(t, &logs); entry["outcome"] != "ok" {
		t.Errorf("unexpected entry: %v", entry)
	}
}

func TestHandle_OtherUserIsIgnoredInSilence(t *testing.T) {
	var logs bytes.Buffer
	sender := &fakeSender{}
	h := newRestrictedHandler(sender, &fakeRegister{}, &logs)

	update := commandUpdate(9, "/gastei 50,90 mercado pix")
	update.Message.From.ID = 999

	h.Handle(context.Background(), update)

	if len(sender.replies) != 0 {
		t.Errorf("a blocked user must get no reply, got %d", len(sender.replies))
	}

	entry := lastEntry(t, &logs)
	if entry["level"] != "DEBUG" || entry["update_id"] != float64(9) || entry["user_id"] != float64(999) {
		t.Errorf("unexpected entry: %v", entry)
	}
	if _, ok := entry["outcome"]; ok {
		t.Errorf("a blocked update must not look like a handled one: %v", entry)
	}
}

func TestHandle_BlockedUserNeverReachesTheUseCase(t *testing.T) {
	var logs bytes.Buffer
	sender := &fakeSender{}
	register := &countingRegister{}
	h := NewHandler(
		controller.NewStartController(sender),
		controller.NewHelpController(sender),
		controller.NewRegisterExpenseController(register, sender),
		controller.NewQueryExpensesController(&fakeQuery{}, sender),
		newVoice(sender, &fakeDownloader{}),
		newConfirm(&countingRegister{}, &fakeResponder{}),
		20,
		slog.New(slog.NewJSONHandler(&logs, nil)),
	)

	update := commandUpdate(1, "/gastei 50,90 mercado pix")
	update.Message.From.ID = 999
	h.Handle(context.Background(), update)

	if register.calls != 0 {
		t.Errorf("the use case must not run for a blocked user, ran %d times", register.calls)
	}
}

func TestHandle_BlockedUserLogsNoUserData(t *testing.T) {
	var logs bytes.Buffer
	h := newRestrictedHandler(&fakeSender{}, &fakeRegister{}, &logs)

	update := commandUpdate(1, "/gastei 50,90 mercado pix")
	update.Message.From.ID = 999
	h.Handle(context.Background(), update)

	output := strings.ToLower(logs.String())
	for _, secret := range []string{"50,90", "mercado", "pix", "gastei", "tester"} {
		if strings.Contains(output, secret) {
			t.Errorf("logs must not contain %q:\n%s", secret, logs.String())
		}
	}
}

type countingRegister struct{ calls int }

func (c *countingRegister) Execute(context.Context, input.RegisterExpense) (domain.Expense, error) {
	c.calls++
	return domain.Expense{}, nil
}

type fakeDownloader struct {
	err   error
	calls int
}

func (f *fakeDownloader) Download(context.Context, string) ([]byte, error) {
	f.calls++
	return []byte("ogg"), f.err
}

type fakeTranscriber struct{}

func (fakeTranscriber) Transcribe(context.Context, []byte) (string, error) {
	return "quanto gastei com segredoxyz hoje", nil
}

type fakeParser struct{}

func (fakeParser) Parse(context.Context, string) (port.Intent, error) {
	return port.Intent{Kind: port.IntentQuery, PeriodKeyword: "hoje"}, nil
}

func newVoice(sender *fakeSender, downloader *fakeDownloader) *controller.VoiceController {
	return controller.NewVoiceController(
		downloader,
		fakeTranscriber{},
		fakeParser{},
		controller.NewQueryExpensesController(&fakeQuery{}, sender),
		sender,
	)
}

func newVoiceHandler(sender *fakeSender, downloader *fakeDownloader, allowedUserID int64, logs *bytes.Buffer) *Handler {
	return NewHandler(
		controller.NewStartController(sender),
		controller.NewHelpController(sender),
		controller.NewRegisterExpenseController(&fakeRegister{}, sender),
		controller.NewQueryExpensesController(&fakeQuery{}, sender),
		newVoice(sender, downloader),
		newConfirm(&countingRegister{}, &fakeResponder{}),
		allowedUserID,
		slog.New(slog.NewJSONHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	)
}

func voiceUpdate(updateID int, userID int64, seconds int) tgbotapi.Update {
	return tgbotapi.Update{
		UpdateID: updateID,
		Message: &tgbotapi.Message{
			Chat:  &tgbotapi.Chat{ID: 10},
			From:  &tgbotapi.User{ID: userID, UserName: "tester"},
			Voice: &tgbotapi.Voice{FileID: "file-1", Duration: seconds},
		},
	}
}

func TestHandle_VoiceIsRoutedToTheVoiceController(t *testing.T) {
	var logs bytes.Buffer
	sender := &fakeSender{}
	downloader := &fakeDownloader{}
	h := newVoiceHandler(sender, downloader, 20, &logs)

	h.Handle(context.Background(), voiceUpdate(5, 20, 4))

	entry := lastEntry(t, &logs)
	if entry["level"] != "INFO" || entry["outcome"] != "ok" || entry["command"] != "voz" {
		t.Errorf("unexpected entry: %v", entry)
	}
	if entry["update_id"] != float64(5) || entry["user_id"] != float64(20) {
		t.Errorf("missing identifiers: %v", entry)
	}
	if downloader.calls != 1 || len(sender.replies) != 1 {
		t.Errorf("expected one download and one reply, got %d and %d", downloader.calls, len(sender.replies))
	}
}

func TestHandle_VoiceFromAnotherUserIsIgnored(t *testing.T) {
	var logs bytes.Buffer
	sender := &fakeSender{}
	downloader := &fakeDownloader{}
	h := newVoiceHandler(sender, downloader, 20, &logs)

	h.Handle(context.Background(), voiceUpdate(6, 999, 4))

	if downloader.calls != 0 || len(sender.replies) != 0 {
		t.Errorf("a blocked user must not trigger a download or a reply, got %d and %d", downloader.calls, len(sender.replies))
	}
}

func TestHandle_VoiceTooLongIsRejected(t *testing.T) {
	var logs bytes.Buffer
	sender := &fakeSender{}
	downloader := &fakeDownloader{}
	h := newVoiceHandler(sender, downloader, 0, &logs)

	h.Handle(context.Background(), voiceUpdate(7, 20, 31))

	entry := lastEntry(t, &logs)
	if entry["level"] != "WARN" || entry["outcome"] != "invalid_input" || entry["command"] != "voz" {
		t.Errorf("unexpected entry: %v", entry)
	}
	if downloader.calls != 0 || len(sender.replies) != 1 {
		t.Errorf("expected no download and one reply, got %d and %d", downloader.calls, len(sender.replies))
	}
}

func TestHandle_VoiceFailureIsLoggedAsError(t *testing.T) {
	var logs bytes.Buffer
	sender := &fakeSender{}
	h := newVoiceHandler(sender, &fakeDownloader{err: errors.New("telegram down")}, 0, &logs)

	h.Handle(context.Background(), voiceUpdate(8, 20, 4))

	entry := lastEntry(t, &logs)
	if entry["level"] != "ERROR" || entry["outcome"] != "error" || entry["command"] != "voz" {
		t.Errorf("unexpected entry: %v", entry)
	}
	if len(sender.replies) != 1 {
		t.Errorf("the user must get exactly one reply, got %d", len(sender.replies))
	}
}

func TestHandle_NeverLogsVoiceContent(t *testing.T) {
	var logs bytes.Buffer
	h := newVoiceHandler(&fakeSender{}, &fakeDownloader{}, 0, &logs)

	h.Handle(context.Background(), voiceUpdate(9, 20, 4))

	output := strings.ToLower(logs.String())
	for _, secret := range []string{"segredoxyz", "quanto", "tester", "file-1"} {
		if strings.Contains(output, secret) {
			t.Errorf("logs must not contain %q:\n%s", secret, logs.String())
		}
	}
}

type fakeResponder struct {
	edits    []string
	answered []string
}

func (f *fakeResponder) Edit(_ context.Context, _ int64, _ int, reply presenter.Reply) error {
	f.edits = append(f.edits, reply.Text)
	return nil
}

func (f *fakeResponder) AnswerCallback(_ context.Context, callbackID string) error {
	f.answered = append(f.answered, callbackID)
	return nil
}

func newConfirm(register controller.RegisterExpenseUseCase, responder *fakeResponder) *controller.ConfirmExpenseController {
	return controller.NewConfirmExpenseController(register, responder)
}

func newCallbackHandler(register *countingRegister, responder *fakeResponder, allowedUserID int64, logs *bytes.Buffer) *Handler {
	sender := &fakeSender{}
	return NewHandler(
		controller.NewStartController(sender),
		controller.NewHelpController(sender),
		controller.NewRegisterExpenseController(&fakeRegister{}, sender),
		controller.NewQueryExpensesController(&fakeQuery{}, sender),
		newVoice(sender, &fakeDownloader{}),
		newConfirm(register, responder),
		allowedUserID,
		slog.New(slog.NewJSONHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	)
}

func callbackUpdate(updateID int, userID int64, data string) tgbotapi.Update {
	return tgbotapi.Update{
		UpdateID: updateID,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:      "cb-1",
			From:    &tgbotapi.User{ID: userID, UserName: "tester"},
			Message: &tgbotapi.Message{MessageID: 7, Chat: &tgbotapi.Chat{ID: 10}},
			Data:    data,
		},
	}
}

func TestHandle_ConfirmButtonRegistersTheExpense(t *testing.T) {
	var logs bytes.Buffer
	register := &countingRegister{}
	responder := &fakeResponder{}
	h := newCallbackHandler(register, responder, 20, &logs)

	h.Handle(context.Background(), callbackUpdate(11, 20, "g|300|Água|Pix"))

	entry := lastEntry(t, &logs)
	if entry["level"] != "INFO" || entry["outcome"] != "ok" || entry["command"] != "callback" {
		t.Errorf("unexpected entry: %v", entry)
	}
	if entry["update_id"] != float64(11) || entry["user_id"] != float64(20) {
		t.Errorf("missing identifiers: %v", entry)
	}
	if register.calls != 1 || len(responder.edits) != 1 || len(responder.answered) != 1 {
		t.Errorf("expected one registration, one edit and one answer, got %d, %d, %d",
			register.calls, len(responder.edits), len(responder.answered))
	}
}

func TestHandle_CancelButtonRegistersNothing(t *testing.T) {
	var logs bytes.Buffer
	register := &countingRegister{}
	responder := &fakeResponder{}
	h := newCallbackHandler(register, responder, 20, &logs)

	h.Handle(context.Background(), callbackUpdate(12, 20, "x"))

	if entry := lastEntry(t, &logs); entry["outcome"] != "ok" || entry["command"] != "callback" {
		t.Errorf("unexpected entry: %v", entry)
	}
	if register.calls != 0 || len(responder.edits) != 1 || !strings.Contains(responder.edits[0], "Cancelado") {
		t.Errorf("unexpected result: calls=%d edits=%v", register.calls, responder.edits)
	}
}

func TestHandle_ButtonFromAnotherUserIsIgnored(t *testing.T) {
	var logs bytes.Buffer
	register := &countingRegister{}
	responder := &fakeResponder{}
	h := newCallbackHandler(register, responder, 20, &logs)

	h.Handle(context.Background(), callbackUpdate(13, 999, "g|300|Água|Pix"))

	if register.calls != 0 || len(responder.edits) != 0 || len(responder.answered) != 0 {
		t.Errorf("a blocked user must trigger nothing, got %d, %d, %d",
			register.calls, len(responder.edits), len(responder.answered))
	}

	entry := lastEntry(t, &logs)
	if entry["level"] != "DEBUG" || entry["update_id"] != float64(13) || entry["user_id"] != float64(999) {
		t.Errorf("unexpected entry: %v", entry)
	}
	if _, ok := entry["outcome"]; ok {
		t.Errorf("a blocked update must not look like a handled one: %v", entry)
	}
}

func TestHandle_InvalidButtonIsRejected(t *testing.T) {
	var logs bytes.Buffer
	register := &countingRegister{}
	responder := &fakeResponder{}
	h := newCallbackHandler(register, responder, 0, &logs)

	h.Handle(context.Background(), callbackUpdate(14, 20, "lixo"))

	entry := lastEntry(t, &logs)
	if entry["level"] != "WARN" || entry["outcome"] != "invalid_input" || entry["command"] != "callback" {
		t.Errorf("unexpected entry: %v", entry)
	}
	if register.calls != 0 {
		t.Error("nothing may be registered for an invalid button")
	}
}

func TestHandle_ButtonWithoutMessageIsRejected(t *testing.T) {
	var logs bytes.Buffer
	register := &countingRegister{}
	responder := &fakeResponder{}
	h := newCallbackHandler(register, responder, 0, &logs)

	update := callbackUpdate(15, 20, "g|300|Água|Pix")
	update.CallbackQuery.Message = nil
	h.Handle(context.Background(), update)

	if entry := lastEntry(t, &logs); entry["outcome"] != "invalid_input" {
		t.Errorf("unexpected entry: %v", entry)
	}
	if register.calls != 0 || len(responder.edits) != 0 || len(responder.answered) != 1 {
		t.Errorf("expected only the answer, got %d, %d, %d", register.calls, len(responder.edits), len(responder.answered))
	}
}

func TestHandle_NeverLogsButtonData(t *testing.T) {
	var logs bytes.Buffer
	h := newCallbackHandler(&countingRegister{}, &fakeResponder{}, 0, &logs)

	h.Handle(context.Background(), callbackUpdate(16, 20, "g|300|Segredoxyz|Pix"))
	h.Handle(context.Background(), callbackUpdate(17, 20, "g|abc|Segredoxyz|Pix"))

	output := strings.ToLower(logs.String())
	for _, secret := range []string{"segredoxyz", "tester", "g|300", "g|abc"} {
		if strings.Contains(output, secret) {
			t.Errorf("logs must not contain %q:\n%s", secret, logs.String())
		}
	}
}
