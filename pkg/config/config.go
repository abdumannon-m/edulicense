package config

import (
	"fmt"
	"net/url"
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
	// DevAutoLogin signs visitors in as DEV_LOGIN_EMAIL without a password.
	// It only takes effect when APP_BASE_URL points at localhost; see LocalDevLogin.
	DevAutoLogin bool
}

func Load() Config {
	cfg := Config{
		Addr:          env("ADDR", ":8080"),
		AppBaseURL:    strings.TrimRight(env("APP_BASE_URL", "http://localhost:8080"), "/"),
		DatabaseURL:   firstEnv("DATABASE_URL", "POSTGRES_URL"),
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
		DevAutoLogin:             os.Getenv("DEV_AUTO_LOGIN") == "1",
	}
	return cfg
}

func (c Config) ValidateServer() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL or POSTGRES_URL is required")
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

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

// LocalDevLogin reports whether the login page should sign in automatically:
// DEV_AUTO_LOGIN=1 and an APP_BASE_URL on localhost, so a production deploy
// can never skip the password even if the variable leaks into its env.
func (c Config) LocalDevLogin() bool {
	if !c.DevAutoLogin {
		return false
	}
	base, err := url.Parse(c.AppBaseURL)
	if err != nil {
		return false
	}
	host := base.Hostname()
	return host == "localhost" || host == "127.0.0.1"
}
