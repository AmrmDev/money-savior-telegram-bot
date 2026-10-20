package controller

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"money-savior-telegram-bot/app/src/core/port"
)

type fakeDownloader struct {
	audio  []byte
	err    error
	calls  int
	fileID string
}

func (f *fakeDownloader) Download(_ context.Context, fileID string) ([]byte, error) {
	f.calls++
	f.fileID = fileID
	return f.audio, f.err
}

type fakeTranscriber struct {
	text  string
	err   error
	calls int
	audio []byte
}

func (f *fakeTranscriber) Transcribe(_ context.Context, audio []byte) (string, error) {
	f.calls++
	f.audio = audio
	return f.text, f.err
}

type fakeIntentParser struct {
	intent port.Intent
	err    error
	calls  int
	text   string
}

func (f *fakeIntentParser) Parse(_ context.Context, text string) (port.Intent, error) {
	f.calls++
	f.text = text
	return f.intent, f.err
}

type fakeMessageHandler struct {
	messages []Message
	err      error
}

func (f *fakeMessageHandler) Handle(_ context.Context, message Message) error {
	f.messages = append(f.messages, message)
	return f.err
}

type voiceFixture struct {
	controller  *VoiceController
	downloader  *fakeDownloader
	transcriber *fakeTranscriber
	parser      *fakeIntentParser
	queries     *fakeMessageHandler
	sender      *fakeSender
}

func newVoiceFixture(intent port.Intent) *voiceFixture {
	f := &voiceFixture{
		downloader:  &fakeDownloader{audio: []byte("ogg-bytes")},
		transcriber: &fakeTranscriber{text: "quanto eu gastei essa semana"},
		parser:      &fakeIntentParser{intent: intent},
		queries:     &fakeMessageHandler{},
		sender:      &fakeSender{},
	}
	f.controller = NewVoiceController(f.downloader, f.transcriber, f.parser, f.queries, f.sender)
	return f
}

func voiceMessage() Message {
	return Message{ChatID: 10, UserID: 20, VoiceFileID: "file-1", VoiceSeconds: 5}
}

func TestVoiceController_QueryGoesThroughTheQueryController(t *testing.T) {
	f := newVoiceFixture(port.Intent{Kind: port.IntentQuery, PeriodKeyword: "mes_passado"})

	if err := f.controller.Handle(context.Background(), voiceMessage()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if f.downloader.fileID != "file-1" || string(f.transcriber.audio) != "ogg-bytes" {
		t.Errorf("the audio did not travel through the pipeline")
	}
	if f.parser.text != "quanto eu gastei essa semana" {
		t.Errorf("the transcript was not sent to the parser: %q", f.parser.text)
	}

	want := []Message{{ChatID: 10, UserID: 20, Args: []string{"mes", "passado"}}}
	if !reflect.DeepEqual(f.queries.messages, want) {
		t.Errorf("unexpected query messages: %+v", f.queries.messages)
	}
	if len(f.sender.replies) != 0 {
		t.Errorf("the voice controller must leave the answer to the query controller, got %d replies", len(f.sender.replies))
	}
}

func TestVoiceController_QueryErrorIsReturnedAsIs(t *testing.T) {
	boom := errors.New("query failed")
	f := newVoiceFixture(port.Intent{Kind: port.IntentQuery, PeriodKeyword: "hoje"})
	f.queries.err = boom

	if err := f.controller.Handle(context.Background(), voiceMessage()); !errors.Is(err, boom) {
		t.Errorf("expected the query error, got %v", err)
	}
}

func TestPeriodArgs(t *testing.T) {
	tests := []struct {
		name   string
		intent port.Intent
		want   []string
	}{
		{"hoje", port.Intent{PeriodKeyword: "hoje"}, []string{"hoje"}},
		{"semana", port.Intent{PeriodKeyword: "semana"}, []string{"semana"}},
		{"mes passado com underline", port.Intent{PeriodKeyword: "mes_passado"}, []string{"mes", "passado"}},
		{"mes passado com espaço", port.Intent{PeriodKeyword: "mes passado"}, []string{"mes", "passado"}},
		{"datas", port.Intent{From: "12/10", To: "15/10"}, []string{"12/10", "15/10"}},
		{"palavra-chave vence as datas", port.Intent{PeriodKeyword: "ontem", From: "12/10", To: "15/10"}, []string{"ontem"}},
		{"sem período", port.Intent{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := periodArgs(tt.intent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestVoiceController_NoPeriodMeansThisMonth(t *testing.T) {
	f := newVoiceFixture(port.Intent{Kind: port.IntentQuery})

	if err := f.controller.Handle(context.Background(), voiceMessage()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.queries.messages) != 1 || len(f.queries.messages[0].Args) != 0 {
		t.Errorf("expected one query with no arguments, got %+v", f.queries.messages)
	}
}

func TestVoiceController_InvalidPeriodFromTheAIIsRejected(t *testing.T) {
	tests := []port.Intent{
		{Kind: port.IntentQuery, PeriodKeyword: "amanha"},
		{Kind: port.IntentQuery, PeriodKeyword: "mes_retrasado"},
		{Kind: port.IntentQuery, From: "12/10"},
		{Kind: port.IntentQuery, From: "12/10", To: "abc"},
	}

	for _, intent := range tests {
		f := newVoiceFixture(intent)

		err := f.controller.Handle(context.Background(), voiceMessage())

		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%+v: expected ErrInvalidInput, got %v", intent, err)
		}
		if len(f.queries.messages) != 0 {
			t.Errorf("%+v: the query must not run", intent)
		}
		if len(f.sender.replies) != 1 {
			t.Errorf("%+v: the user must get exactly one reply, got %d", intent, len(f.sender.replies))
		}
	}
}

func TestVoiceController_UnknownIntentAndEmptyTranscript(t *testing.T) {
	t.Run("intenção desconhecida", func(t *testing.T) {
		f := newVoiceFixture(port.Intent{Kind: port.IntentUnknown})
		f.transcriber.text = "minha senha é 1234"

		err := f.controller.Handle(context.Background(), voiceMessage())

		if !errors.Is(err, ErrInvalidInput) || strings.Contains(err.Error(), "1234") {
			t.Errorf("expected a clean ErrInvalidInput, got %v", err)
		}
		if len(f.sender.replies) != 1 {
			t.Errorf("expected one reply, got %d", len(f.sender.replies))
		}
	})

	t.Run("transcrição vazia não chama o interpretador", func(t *testing.T) {
		f := newVoiceFixture(port.Intent{Kind: port.IntentQuery, PeriodKeyword: "hoje"})
		f.transcriber.text = "   "

		err := f.controller.Handle(context.Background(), voiceMessage())

		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
		if f.parser.calls != 0 {
			t.Error("the parser must not run without a transcript")
		}
	})
}

func TestVoiceController_RejectsBeforeDownloading(t *testing.T) {
	t.Run("áudio longo demais", func(t *testing.T) {
		f := newVoiceFixture(port.Intent{Kind: port.IntentQuery})
		message := voiceMessage()
		message.VoiceSeconds = maxVoiceSeconds + 1

		err := f.controller.Handle(context.Background(), message)

		if !errors.Is(err, ErrInvalidInput) || f.downloader.calls != 0 {
			t.Errorf("expected a rejection before the download, got err=%v calls=%d", err, f.downloader.calls)
		}
		if len(f.sender.replies) != 1 || !strings.Contains(f.sender.replies[0].Text, "30") {
			t.Errorf("the reply should state the limit: %+v", f.sender.replies)
		}
	})

	t.Run("no limite exato ainda passa", func(t *testing.T) {
		f := newVoiceFixture(port.Intent{Kind: port.IntentQuery, PeriodKeyword: "hoje"})
		message := voiceMessage()
		message.VoiceSeconds = maxVoiceSeconds

		if err := f.controller.Handle(context.Background(), message); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("sem arquivo", func(t *testing.T) {
		f := newVoiceFixture(port.Intent{Kind: port.IntentQuery})
		message := voiceMessage()
		message.VoiceFileID = ""

		err := f.controller.Handle(context.Background(), message)

		if !errors.Is(err, ErrInvalidInput) || f.downloader.calls != 0 {
			t.Errorf("expected a rejection before the download, got err=%v calls=%d", err, f.downloader.calls)
		}
	})
}

func TestVoiceController_TechnicalFailuresAnswerAndReturnTheError(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name  string
		setup func(*voiceFixture)
	}{
		{"download", func(f *voiceFixture) { f.downloader.err = boom }},
		{"transcrição", func(f *voiceFixture) { f.transcriber.err = boom }},
		{"interpretação", func(f *voiceFixture) { f.parser.err = boom }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newVoiceFixture(port.Intent{Kind: port.IntentQuery, PeriodKeyword: "hoje"})
			tt.setup(f)

			err := f.controller.Handle(context.Background(), voiceMessage())

			if !errors.Is(err, boom) {
				t.Errorf("expected the original error, got %v", err)
			}
			if errors.Is(err, ErrInvalidInput) {
				t.Error("a technical failure must not be reported as invalid input")
			}
			if len(f.sender.replies) != 1 {
				t.Errorf("the user must get exactly one reply, got %d", len(f.sender.replies))
			}
			if len(f.queries.messages) != 0 {
				t.Error("the query must not run after a failure")
			}
		})
	}
}

func TestVoiceController_DatesThatDoNotExistGoToTheQueryController(t *testing.T) {
	f := newVoiceFixture(port.Intent{Kind: port.IntentQuery, From: "99/99", To: "15/10"})

	if err := f.controller.Handle(context.Background(), voiceMessage()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Message{{ChatID: 10, UserID: 20, Args: []string{"99/99", "15/10"}}}
	if !reflect.DeepEqual(f.queries.messages, want) {
		t.Errorf("unexpected query messages: %+v", f.queries.messages)
	}
}

func TestVoiceController_ExpenseIsPreviewedNotSaved(t *testing.T) {
	f := newVoiceFixture(port.Intent{Kind: port.IntentExpense, Amount: "3", Category: "água", Method: "pix"})

	if err := f.controller.Handle(context.Background(), voiceMessage()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.queries.messages) != 0 {
		t.Error("an expense must not run the query")
	}
	if len(f.sender.replies) != 1 {
		t.Fatalf("expected exactly one reply, got %d", len(f.sender.replies))
	}

	reply := f.sender.replies[0]
	if !strings.Contains(reply.Text, "Água, R$ 3,00 (Pix)") {
		t.Errorf("the preview must show the normalized expense: %q", reply.Text)
	}
	if len(reply.Buttons) != 2 || reply.Buttons[0].Data != "g|300|Água|Pix" || reply.Buttons[1].Data != "x" {
		t.Errorf("unexpected buttons: %+v", reply.Buttons)
	}
}

func TestVoiceController_PreviewWithoutMethod(t *testing.T) {
	f := newVoiceFixture(port.Intent{Kind: port.IntentExpense, Amount: "45,73", Category: "Mercado"})

	if err := f.controller.Handle(context.Background(), voiceMessage()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reply := f.sender.replies[0]
	if reply.Buttons[0].Data != "g|4573|Mercado|" || strings.Contains(reply.Text, "(") {
		t.Errorf("unexpected preview: %+v", reply)
	}
}

func TestVoiceController_PreviewAmounts(t *testing.T) {
	tests := []struct {
		amount string
		want   string
	}{
		{"3", "g|300|Água|"},
		{"0,5", "g|50|Água|"},
		{"45,73", "g|4573|Água|"},
		{"45.73", "g|4573|Água|"},
		{"1200", "g|120000|Água|"},
	}

	for _, tt := range tests {
		t.Run(tt.amount, func(t *testing.T) {
			f := newVoiceFixture(port.Intent{Kind: port.IntentExpense, Amount: tt.amount, Category: "Água"})

			if err := f.controller.Handle(context.Background(), voiceMessage()); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := f.sender.replies[0].Buttons[0].Data; got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVoiceController_InvalidExpenseFromTheAIIsRejected(t *testing.T) {
	tests := []struct {
		name   string
		intent port.Intent
	}{
		{"valor não numérico", port.Intent{Kind: port.IntentExpense, Amount: "abc", Category: "Mercado"}},
		{"valor vazio", port.Intent{Kind: port.IntentExpense, Amount: "", Category: "Mercado"}},
		{"valor zero", port.Intent{Kind: port.IntentExpense, Amount: "0", Category: "Mercado"}},
		{"valor negativo", port.Intent{Kind: port.IntentExpense, Amount: "-5", Category: "Mercado"}},
		{"separador de milhar", port.Intent{Kind: port.IntentExpense, Amount: "1.200,50", Category: "Mercado"}},
		{"categoria vazia", port.Intent{Kind: port.IntentExpense, Amount: "10", Category: ""}},
		{"categoria só com espaços", port.Intent{Kind: port.IntentExpense, Amount: "10", Category: "   "}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newVoiceFixture(tt.intent)

			err := f.controller.Handle(context.Background(), voiceMessage())

			if !errors.Is(err, ErrInvalidInput) || strings.Contains(err.Error(), "Mercado") {
				t.Errorf("expected a clean ErrInvalidInput, got %v", err)
			}
			if len(f.sender.replies) != 1 || len(f.sender.replies[0].Buttons) != 0 {
				t.Errorf("expected one reply without buttons: %+v", f.sender.replies)
			}
		})
	}
}

func TestVoiceController_ExpenseThatDoesNotFitTheButtonPointsToTheCommand(t *testing.T) {
	tests := []struct {
		name     string
		category string
	}{
		{"categoria longa demais", strings.Repeat("a", 70)},
		{"categoria com separador", "a|b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newVoiceFixture(port.Intent{Kind: port.IntentExpense, Amount: "10", Category: tt.category})

			err := f.controller.Handle(context.Background(), voiceMessage())

			if !errors.Is(err, ErrInvalidInput) || strings.Contains(err.Error(), tt.category) {
				t.Errorf("expected a clean ErrInvalidInput, got %v", err)
			}
			if len(f.sender.replies) != 1 || !strings.Contains(f.sender.replies[0].Text, "/gastei") || len(f.sender.replies[0].Buttons) != 0 {
				t.Errorf("the user should be pointed to /gastei: %+v", f.sender.replies)
			}
		})
	}
}
