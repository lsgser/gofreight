package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type sessionKey struct{}

// Session stores request-scoped session data.
type Session struct {
	data map[string]any
	mu   sync.RWMutex
	id   string
}

// NewSession creates a new session with a random ID.
func NewSession() *Session {
	return &Session{
		data: make(map[string]any),
		id:   generateToken(32),
	}
}

// ID returns the session identifier.
func (s *Session) ID() string { return s.id }

// Get retrieves a value from the session.
func (s *Session) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

// Set stores a value in the session.
func (s *Session) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Delete removes a key from the session.
func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// GetString retrieves a string value from the session.
func (s *Session) GetString(key string) string {
	v, _ := s.Get(key).(string)
	return v
}

// Flash sets a one-time message stored under "_flash".
func (s *Session) Flash(key, message string) {
	flashes, _ := s.Get("_flash").(map[string]string)
	if flashes == nil {
		flashes = make(map[string]string)
	}
	flashes[key] = message
	s.Set("_flash", flashes)
}

// Flashes returns and clears flash messages.
func (s *Session) Flashes() map[string]string {
	flashes, _ := s.Get("_flash").(map[string]string)
	s.Delete("_flash")
	if flashes == nil {
		return make(map[string]string)
	}
	return flashes
}

// SessionFromContext retrieves the session from the request context.
func SessionFromContext(ctx context.Context) *Session {
	s, _ := ctx.Value(sessionKey{}).(*Session)
	return s
}

// Sessions provides cookie-based session middleware.
type Sessions struct {
	Secret     string
	CookieName string
	MaxAge     time.Duration
	store      map[string]*Session
	mu         sync.RWMutex
}

// NewSessions creates session middleware with an in-memory store.
func NewSessions(secret string) *Sessions {
	return &Sessions{
		Secret:     secret,
		CookieName: "_gofreight_session",
		MaxAge:     24 * time.Hour,
		store:      make(map[string]*Session),
	}
}

// Middleware loads or creates a session for each request.
func (sm *Sessions) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := sm.load(r)
		ctx := context.WithValue(r.Context(), sessionKey{}, session)
		next.ServeHTTP(w, r.WithContext(ctx))
		sm.save(w, session)
	})
}

func (sm *Sessions) load(r *http.Request) *Session {
	cookie, err := r.Cookie(sm.CookieName)
	if err != nil {
		return NewSession()
	}

	sm.mu.RLock()
	session, ok := sm.store[cookie.Value]
	sm.mu.RUnlock()

	if !ok {
		return NewSession()
	}
	return session
}

func (sm *Sessions) save(w http.ResponseWriter, session *Session) {
	sm.mu.Lock()
	sm.store[session.id] = session
	sm.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sm.CookieName,
		Value:    session.id,
		Path:     "/",
		MaxAge:   int(sm.MaxAge.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// CSRF provides cross-site request forgery protection.
type CSRF struct {
	HeaderName string
	FieldName  string
	Excluded   map[string]bool
}

// NewCSRF creates CSRF middleware. Safe methods (GET, HEAD, OPTIONS) are excluded.
func NewCSRF() *CSRF {
	return &CSRF{
		HeaderName: "X-CSRF-Token",
		FieldName:  "authenticity_token",
		Excluded: map[string]bool{
			"GET": true, "HEAD": true, "OPTIONS": true,
		},
	}
}

// Middleware validates CSRF tokens on mutating requests.
func (c *CSRF) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.Excluded[r.Method] {
			next.ServeHTTP(w, r)
			return
		}

		session := SessionFromContext(r.Context())
		if session == nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		expected, _ := session.Get("_csrf_token").(string)
		if expected == "" {
			expected = generateToken(32)
			session.Set("_csrf_token", expected)
		}

		token := r.Header.Get(c.HeaderName)
		if token == "" {
			token = r.FormValue(c.FieldName)
		}

		if token != expected {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Token returns the CSRF token for the current session.
func (c *CSRF) Token(r *http.Request) string {
	session := SessionFromContext(r.Context())
	if session == nil {
		return ""
	}
	token, _ := session.Get("_csrf_token").(string)
	if token == "" {
		token = generateToken(32)
		session.Set("_csrf_token", token)
	}
	return token
}

// Authenticator provides basic authentication middleware.
type Authenticator struct {
	// Validate checks credentials and returns the user ID (or nil if invalid).
	Validate func(username, password string) (userID any, ok bool)
	// CurrentUserKey is the session key for storing the authenticated user.
	CurrentUserKey string
}

// NewAuthenticator creates authentication middleware.
func NewAuthenticator(validate func(string, string) (any, bool)) *Authenticator {
	return &Authenticator{
		Validate:       validate,
		CurrentUserKey: "current_user_id",
	}
}

// Middleware requires HTTP Basic Auth credentials.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Gofreight"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, valid := a.Validate(user, pass)
		if !valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		session := SessionFromContext(r.Context())
		if session != nil {
			session.Set(a.CurrentUserKey, userID)
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAuth ensures a user is logged in via session.
func (a *Authenticator) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := SessionFromContext(r.Context())
		if session == nil || session.Get(a.CurrentUserKey) == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CurrentUser returns the authenticated user ID from the session.
func (a *Authenticator) CurrentUser(r *http.Request) any {
	session := SessionFromContext(r.Context())
	if session == nil {
		return nil
	}
	return session.Get(a.CurrentUserKey)
}

func generateToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// SessionJSON serializes session data (for debugging).
func (s *Session) JSON() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, _ := json.Marshal(s.data)
	return string(b)
}
