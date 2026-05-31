package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"edu-license/pkg/app"
	"edu-license/pkg/auth"
	"edu-license/pkg/config"
	"edu-license/pkg/notify"
	"edu-license/pkg/storage"
	"edu-license/pkg/store"
	"edu-license/pkg/web"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	ctx := context.Background()
	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	cfg := config.Load()
	if err := cfg.ValidateServer(); err != nil {
		log.Fatal(err)
	}

	switch command {
	case "serve":
		server, cleanup := mustServer(ctx, cfg)
		defer cleanup()
		log.Printf("listening on %s", cfg.Addr)
		if err := http.ListenAndServe(cfg.Addr, server.Routes()); err != nil {
			log.Fatal(err)
		}
	case "migrate":
		if err := runMigrations(cfg.DatabaseURL); err != nil {
			log.Fatal(err)
		}
	case "reminders":
		server, cleanup := mustServer(ctx, cfg)
		defer cleanup()
		loc, err := time.LoadLocation(cfg.Timezone)
		if err != nil {
			log.Fatal(err)
		}
		count, err := server.SendDueReminders(ctx, time.Now().In(loc))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("sent %d reminder notifications", count)
	case "seed-admin":
		if err := seedAdmin(ctx, cfg); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown command %q; use serve, migrate, reminders, or seed-admin", command)
	}
}

func mustServer(ctx context.Context, cfg config.Config) (*web.Server, func()) {
	st, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	renderer, err := app.NewRenderer("pkg/templates/*.html")
	if err != nil {
		st.Close()
		log.Fatal(err)
	}
	authSvc := auth.NewService(st, cfg.SessionSecret, cfg.CookieSecure, cfg.SessionTTL)
	uploader, err := storage.NewS3Uploader(ctx, cfg)
	if err != nil {
		st.Close()
		log.Fatal(err)
	}
	telegram := notify.NewTelegram(cfg.TelegramBotToken, cfg.TelegramOperationsChatID, cfg.AppBaseURL)
	return web.New(cfg, st, authSvc, renderer, uploader, telegram), st.Close
}

func runMigrations(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, "migrations")
}

func seedAdmin(ctx context.Context, cfg config.Config) error {
	name := strings.TrimSpace(os.Getenv("SEED_ADMIN_NAME"))
	email := strings.TrimSpace(os.Getenv("SEED_ADMIN_EMAIL"))
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	if name == "" || email == "" || password == "" {
		return fmt.Errorf("SEED_ADMIN_NAME, SEED_ADMIN_EMAIL, and SEED_ADMIN_PASSWORD are required")
	}
	st, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	user, err := st.CreateUser(ctx, auth.CreateUserInput{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		Role:         app.RoleSuperAdmin,
		Active:       true,
	})
	if err != nil {
		return err
	}
	log.Printf("created super admin %s <%s>", user.Name, user.Email)
	return nil
}
