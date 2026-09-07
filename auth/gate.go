package auth

import (
	"context"
	"net/http"
)

// Authorizer checks whether a user can perform an action on a model.
type Authorizer interface {
	Authorize(ctx context.Context, userID int64, action string, model any) bool
}

// AuthorizeFunc is a function-based authorizer.
type AuthorizeFunc func(ctx context.Context, userID int64, action string, model any) bool

func (fn AuthorizeFunc) Authorize(ctx context.Context, userID int64, action string, model any) bool {
	return fn(ctx, userID, action, model)
}

// Gate provides Laravel-style authorization gates.
type Gate struct {
	rules map[string]AuthorizeFunc
}

// NewGate creates an authorization gate registry.
func NewGate() *Gate {
	return &Gate{rules: make(map[string]AuthorizeFunc)}
}

// Define registers an authorization rule.
func (g *Gate) Define(action string, fn AuthorizeFunc) {
	g.rules[action] = fn
}

// Allows checks authorization for the current request user.
func (g *Gate) Allows(r *http.Request, sessionKey, action string, model any) bool {
	userID, ok := UserIDFromRequest(r, sessionKey)
	if !ok {
		return false
	}
	fn, ok := g.rules[action]
	if !ok {
		return false
	}
	return fn(r.Context(), userID, action, model)
}

// RequireGate middleware denies unauthorized requests.
func (g *Gate) RequireGate(action string, sessionKey string, model any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !g.Allows(r, sessionKey, action, model) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
