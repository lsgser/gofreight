package middleware

import "time"

// SessionStore persists session data (memory, Redis, etc.).
type SessionStore interface {
	Get(id string) (*Session, bool)
	Save(session *Session, maxAge time.Duration)
	Delete(id string)
}
