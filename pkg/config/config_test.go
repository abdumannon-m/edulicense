package config

import "testing"

func TestLoadDatabaseURLPrefersDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://primary")
	t.Setenv("POSTGRES_URL", "postgres://fallback")

	cfg := Load()

	if cfg.DatabaseURL != "postgres://primary" {
		t.Fatalf("DatabaseURL = %q, want primary DATABASE_URL", cfg.DatabaseURL)
	}
}

func TestLoadDatabaseURLFallsBackToPostgresURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("POSTGRES_URL", "postgres://fallback")

	cfg := Load()

	if cfg.DatabaseURL != "postgres://fallback" {
		t.Fatalf("DatabaseURL = %q, want fallback POSTGRES_URL", cfg.DatabaseURL)
	}
}
