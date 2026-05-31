package handler

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"edu-license/internal/app"
	"edu-license/internal/auth"
	"edu-license/internal/config"
	"edu-license/internal/notify"
	"edu-license/internal/storage"
	"edu-license/internal/store"
	apptemplates "edu-license/internal/templates"
	"edu-license/internal/web"
)

var (
	once       sync.Once
	appHandler http.Handler
	initErr    error
)

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initApp)
	if initErr != nil {
		http.Error(w, "admin runtime is not configured: "+initErr.Error(), http.StatusInternalServerError)
		return
	}
	if path := r.URL.Query().Get("path"); path != "" {
		r = rewritePath(r, path)
	}
	appHandler.ServeHTTP(w, r)
}

func initApp() {
	cfg := config.Load()
	if err := cfg.ValidateServer(); err != nil {
		initErr = err
		return
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		initErr = err
		return
	}
	renderer, err := app.NewRendererFS(apptemplates.FS, "*.html")
	if err != nil {
		initErr = err
		return
	}
	uploader, err := storage.NewS3Uploader(ctx, cfg)
	if err != nil {
		initErr = err
		return
	}
	authSvc := auth.NewService(st, cfg.SessionSecret, cfg.CookieSecure, cfg.SessionTTL)
	telegram := notify.NewTelegram(cfg.TelegramBotToken, cfg.TelegramOperationsChatID, cfg.AppBaseURL)
	appHandler = web.New(cfg, st, authSvc, renderer, uploader, telegram).Routes()
}

func rewritePath(r *http.Request, path string) *http.Request {
	copy := r.Clone(r.Context())
	copy.URL.Path = "/" + strings.TrimLeft(path, "/")
	query := copy.URL.Query()
	query.Del("path")
	copy.URL.RawQuery = query.Encode()
	return copy
}
