package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"edu-license/pkg/app"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookie = "edu_session"
	CSRFCookie    = "edu_csrf"
)

type Store interface {
	UserByEmail(ctx context.Context, email string) (StoredUser, error)
	UserBySessionTokenHash(ctx context.Context, hash string, now time.Time) (app.User, error)
	CreateSession(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, tokenHash string) error
	CreateUser(ctx context.Context, input CreateUserInput) (app.User, error)
}

type StoredUser struct {
	app.User
	PasswordHash string
}

type CreateUserInput struct {
	Name         string
	Email        string
	PasswordHash string
	Role         app.Role
	Active       bool
}

type Service struct {
	store        Store
	sessionKey   []byte
	csrfKey      []byte
	cookieSecure bool
	sessionTTL   time.Duration
}

func NewService(store Store, secret string, cookieSecure bool, sessionTTL time.Duration) *Service {
	key := sha256.Sum256([]byte(secret))
	csrf := sha256.Sum256([]byte("csrf:" + secret))
	return &Service{
		store:        store,
		sessionKey:   key[:],
		csrfKey:      csrf[:],
		cookieSecure: cookieSecure,
		sessionTTL:   sessionTTL,
	}
}

func (s *Service) Login(ctx context.Context, email, password string) (string, time.Time, error) {
	user, err := s.store.UserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("invalid email or password")
	}
	if !user.Active {
		return "", time.Time{}, fmt.Errorf("account is inactive")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", time.Time{}, fmt.Errorf("invalid email or password")
	}
	token, err := randomToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(s.sessionTTL)
	if err := s.store.CreateSession(ctx, user.ID, tokenHash(token), expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return s.sign(token, s.sessionKey), expiresAt, nil
}

func (s *Service) Logout(ctx context.Context, signedToken string) {
	token, ok := s.unsign(signedToken, s.sessionKey)
	if !ok {
		return
	}
	_ = s.store.DeleteSession(ctx, tokenHash(token))
}

func (s *Service) UserFromRequest(r *http.Request) (app.User, bool) {
	cookie, err := r.Cookie(SessionCookie)
	if err != nil {
		return app.User{}, false
	}
	token, ok := s.unsign(cookie.Value, s.sessionKey)
	if !ok {
		return app.User{}, false
	}
	user, err := s.store.UserBySessionTokenHash(r.Context(), tokenHash(token), time.Now())
	if err != nil {
		return app.User{}, false
	}
	return user, true
}

func (s *Service) SetSessionCookie(w http.ResponseWriter, value string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    value,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Service) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Service) EnsureCSRF(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(CSRFCookie); err == nil {
		if token, ok := s.unsign(cookie.Value, s.csrfKey); ok {
			return token
		}
	}
	token, err := randomToken()
	if err != nil {
		token = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookie,
		Value:    s.sign(token, s.csrfKey),
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	return token
}

func (s *Service) ValidateCSRF(r *http.Request) bool {
	cookie, err := r.Cookie(CSRFCookie)
	if err != nil {
		return false
	}
	expected, ok := s.unsign(cookie.Value, s.csrfKey)
	if !ok {
		return false
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil && err != http.ErrNotMultipart {
		return false
	}
	actual := r.FormValue("csrf_token")
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *Service) sign(value string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(value))
	sig := mac.Sum(nil)
	return value + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func (s *Service) unsign(signed string, key []byte) (string, bool) {
	value, sig, ok := strings.Cut(signed, ".")
	if !ok {
		return "", false
	}
	rawSig, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(value))
	expected := mac.Sum(nil)
	if !hmac.Equal(rawSig, expected) {
		return "", false
	}
	return value, true
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
