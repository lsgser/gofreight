package middleware

/*
|--------------------------------------------------------------------------
| Ratelimit
|--------------------------------------------------------------------------
|
| Implements Ratelimit as part of the middleware package in the Gofreight
| framework. Key symbols: RateLimiter, NewRateLimiter, Middleware.
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
| Symbols defined here include: RateLimiter (exported type);
| NewRateLimiter (NewRateLimiter creates a rate limiter (limit requests
| per window).); Middleware (Middleware returns rate limiting
| middleware.).
| 
*/

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter limits requests per IP address.
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a rate limiter (limit requests per window).
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Middleware returns rate limiting middleware.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = fwd
		}

		if !rl.allow(ip) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	var recent []time.Time
	for _, t := range rl.requests[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= rl.limit {
		rl.requests[key] = recent
		return false
	}
	rl.requests[key] = append(recent, now)
	return true
}
