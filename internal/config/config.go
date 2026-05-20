package config

import (
	"errors"
	"os"
)

type Config struct {
	TelegramToken string
	DatabaseURL   string
}

func Load() (*Config, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, errors.New("TELEGRAM_TOKEN is required")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	return &Config{TelegramToken: token, DatabaseURL: dbURL}, nil
}
