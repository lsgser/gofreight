package auth

/*
|--------------------------------------------------------------------------
| Verification
|--------------------------------------------------------------------------
|
| Implements Verification as part of the auth package in the Gofreight
| framework. Key symbols: VerificationStore, MemoryVerificationStore,
| NewMemoryVerificationStore, Create, Consume, SendVerificationEmail.
| 
| The auth package covers session login, password hashing, API token
| storage, OAuth callbacks, email verification, and password reset flows.
| 
| Controllers compose auth helpers with your User model; tokens and
| verification stores can be in-memory or database-backed.
| 
| Install scaffolding with gofreight make:auth and wire find-user
| callbacks in app/auth.
| 
| Symbols defined here include: VerificationStore (exported type);
| MemoryVerificationStore (exported type); SendVerificationEmail
| (SendVerificationEmail creates a token and invokes the mail callback.);
| VerifyEmailHandler (VerifyEmailHandler handles GET
| /email/verify?token=...); ResendVerificationHandler
| (ResendVerificationHandler handles POST /email/verification/resend.).
| 
*/

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/lsgser/gofreight/controller"
)

// VerificationStore persists email verification tokens.
type VerificationStore interface {
	Create(userID int64, email string, expiresAt time.Time) (string, error)
	Consume(token string) (userID int64, email string, ok bool)
}

type verifyEntry struct {
	userID    int64
	email     string
	expiresAt time.Time
}

// MemoryVerificationStore is an in-memory verification token store.
type MemoryVerificationStore struct {
	tokens map[string]verifyEntry
}

func NewMemoryVerificationStore() *MemoryVerificationStore {
	return &MemoryVerificationStore{tokens: make(map[string]verifyEntry)}
}

func (s *MemoryVerificationStore) Create(userID int64, email string, expiresAt time.Time) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	s.tokens[token] = verifyEntry{userID: userID, email: email, expiresAt: expiresAt}
	return token, nil
}

func (s *MemoryVerificationStore) Consume(token string) (int64, string, bool) {
	e, ok := s.tokens[token]
	if !ok || time.Now().After(e.expiresAt) {
		return 0, "", false
	}
	delete(s.tokens, token)
	return e.userID, e.email, true
}

// SendVerificationEmail creates a token and invokes the mail callback.
func SendVerificationEmail(store VerificationStore, userID int64, email string, send func(email, token string) error) (string, error) {
	token, err := store.Create(userID, email, time.Now().Add(24*time.Hour))
	if err != nil {
		return "", err
	}
	if send != nil {
		_ = send(email, token)
	}
	return token, nil
}

// VerifyEmailHandler handles GET /email/verify?token=...
func VerifyEmailHandler(store VerificationStore, markVerified func(userID int64, email string) error) func(controller.Base) error {
	return func(base controller.Base) error {
		token := base.Query("token")
		userID, email, ok := store.Consume(token)
		if !ok {
			base.Unprocessable(map[string][]string{"token": {"invalid or expired"}})
			return nil
		}
		if markVerified != nil {
			if err := markVerified(userID, email); err != nil {
				return err
			}
		}
		base.RenderJSON(map[string]any{"message": "Email verified.", "user_id": userID})
		return nil
	}
}

// ResendVerificationHandler handles POST /email/verification/resend.
func ResendVerificationHandler(store VerificationStore, findUser func(email string) (int64, error), send func(email, token string) error) func(controller.Base) error {
	return func(base controller.Base) error {
		if err := base.Request.ParseForm(); err != nil {
			return err
		}
		email := base.Request.FormValue("email")
		userID, err := findUser(email)
		if err != nil {
			base.RenderJSON(map[string]string{"message": "If that email exists, a verification link was sent."})
			return nil
		}
		_, _ = SendVerificationEmail(store, userID, email, send)
		base.RenderJSON(map[string]string{"message": "If that email exists, a verification link was sent."})
		return nil
	}
}
