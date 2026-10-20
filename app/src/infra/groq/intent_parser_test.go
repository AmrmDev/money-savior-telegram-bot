package groq

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"money-savior-telegram-bot/app/src/core/port"
)

func newTestIntentParser(serverURL string) *IntentParser {
	return newIntentParser("test-key", serverURL, &http.Client{Timeout: 2 * time.Second})
}

// completionBody builds a chat completion whose content is the given intent JSON.
func completionBody(t *testing.T, content, finishReason string) string {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"choices": []map[string]any{{
			"finish_reason": finishReason,
			"message":       map[string]any{"role": "assistant", "content": content},
		}},
	})
	if err != nil {
		t.Fatalf("cannot build the fake response: %v", err)
	}
	return string(body)
}

func serveBody(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
}

func TestParse_SendsTheExpectedRequest(t *testing.T) {
	sentence := "Quanto que eu gastei essa semana agora?"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("unexpected Authorization header: %q", got)
		}

		var sent map[string]any
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Fatalf("request is not JSON: %v", err)
		}

		if sent["model"] != "openai/gpt-oss-20b" || sent["reasoning_effort"] != "low" || sent["include_reasoning"] != false {
			t.Errorf("unexpected model settings: %v", sent)
		}

		messages := sent["messages"].([]any)
		system := messages[0].(map[string]any)
		user := messages[1].(map[string]any)
		if system["role"] != "system" || user["role"] != "user" || user["content"] != sentence {
			t.Errorf("unexpected messages: %v", messages)
		}

		format := sent["response_format"].(map[string]any)
		schema := format["json_schema"].(map[string]any)
		if format["type"] != "json_schema" || schema["strict"] != true {
			t.Errorf("the schema must be strict: %v", format)
		}

		body := schema["schema"].(map[string]any)
		if body["additionalProperties"] != false {
			t.Errorf("additionalProperties must be false: %v", body)
		}
		if required := body["required"].([]any); len(required) != len(body["properties"].(map[string]any)) {
			t.Errorf("every property must be required: %v", required)
		}

		_, _ = w.Write([]byte(completionBody(t,
			`{"kind":"consulta","amount":null,"category":null,"method":null,"period_keyword":"semana","from":null,"to":null}`,
			"stop")))
	}))
	defer server.Close()

	got, err := newTestIntentParser(server.URL).Parse(context.Background(), sentence)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != (port.Intent{Kind: port.IntentQuery, PeriodKeyword: "semana"}) {
		t.Errorf("unexpected intent: %+v", got)
	}
}

func TestParse_Intents(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    port.Intent
	}{
		{
			name:    "gasto completo",
			content: `{"kind":"gasto","amount":"45,73","category":" Mercado ","method":"Pix","period_keyword":null,"from":null,"to":null}`,
			want:    port.Intent{Kind: port.IntentExpense, Amount: "45,73", Category: "Mercado", Method: "Pix"},
		},
		{
			name:    "gasto sem método",
			content: `{"kind":"gasto","amount":"3","category":"Água","method":null,"period_keyword":null,"from":null,"to":null}`,
			want:    port.Intent{Kind: port.IntentExpense, Amount: "3", Category: "Água"},
		},
		{
			name:    "consulta por palavra-chave",
			content: `{"kind":"consulta","amount":null,"category":null,"method":null,"period_keyword":"Mes_Passado","from":null,"to":null}`,
			want:    port.Intent{Kind: port.IntentQuery, PeriodKeyword: "mes_passado"},
		},
		{
			name:    "consulta por datas",
			content: `{"kind":"consulta","amount":null,"category":null,"method":null,"period_keyword":null,"from":"12/10","to":"15/10"}`,
			want:    port.Intent{Kind: port.IntentQuery, From: "12/10", To: "15/10"},
		},
		{
			name:    "desconhecido",
			content: `{"kind":"desconhecido","amount":null,"category":null,"method":null,"period_keyword":null,"from":null,"to":null}`,
			want:    port.Intent{Kind: port.IntentUnknown},
		},
		{
			name:    "tipo inventado pela IA vira desconhecido",
			content: `{"kind":"deletar","amount":"10","category":"x","method":null,"period_keyword":null,"from":null,"to":null}`,
			want:    port.Intent{Kind: port.IntentUnknown},
		},
		{
			name:    "consulta não carrega valor",
			content: `{"kind":"consulta","amount":"99","category":"Mercado","method":"Pix","period_keyword":"hoje","from":null,"to":null}`,
			want:    port.Intent{Kind: port.IntentQuery, PeriodKeyword: "hoje"},
		},
		{
			name:    "gasto não carrega período",
			content: `{"kind":"gasto","amount":"5","category":"Café","method":null,"period_keyword":"hoje","from":"01/01","to":"02/01"}`,
			want:    port.Intent{Kind: port.IntentExpense, Amount: "5", Category: "Café"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := serveBody(completionBody(t, tt.content, "stop"))
			defer server.Close()

			got, err := newTestIntentParser(server.URL).Parse(context.Background(), "uma frase")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParse_EmptyOrHugeTextIsUnknownWithoutCallingTheAPI(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	defer server.Close()

	parser := newTestIntentParser(server.URL)

	for _, text := range []string{"", "   ", strings.Repeat("a", maxIntentTextRunes+1)} {
		got, err := parser.Parse(context.Background(), text)
		if err != nil || got.Kind != port.IntentUnknown {
			t.Errorf("expected unknown and no error, got %+v, %v", got, err)
		}
	}
	if called {
		t.Error("the API must not be called for text that is rejected up front")
	}
}

func TestParse_FailsOnBadResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"corte por tamanho", completionBody(t, `{"kind":"gasto"}`, "length")},
		{"conteúdo não é JSON", completionBody(t, `isso não é json`, "stop")},
		{"sem escolhas", `{"choices":[]}`},
		{"corpo inválido", `not json`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := serveBody(tt.body)
			defer server.Close()

			if _, err := newTestIntentParser(server.URL).Parse(context.Background(), "uma frase"); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestParse_ErrorsNeverLeakData(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusTooManyRequests, http.StatusInternalServerError} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":{"message":"details that must not leak"}}`))
		}))

		_, err := newTestIntentParser(server.URL).Parse(context.Background(), "gastei 10 no segredo")
		server.Close()

		if err == nil {
			t.Fatalf("status %d: expected an error", status)
		}
		for _, secret := range []string{"must not leak", "test-key", "segredo"} {
			if strings.Contains(err.Error(), secret) {
				t.Errorf("status %d: error leaks %q: %v", status, secret, err)
			}
		}
	}

	server := serveBody(completionBody(t, `texto com segredo que não é json`, "stop"))
	defer server.Close()

	_, err := newTestIntentParser(server.URL).Parse(context.Background(), "uma frase")
	if err == nil || strings.Contains(err.Error(), "segredo") {
		t.Errorf("a bad intent must fail without echoing the content: %v", err)
	}
}