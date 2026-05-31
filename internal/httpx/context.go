package httpx

import (
	"context"
	"net/http"

	"edu-license/internal/app"
)

type contextKey string

const userKey contextKey = "current_user"

func WithUser(r *http.Request, user app.User) *http.Request {
	ctx := context.WithValue(r.Context(), userKey, user)
	return r.WithContext(ctx)
}

func CurrentUser(r *http.Request) (app.User, bool) {
	user, ok := r.Context().Value(userKey).(app.User)
	return user, ok
}
