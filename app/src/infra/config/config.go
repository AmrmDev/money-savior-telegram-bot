package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	TelegramToken string
	TableName     string
	AllowedUserID int64
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

	allowedUserID, err := optionalUserID("ALLOWED_USER_ID")
	if err != nil {
		return Config{}, err
	}

	return Config{
		TelegramToken: token,
		TableName:     table,
		AllowedUserID: allowedUserID,
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

func optionalUserID(key string) (int64, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, nil
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("environment variable " + key + " must be a positive integer")
	}
	return id, nil
}

func defaultLogFormat() string {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		return "json"
	}
	return "text"
}