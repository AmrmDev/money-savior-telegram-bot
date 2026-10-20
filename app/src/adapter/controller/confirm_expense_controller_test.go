package controller

import (
	"context"
	"errors"
	"strings"
	"testing"

	"money-savior-telegram-bot/app/src/adapter/presenter"
	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/domain/exception"
	"money-savior-telegram-bot/app/src/core/usecase/input"
)

type stubRegisterUseCase struct {
	expense domain.Expense
	err     error
	inputs  []input.RegisterExpense
}

func (s *stubRegisterUseCase) Execute(_ context.Context, in input.RegisterExpense) (domain.Expense, error) {
	s.inputs = append(s.inputs, in)
	return s.expense, s.err
}

type recordedEdit struct {
	chatID    int64
	messageID int
	text      string
}

type stubResponder struct {
	edits     []recordedEdit
	answered  []string
	editErr   error
	answerErr error
}

func (s *stubResponder) Edit(_ context.Context, chatID int64, messageID int, reply presenter.Reply) error {
	s.edits = append(s.edits, recordedEdit{chatID: chatID, messageID: messageID, text: reply.Text})
	return s.editErr
}

func (s *stubResponder) AnswerCallback(_ context.Context, callbackID string) error {
	s.answered = append(s.answered, callbackID)
	return s.answerErr
}

func newConfirmFixture() (*ConfirmExpenseController, *stubRegisterUseCase, *stubResponder) {
	usecase := &stubRegisterUseCase{expense: domain.Expense{ID: "123MSB456PA", AmountCents: 300, Category: "Água", Method: "Pix"}}
	responder := &stubResponder{}
	return NewConfirmExpenseController(usecase, responder), usecase, responder
}

func pressed(data string) Callback {
	return Callback{ID: "cb-1", ChatID: 10, UserID: 20, MessageID: 7, Data: data}
}

func TestConfirmExpense_ConfirmRegistersAndEditsTheMessage(t *testing.T) {
	controller, usecase, responder := newConfirmFixture()

	if err := controller.Handle(context.Background(), pressed("g|300|Água|Pix")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := input.RegisterExpense{UserID: 20, AmountCents: 300, Category: "Água", Method: "Pix"}
	if len(usecase.inputs) != 1 || usecase.inputs[0] != want {
		t.Errorf("unexpected use case input: %+v", usecase.inputs)
	}

	if len(responder.edits) != 1 {
		t.Fatalf("expected one edit, got %d", len(responder.edits))
	}
	edit := responder.edits[0]
	if edit.chatID != 10 || edit.messageID != 7 || !strings.Contains(edit.text, "gasto registrado") {
		t.Errorf("unexpected edit: %+v", edit)
	}
	if len(responder.answered) != 1 || responder.answered[0] != "cb-1" {
		t.Errorf("the callback must be answered once: %v", responder.answered)
	}
}

func TestConfirmExpense_CancelDoesNotRegister(t *testing.T) {
	controller, usecase, responder := newConfirmFixture()

	if err := controller.Handle(context.Background(), pressed("x")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(usecase.inputs) != 0 {
		t.Error("cancel must not register anything")
	}
	if len(responder.edits) != 1 || !strings.Contains(responder.edits[0].text, "Cancelado") {
		t.Errorf("unexpected edits: %+v", responder.edits)
	}
	if len(responder.answered) != 1 {
		t.Errorf("the callback must be answered once: %v", responder.answered)
	}
}

func TestConfirmExpense_InvalidButtonDataIsRejected(t *testing.T) {
	for _, data := range []string{"", "z", "g|abc|Segredo|Pix", "g|0|Água|Pix", "g|300||Pix", "g|300|Água|Pix|extra"} {
		controller, usecase, responder := newConfirmFixture()

		err := controller.Handle(context.Background(), pressed(data))

		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%q: expected ErrInvalidInput, got %v", data, err)
		}
		if strings.Contains(err.Error(), "Segredo") {
			t.Errorf("%q: the error carries the button data: %v", data, err)
		}
		if len(usecase.inputs) != 0 {
			t.Errorf("%q: nothing may be registered", data)
		}
		if len(responder.edits) != 1 || !strings.Contains(responder.edits[0].text, "botão") {
			t.Errorf("%q: the user should be told the button is no longer valid: %+v", data, responder.edits)
		}
	}
}

func TestConfirmExpense_DomainRejectionIsInvalidInput(t *testing.T) {
	controller, usecase, responder := newConfirmFixture()
	usecase.err = exception.ErrInvalidAmount

	err := controller.Handle(context.Background(), pressed("g|300|Água|Pix"))

	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
	if len(responder.edits) != 1 || !strings.Contains(responder.edits[0].text, "Valor inválido") {
		t.Errorf("unexpected edits: %+v", responder.edits)
	}
}

func TestConfirmExpense_TechnicalFailureIsReturnedAndShown(t *testing.T) {
	controller, usecase, responder := newConfirmFixture()
	boom := errors.New("dynamo down")
	usecase.err = boom

	err := controller.Handle(context.Background(), pressed("g|300|Água|Pix"))

	if !errors.Is(err, boom) || errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected the original error, not invalid input: %v", err)
	}
	if len(responder.edits) != 1 || !strings.Contains(responder.edits[0].text, "Erro ao salvar") {
		t.Errorf("the user must be told the save failed: %+v", responder.edits)
	}
}

func TestConfirmExpense_FailingToAnswerDoesNotStopTheExpense(t *testing.T) {
	controller, usecase, responder := newConfirmFixture()
	boom := errors.New("query is too old")
	responder.answerErr = boom

	err := controller.Handle(context.Background(), pressed("g|300|Água|Pix"))

	if !errors.Is(err, boom) {
		t.Errorf("the answer error should be reported: %v", err)
	}
	if len(usecase.inputs) != 1 || len(responder.edits) != 1 {
		t.Errorf("the expense must still be registered and the message edited: inputs=%d edits=%d", len(usecase.inputs), len(responder.edits))
	}
}

func TestConfirmExpense_FailingToEditIsReported(t *testing.T) {
	controller, usecase, responder := newConfirmFixture()
	boom := errors.New("message can't be edited")
	responder.editErr = boom

	err := controller.Handle(context.Background(), pressed("g|300|Água|Pix"))

	if !errors.Is(err, boom) {
		t.Errorf("the edit error should be reported: %v", err)
	}
	if len(usecase.inputs) != 1 {
		t.Error("the expense was already registered")
	}
}

func TestConfirmExpense_CallbackWithoutMessageIsIgnored(t *testing.T) {
	controller, usecase, responder := newConfirmFixture()
	callback := pressed("g|300|Água|Pix")
	callback.MessageID = 0

	err := controller.Handle(context.Background(), callback)

	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
	if len(usecase.inputs) != 0 || len(responder.edits) != 0 {
		t.Error("nothing may be registered or edited without a message")
	}
	if len(responder.answered) != 1 {
		t.Error("the button must still stop spinning")
	}
}
