package middleware

/*
|--------------------------------------------------------------------------
| Session
|--------------------------------------------------------------------------
|
| Implements Session as part of the middleware package in the Gofreight
| framework. Key symbols: Session, NewSession, ID, Get, Set, Delete.
| 
| HTTP middleware implements sessions, CSRF, CORS, locale, structured
| logging, rate limiting, and maintenance mode.
| 
| Register global middleware in bootstrap or attach to route groups for
| API-specific stacks.
| 
| Session drivers include file, cookie, and Redis variants selected by
| SESSION_DRIVER.
| 
| Symbols defined here include: Session (exported type); NewSession
| (NewSession creates a new session with a random ID.); ID (ID returns the
| session identifier.); Get (Get retrieves a value from the session.); Set
| (Set stores a value in the session.); Delete (Delete removes a key from
| the session.); GetString (GetString retrieves a string value from the
| session.); Flash (Flash sets a one-time message stored under "_flash".).
| 
*/

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

// FlashValidationErrors stores validation errors for the next request.
func (s *Session) FlashValidationErrors(errs map[string][]string) {
	s.Set("_validation_errors", errs)
}

// PullValidationErrors returns and clears flashed validation errors.
func (s *Session) PullValidationErrors() map[string][]string {
	errs, _ := s.Get("_validation_errors").(map[string][]string)
	s.Delete("_validation_errors")
	if errs == nil {
		return make(map[string][]string)
	}
	return errs
}

// FlashOldInput stores submitted form values for the next request.
func (s *Session) FlashOldInput(data map[string]string) {
	s.Set("_old_input", data)
}

// PullOldInput returns and clears flashed old input.
func (s *Session) PullOldInput() map[string]string {
	old, _ := s.Get("_old_input").(map[string]string)
	s.Delete("_old_input")
	if old == nil {
		return make(map[string]string)
	}
	return old
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
	store      SessionStore
}

// memorySessionStore is the default in-memory session backend.
type memorySessionStore struct {
	mu    sync.RWMutex
	items map[string]*Session
}

func newMemorySessionStore() *memorySessionStore {
	return &memorySessionStore{items: make(map[string]*Session)}
}

func (m *memorySessionStore) Get(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.items[id]
	return s, ok
}

func (m *memorySessionStore) Save(session *Session, _ time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[session.id] = session
}

func (m *memorySessionStore) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, id)
}

// NewSessions creates session middleware with an in-memory store.
func NewSessions(secret string) *Sessions {
	return &Sessions{
		Secret:     secret,
		CookieName: "_gofreight_session",
		MaxAge:     24 * time.Hour,
		store:      newMemorySessionStore(),
	}
}

// UseStore sets a custom session store (e.g. Redis for production).
func (sm *Sessions) UseStore(store SessionStore) {
	sm.store = store
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
	if session, ok := sm.store.Get(cookie.Value); ok {
		return session
	}
	return NewSession()
}

func (sm *Sessions) save(w http.ResponseWriter, session *Session) {
	sm.store.Save(session, sm.MaxAge)

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
		if token == "" {
			token = r.FormValue("_csrf")
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
	return CSRFTokenFromSession(session)
}

// CSRFTokenFromSession returns or creates a CSRF token for the session.
func CSRFTokenFromSession(session *Session) string {
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
