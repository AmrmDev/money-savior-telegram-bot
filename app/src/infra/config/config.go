package config

import (
	"fmt"
	"os"
)

type Config struct {
	TelegramToken string
	TableName     string
	LogLevel      string
	LogFormat     string
}

func Load() (Config, error) {
	token, err := required("TELEGRAM_BOT_TOKEN")
	if err != nil {
		return Config{}, err
	}

	table, err := required("TABLE_NAME")
	if err != nil {
		return Config{}, err
	}

	return Config{
		TelegramToken: token,
		TableName:     table,
		LogLevel:      optional("LOG_LEVEL", "info"),
		LogFormat:     optional("LOG_FORMAT", defaultLogFormat()),
	}, nil
}

func required(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not defined", key)
	}
	return value, nil
}

func optional(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func defaultLogFormat() string {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		return "json"
	}
	return "text"
}