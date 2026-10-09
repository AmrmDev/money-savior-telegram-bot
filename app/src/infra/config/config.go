package config

import (
	"fmt"
	"os"
)

type Config struct {
	TelegramToken string
	TableName     string
}

func Load() (Config, error) {
	token, err := required("TELEGRAM_BOT_TOKEN")
	if err != nil  {
		return Config{}, err
	}

	table, err := required("TABLE_NAME")
	if err != nil {
		return Config{}, err
	}

	return Config{
		TelegramToken: token,
		TableName: table,
	}, nil
}

func required(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("variável de ambiente %s não definida", key)
	}
	return value, nil
}
