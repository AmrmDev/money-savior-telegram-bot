package config

import "testing"

func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	t.Setenv("TABLE_NAME", "test-table")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("AWS_LAMBDA_FUNCTION_NAME", "")
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
