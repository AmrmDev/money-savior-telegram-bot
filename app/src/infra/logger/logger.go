package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

const redacted = "[REDACTED]"

// New builds the application logger. Every value passed as secret is masked wherever it
// shows up in a log line, including inside error messages and URLs.
func New(level, format string, secrets ...string) *slog.Logger {
	return newWithWriter(os.Stdout, level, format, secrets...)
}

func newWithWriter(w io.Writer, level, format string, secrets ...string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:       parseLevel(level),
		ReplaceAttr: redactor(secrets),
	}

	if strings.EqualFold(strings.TrimSpace(format), "json") {
		return slog.New(slog.NewJSONHandler(w, opts))
	}
	return slog.New(slog.NewTextHandler(w, opts))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func redactor(secrets []string) func([]string, slog.Attr) slog.Attr {
	var active []string
	for _, secret := range secrets {
		if secret != "" {
			active = append(active, secret)
		}
	}
	if len(active) == 0 {
		return nil
	}

	scrub := func(text string) string {
		for _, secret := range active {
			text = strings.ReplaceAll(text, secret, redacted)
		}
		return text
	}

	return func(_ []string, attr slog.Attr) slog.Attr {
		switch attr.Value.Kind() {
		case slog.KindString:
			attr.Value = slog.StringValue(scrub(attr.Value.String()))
		case slog.KindAny:
			if err, ok := attr.Value.Any().(error); ok {
				attr.Value = slog.StringValue(scrub(err.Error()))
			}
		}
		return attr
	}
}
