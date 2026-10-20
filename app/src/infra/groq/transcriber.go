package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"money-savior-telegram-bot/app/src/core/port"
)

const (
	defaultBaseURL        = "https://api.groq.com/openai/v1"
	transcriptionModel    = "whisper-large-v3-turbo"
	transcriptionLanguage = "pt"
	transcriptionTimeout  = 20 * time.Second

	// The free plan accepts up to 25 MB; a voice message is far smaller than this.
	maxAudioBytes    = 20 << 20
	maxResponseBytes = 1 << 20
)

var _ port.Transcriber = (*Transcriber)(nil)

// Transcriber turns voice messages into text with Groq's Whisper.
type Transcriber struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewTranscriber(apiKey string) *Transcriber {
	return newTranscriber(apiKey, defaultBaseURL, &http.Client{Timeout: transcriptionTimeout})
}

func newTranscriber(apiKey, baseURL string, httpClient *http.Client) *Transcriber {
	return &Transcriber{apiKey: apiKey, baseURL: baseURL, httpClient: httpClient}
}

// Transcribe sends the audio (OGG/Opus, as Telegram records it) and returns the text.
// Errors never carry the API key or the response body.
func (t *Transcriber) Transcribe(ctx context.Context, audio []byte) (string, error) {
	if len(audio) == 0 {
		return "", errors.New("groq transcription: empty audio")
	}
	if len(audio) > maxAudioBytes {
		return "", errors.New("groq transcription: audio too large")
	}

	body, contentType, err := transcriptionForm(audio)
	if err != nil {
		return "", fmt.Errorf("groq transcription: build request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("groq transcription: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", contentType)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("groq transcription: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseBytes))
		return "", fmt.Errorf("groq transcription: unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&result); err != nil {
		return "", fmt.Errorf("groq transcription: decode response: %w", err)
	}

	return strings.TrimSpace(result.Text), nil
}

func transcriptionForm(audio []byte) (*bytes.Buffer, string, error) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)

	fields := map[string]string{
		"model":           transcriptionModel,
		"language":        transcriptionLanguage,
		"response_format": "json",
	}
	for name, value := range fields {
		if err := form.WriteField(name, value); err != nil {
			return nil, "", err
		}
	}

	file, err := form.CreateFormFile("file", "voice.ogg")
	if err != nil {
		return nil, "", err
	}
	if _, err := file.Write(audio); err != nil {
		return nil, "", err
	}

	if err := form.Close(); err != nil {
		return nil, "", err
	}
	return &body, form.FormDataContentType(), nil
}