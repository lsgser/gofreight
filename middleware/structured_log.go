package middleware

/*
|--------------------------------------------------------------------------
| Structured Log
|--------------------------------------------------------------------------
|
| Implements Structured Log as part of the middleware package in the
| Gofreight framework. Key symbols: StructuredLogger, RequestID,
| JSONErrorLogger.
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
| Symbols defined here include: StructuredLogger (StructuredLogger logs
| requests as JSON via slog.); RequestID (RequestID adds a unique request
| ID header.); JSONErrorLogger (JSONLogger returns a middleware that logs
| errors as JSON.).
| 
*/

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// StructuredLogger logs requests as JSON via slog.
func StructuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"ip", r.RemoteAddr,
		)
	})
}

// RequestID adds a unique request ID header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = generateToken(16)
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

// JSONLogger returns a middleware that logs errors as JSON.
func JSONErrorLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic", "error", err, "path", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
