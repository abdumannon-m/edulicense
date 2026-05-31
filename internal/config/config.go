package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Addr                     string
	AppBaseURL               string
	DatabaseURL              string
	SessionSecret            string
	CookieSecure             bool
	S3Endpoint               string
	S3Bucket                 string
	S3AccessKeyID            string
	S3SecretAccessKey        string
	S3Region                 string
	TelegramBotToken         string
	TelegramOperationsChatID string
	Timezone                 string
	SessionTTL               time.Duration
}

func Load() Config {
	cfg := Config{
		Addr:          env("ADDR", ":8080"),
		AppBaseURL:    strings.TrimRight(env("APP_BASE_URL", "http://localhost:8080"), "/"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		SessionSecret: env("SESSION_SECRET", "dev-only-change-me"),
		CookieSecure:  env("COOKIE_SECURE", "false") == "true",
		S3Endpoint:    os.Getenv("S3_ENDPOINT"),
		S3Bucket:      os.Getenv("S3_BUCKET"),
		S3AccessKeyID: os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv(
			"S3_SECRET_ACCESS_KEY",
		),
		S3Region:                 env("S3_REGION", "auto"),
		TelegramBotToken:         os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramOperationsChatID: os.Getenv("TELEGRAM_OPERATIONS_CHAT_ID"),
		Timezone:                 env("APP_TIMEZONE", "Asia/Tashkent"),
		SessionTTL:               14 * 24 * time.Hour,
	}
	return cfg
}

func (c Config) ValidateServer() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.SessionSecret == "" || c.SessionSecret == "dev-only-change-me" {
		return fmt.Errorf("SESSION_SECRET must be set to a strong random value")
	}
	return nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
