package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"money-savior-telegram-bot/app/src/adapter/presenter"
)

// fakeTelegram answers the Bot API calls and records the parameters of each one.
type fakeTelegram struct {
	server *httptest.Server

	mu     sync.Mutex
	calls  map[string]url.Values
	failOn string
}

func newFakeTelegram(t *testing.T) *fakeTelegram {
	t.Helper()

	f := &fakeTelegram{calls: map[string]url.Values{}}
	f.server = httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_ = r.ParseForm()
		method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]

		f.mu.Lock()
		f.calls[method] = r.PostForm
		failing := method == f.failOn
		f.mu.Unlock()

		switch {
		case failing:
			_, _ = w.Write([]byte(`{"ok":false,"error_code":400,"description":"Bad Request: boom"}`))
		case method == "getMe":
			_, _ = w.Write([]byte(`{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"bot","username":"testbot"}}`))
		case method == "answerCallbackQuery":
			_, _ = w.Write([]byte(`{"ok":true,"result":true}`))
		default:
			_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":7,"date":0,"chat":{"id":10,"type":"private"}}}`))
		}
	}))
	t.Cleanup(f.server.Close)

	return f
}

func (f *fakeTelegram) sender(t *testing.T) *TelegramSender {
	t.Helper()

	bot, err := tgbotapi.NewBotAPIWithClient("123:TEST", f.server.URL+"/bot%s/%s", f.server.Client())
	if err != nil {
		t.Fatalf("cannot build the bot: %v", err)
	}
	return NewTelegramSender(bot)
}

func (f *fakeTelegram) failOnMethod(method string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failOn = method
}

func (f *fakeTelegram) call(method string) (url.Values, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	values, ok := f.calls[method]
	return values, ok
}

type keyboardJSON struct {
	InlineKeyboard [][]struct {
		Text         string `json:"text"`
		CallbackData string `json:"callback_data"`
	} `json:"inline_keyboard"`
}

func TestSend_PutsTheButtonsInOneRow(t *testing.T) {
	telegram := newFakeTelegram(t)
	sender := telegram.sender(t)

	err := sender.Send(context.Background(), 10, presenter.Reply{
		Text: "Entendi",
		Buttons: []presenter.Button{
			{Label: "✅ Confirmar", Data: "g|300|Água|"},
			{Label: "❌ Cancelar", Data: "x"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	values, ok := telegram.call("sendMessage")
	if !ok || values.Get("chat_id") != "10" || values.Get("text") != "Entendi" {
		t.Fatalf("unexpected sendMessage: %v", values)
	}

	var keyboard keyboardJSON
	if err := json.Unmarshal([]byte(values.Get("reply_markup")), &keyboard); err != nil {
		t.Fatalf("reply_markup is not valid JSON: %v", err)
	}
	if len(keyboard.InlineKeyboard) != 1 || len(keyboard.InlineKeyboard[0]) != 2 {
		t.Fatalf("expected one row with two buttons: %+v", keyboard)
	}

	confirm, cancel := keyboard.InlineKeyboard[0][0], keyboard.InlineKeyboard[0][1]
	if confirm.Text != "✅ Confirmar" || confirm.CallbackData != "g|300|Água|" || cancel.CallbackData != "x" {
		t.Errorf("unexpected buttons: %+v", keyboard.InlineKeyboard[0])
	}
}

func TestSend_WithoutButtonsHasNoKeyboard(t *testing.T) {
	telegram := newFakeTelegram(t)

	if err := telegram.sender(t).Send(context.Background(), 10, presenter.Reply{Text: "Oi"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	values, _ := telegram.call("sendMessage")
	if _, has := values["reply_markup"]; has {
		t.Errorf("a reply without buttons must not send a keyboard: %v", values)
	}
}

func TestEdit_ReplacesTheTextAndDropsTheKeyboard(t *testing.T) {
	telegram := newFakeTelegram(t)

	err := telegram.sender(t).Edit(context.Background(), 10, 7, presenter.Reply{Text: "Feito ✅"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	values, ok := telegram.call("editMessageText")
	if !ok {
		t.Fatal("editMessageText was not called")
	}
	if values.Get("chat_id") != "10" || values.Get("message_id") != "7" || values.Get("text") != "Feito ✅" {
		t.Errorf("unexpected edit: %v", values)
	}
	if _, has := values["reply_markup"]; has {
		t.Errorf("an edit must not send a keyboard, or the buttons would stay: %v", values)
	}
}

func TestAnswerCallback_SendsTheCallbackID(t *testing.T) {
	telegram := newFakeTelegram(t)

	if err := telegram.sender(t).AnswerCallback(context.Background(), "cb-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	values, ok := telegram.call("answerCallbackQuery")
	if !ok || values.Get("callback_query_id") != "cb-1" {
		t.Errorf("unexpected answerCallbackQuery: %v", values)
	}
}

func TestTelegramFailuresAreReturned(t *testing.T) {
	telegram := newFakeTelegram(t)
	sender := telegram.sender(t)

	telegram.failOnMethod("sendMessage")
	if err := sender.Send(context.Background(), 10, presenter.Reply{Text: "Oi"}); err == nil {
		t.Error("Send must return the Telegram error")
	}

	telegram.failOnMethod("editMessageText")
	if err := sender.Edit(context.Background(), 10, 7, presenter.Reply{Text: "Oi"}); err == nil {
		t.Error("Edit must return the Telegram error")
	}

	telegram.failOnMethod("answerCallbackQuery")
	if err := sender.AnswerCallback(context.Background(), "cb-1"); err == nil {
		t.Error("AnswerCallback must return the Telegram error")
	}
}
