package groq

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestTranscriber(serverURL string) *Transcriber {
	return newTranscriber("test-key", serverURL, &http.Client{Timeout: 2 * time.Second})
}

func TestTranscribe_SendsTheExpectedRequest(t *testing.T) {
	audio := []byte("OggS-fake-audio")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/audio/transcriptions" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("unexpected Authorization header: %q", got)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("request is not multipart: %v", err)
		}
		for field, want := range map[string]string{
			"model":           "whisper-large-v3-turbo",
			"language":        "pt",
			"response_format": "json",
		} {
			if got := r.FormValue(field); got != want {
				t.Errorf("field %s = %q, want %q", field, got, want)
			}
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("missing file: %v", err)
		}
		defer file.Close()
		content, _ := io.ReadAll(file)
		if string(content) != string(audio) || header.Filename != "voice.ogg" {
			t.Errorf("unexpected file: name=%q content=%q", header.Filename, content)
		}

		_, _ = w.Write([]byte(`{"text":" Quanto que eu gastei essa semana agora?","x_groq":{"id":"req_1"}}`))
	}))
	defer server.Close()

	text, err := newTestTranscriber(server.URL).Transcribe(context.Background(), audio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "Quanto que eu gastei essa semana agora?" {
		t.Errorf("unexpected text: %q", text)
	}
}

func TestTranscribe_NoSpeechReturnsEmptyText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"text":"  "}`))
	}))
	defer server.Close()

	text, err := newTestTranscriber(server.URL).Transcribe(context.Background(), []byte("audio"))
	if err != nil || text != "" {
		t.Errorf("expected empty text and no error, got %q, %v", text, err)
	}
}

func TestTranscribe_FailsOnErrorStatus(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusTooManyRequests, http.StatusInternalServerError} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":{"message":"details that must not leak"}}`))
		}))

		_, err := newTestTranscriber(server.URL).Transcribe(context.Background(), []byte("audio"))
		server.Close()

		if err == nil {
			t.Fatalf("status %d: expected an error", status)
		}
		if strings.Contains(err.Error(), "must not leak") || strings.Contains(err.Error(), "test-key") {
			t.Errorf("status %d: error leaks data: %v", status, err)
		}
	}
}

func TestTranscribe_FailsOnInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	if _, err := newTestTranscriber(server.URL).Transcribe(context.Background(), []byte("audio")); err == nil {
		t.Error("expected an error for an invalid response")
	}
}

func TestTranscribe_RejectsBadAudioWithoutCallingTheAPI(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	defer server.Close()

	transcriber := newTestTranscriber(server.URL)

	if _, err := transcriber.Transcribe(context.Background(), nil); err == nil {
		t.Error("expected an error for empty audio")
	}
	if _, err := transcriber.Transcribe(context.Background(), make([]byte, maxAudioBytes+1)); err == nil {
		t.Error("expected an error for oversized audio")
	}
	if called {
		t.Error("the API must not be called for audio that is rejected up front")
	}
}

func TestTranscribe_StopsWhenTheContextExpires(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if _, err := newTestTranscriber(server.URL).Transcribe(ctx, []byte("audio")); err == nil {
		t.Error("expected an error when the context expires")
	}
}