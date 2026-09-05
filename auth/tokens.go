package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/lsgser/gofreight/controller"
)

// TokenStore persists API tokens.
type TokenStore interface {
	Create(userID int64, name string, expiresAt time.Time) (string, error)
	Validate(token string) (userID int64, ok bool)
	Revoke(token string) error
}

// MemoryTokenStore is an in-memory token store (use database in production).
type MemoryTokenStore struct {
	tokens map[string]tokenEntry
}

type tokenEntry struct {
	userID    int64
	expiresAt time.Time
}

func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{tokens: make(map[string]tokenEntry)}
}

func (s *MemoryTokenStore) Create(userID int64, name string, expiresAt time.Time) (string, error) {
	_ = name
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	s.tokens[token] = tokenEntry{userID: userID, expiresAt: expiresAt}
	return token, nil
}

func (s *MemoryTokenStore) Validate(token string) (int64, bool) {
	e, ok := s.tokens[token]
	if !ok || time.Now().After(e.expiresAt) {
		return 0, false
	}
	return e.userID, true
}

func (s *MemoryTokenStore) Revoke(token string) error {
	delete(s.tokens, token)
	return nil
}

// APITokenMiddleware authenticates requests via Bearer opaque API token.
func APITokenMiddleware(store TokenStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			userID, ok := store.Validate(token)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type ctxUserID struct{}

func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, ctxUserID{}, id)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(ctxUserID{}).(int64)
	return id, ok
}

// PasswordResetStore stores password reset tokens.
type PasswordResetStore interface {
	Create(email string, expiresAt time.Time) (string, error)
	Consume(token string) (email string, ok bool)
}

type resetEntry struct {
	email     string
	expiresAt time.Time
}

// MemoryPasswordResetStore is an in-memory reset token store.
type MemoryPasswordResetStore struct {
	tokens map[string]resetEntry
}

func NewMemoryPasswordResetStore() *MemoryPasswordResetStore {
	return &MemoryPasswordResetStore{tokens: make(map[string]resetEntry)}
}

func (s *MemoryPasswordResetStore) Create(email string, expiresAt time.Time) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	s.tokens[token] = resetEntry{email: email, expiresAt: expiresAt}
	return token, nil
}

func (s *MemoryPasswordResetStore) Consume(token string) (string, bool) {
	e, ok := s.tokens[token]
	if !ok || time.Now().After(e.expiresAt) {
		return "", false
	}
	delete(s.tokens, token)
	return e.email, true
}

// RequestPasswordReset handles POST /password/forgot.
func RequestPasswordReset(store PasswordResetStore, findUser func(email string) (*User, error), sendMail func(email, token string) error) func(controller.Base) error {
	return func(base controller.Base) error {
		if err := base.Request.ParseForm(); err != nil {
			return err
		}
		email := base.Request.FormValue("email")
		user, err := findUser(email)
		if err != nil || user == nil {
			base.RenderJSON(map[string]string{"message": "If that email exists, a reset link was sent."})
			return nil
		}
		token, err := store.Create(email, time.Now().Add(time.Hour))
		if err != nil {
			return err
		}
		if sendMail != nil {
			_ = sendMail(email, token)
		}
		base.RenderJSON(map[string]string{"message": "If that email exists, a reset link was sent."})
		return nil
	}
}

// ResetPassword handles POST /password/reset.
func ResetPassword(store PasswordResetStore, updatePassword func(email, hash string) error) func(controller.Base) error {
	return func(base controller.Base) error {
		if err := base.Request.ParseForm(); err != nil {
			return err
		}
		token := base.Request.FormValue("token")
		password := base.Request.FormValue("password")
		email, ok := store.Consume(token)
		if !ok || password == "" {
			base.Unprocessable(map[string][]string{"token": {"invalid or expired"}})
			return nil
		}
		hash, err := HashPassword(password)
		if err != nil {
			return err
		}
		if err := updatePassword(email, hash); err != nil {
			return err
		}
		base.RenderJSON(map[string]string{"message": "Password updated."})
		return nil
	}
}
