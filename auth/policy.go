package auth

/*
|--------------------------------------------------------------------------
| Policy
|--------------------------------------------------------------------------
|
| Implements Policy as part of the auth package in the Gofreight
| framework. Key symbols: Policy, NewPolicy, Define, Allows,
| RequirePolicy, RequireRole.
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
| Symbols defined here include: Policy (exported type); NewPolicy
| (NewPolicy creates an empty policy registry.); Define (Define registers
| an authorization rule by name.); Allows (Allows checks if the current
| request passes a policy.); RequirePolicy (RequirePolicy middleware
| denies requests that fail a policy check.); RequireRole (RequireRole
| middleware allows only users with a given role (JWT, session, or API
| context).).
| 
*/

import "net/http"

// Policy defines named authorization rules.
type Policy struct {
	rules map[string]func(r *http.Request) bool
}

// NewPolicy creates an empty policy registry.
func NewPolicy() *Policy {
	return &Policy{rules: make(map[string]func(r *http.Request) bool)}
}

// Define registers an authorization rule by name.
func (p *Policy) Define(name string, fn func(r *http.Request) bool) {
	p.rules[name] = fn
}

// Allows checks if the current request passes a policy.
func (p *Policy) Allows(name string, r *http.Request) bool {
	fn, ok := p.rules[name]
	if !ok {
		return false
	}
	return fn(r)
}

// RequirePolicy middleware denies requests that fail a policy check.
func (p *Policy) RequirePolicy(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !p.Allows(name, r) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole middleware allows only users with a given role (JWT, session, or API context).
func RequireRole(role, sessionKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := RoleFromRequest(r)
			if !ok {
				if _, ok := UserIDFromRequest(r, sessionKey); !ok {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			if userRole != role && userRole != "admin" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
