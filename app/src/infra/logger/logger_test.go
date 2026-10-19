package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"  Debug ", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},
		{"banana", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := parseLevel(tt.in); got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestNew_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	log := newWithWriter(&buf, "info", "json")

	log.Info("update handled", "command", "gastei")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if entry["msg"] != "update handled" || entry["command"] != "gastei" || entry["level"] != "INFO" {
		t.Errorf("unexpected entry: %v", entry)
	}
}

func TestNew_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	log := newWithWriter(&buf, "info", "text")

	log.Info("update handled", "command", "gastei")

	out := buf.String()
	if strings.HasPrefix(out, "{") {
		t.Errorf("expected text output, got JSON: %s", out)
	}
	if !strings.Contains(out, "command=gastei") {
		t.Errorf("expected key=value output, got: %s", out)
	}
}

func TestNew_UnknownFormatFallsBackToText(t *testing.T) {
	var buf bytes.Buffer
	log := newWithWriter(&buf, "info", "xml")

	log.Info("hello")

	if strings.HasPrefix(buf.String(), "{") {
		t.Errorf("expected text fallback, got: %s", buf.String())
	}
}

func TestNew_RespectsLevel(t *testing.T) {
	var buf bytes.Buffer
	log := newWithWriter(&buf, "warn", "text")

	log.Info("should be dropped")
	log.Warn("should appear")

	out := buf.String()
	if strings.Contains(out, "should be dropped") {
		t.Errorf("info entry should be filtered at warn level: %s", out)
	}
	if !strings.Contains(out, "should appear") {
		t.Errorf("warn entry should be present: %s", out)
	}
}

func TestNew_RedactsSecrets(t *testing.T) {
	var buf bytes.Buffer
	log := newWithWriter(&buf, "info", "json", "123456:SECRET-TOKEN")

	log.Error("send failed with 123456:SECRET-TOKEN",
		"error", errors.New(`Post "https://api.telegram.org/bot123456:SECRET-TOKEN/sendMessage": timeout`),
		"detail", "token=123456:SECRET-TOKEN",
	)

	out := buf.String()
	if strings.Contains(out, "SECRET-TOKEN") {
		t.Errorf("secret leaked into the log: %s", out)
	}
	if !strings.Contains(out, redacted) {
		t.Errorf("expected the redaction marker in the log: %s", out)
	}
}

func TestNew_WithoutSecretsKeepsValues(t *testing.T) {
	var buf bytes.Buffer
	log := newWithWriter(&buf, "info", "json")

	log.Info("hello", "detail", "plain value")

	if !strings.Contains(buf.String(), "plain value") {
		t.Errorf("values should be untouched without secrets: %s", buf.String())
	}
}
