package middleware

/*
|--------------------------------------------------------------------------
| Session Store
|--------------------------------------------------------------------------
|
| Implements Session Store as part of the middleware package in the
| Gofreight framework. Key symbols: SessionStore.
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
| Symbols defined here include: SessionStore (exported type).
| 
*/

import "time"

// SessionStore persists session data (memory, Redis, etc.).
type SessionStore interface {
	Get(id string) (*Session, bool)
	Save(session *Session, maxAge time.Duration)
	Delete(id string)
}
