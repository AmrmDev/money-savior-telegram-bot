package config

import (
	"strings"
	"testing"
)

func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	t.Setenv("TABLE_NAME", "test-table")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("AWS_LAMBDA_FUNCTION_NAME", "")
	t.Setenv("ALLOWED_USER_ID", "")
}

func TestLoad_LogDefaultsOnLocal(t *testing.T) {
	setBaseEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogLevel != "info" || cfg.LogFormat != "text" {
		t.Errorf("unexpected defaults: level=%q format=%q", cfg.LogLevel, cfg.LogFormat)
	}
}

func TestLoad_LogFormatDefaultsToJSONOnLambda(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("AWS_LAMBDA_FUNCTION_NAME", "money-savior-webhook")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("expected json on lambda, got %q", cfg.LogFormat)
	}
}

func TestLoad_LogEnvOverridesDefaults(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("AWS_LAMBDA_FUNCTION_NAME", "money-savior-webhook")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "text")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogLevel != "debug" || cfg.LogFormat != "text" {
		t.Errorf("env should win: level=%q format=%q", cfg.LogLevel, cfg.LogFormat)
	}
}

func TestLoad_RequiresToken(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("TELEGRAM_BOT_TOKEN", "")

	if _, err := Load(); err == nil {
		t.Error("expected an error when the token is missing")
	}
}

func TestLoad_RequiresTableName(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("TABLE_NAME", "")

	if _, err := Load(); err == nil {
		t.Error("expected an error when the table name is missing")
	}
}

func TestLoad_AllowedUserIDDefaultsToZero(t *testing.T) {
	setBaseEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AllowedUserID != 0 {
		t.Errorf("expected 0 when unset, got %d", cfg.AllowedUserID)
	}
}

func TestLoad_AllowedUserIDIsParsed(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("ALLOWED_USER_ID", " 123456789 ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AllowedUserID != 123456789 {
		t.Errorf("unexpected id: %d", cfg.AllowedUserID)
	}
}

func TestLoad_AllowedUserIDRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"abc", "12.5", "-7", "0", "123abc"} {
		t.Run(value, func(t *testing.T) {
			setBaseEnv(t)
			t.Setenv("ALLOWED_USER_ID", value)

			_, err := Load()
			if err == nil {
				t.Fatal("expected an error for an invalid ALLOWED_USER_ID")
			}
			if strings.Contains(err.Error(), value) {
				t.Errorf("error should not echo the value: %v", err)
			}
		})
	}
}
