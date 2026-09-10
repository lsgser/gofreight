package auth

/*
|--------------------------------------------------------------------------
| Guard
|--------------------------------------------------------------------------
|
| Implements Guard as part of the auth package in the Gofreight framework.
| Key symbols: Guard, Middleware, OptionalMiddleware, UserIDFromRequest,
| RoleFromRequest, DefaultJWT.
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
| Symbols defined here include: Guard (exported type); Middleware
| (Middleware tries JWT, then opaque API tokens, then session
| authentication.); OptionalMiddleware (OptionalMiddleware attaches user
| context when credentials are present but allows anonymous access.);
| UserIDFromRequest (UserIDFromRequest returns the authenticated user ID
| from context, JWT, API token, or session.); RoleFromRequest
| (RoleFromRequest returns the authenticated user's role from context or
| session.); DefaultJWT (DefaultJWT creates a JWT manager from APP_KEY
| material with optional TTL.); JWTFromEnv (JWTFromEnv builds a JWT
| manager using APP_KEY and optional JWT_TTL.).
| 
*/

import (
	"net/http"
	"os"
	"time"

	"github.com/lsgser/gofreight/middleware"
)

// Guard authenticates requests via JWT, opaque API tokens, or session.
type Guard struct {
	JWT        *JWT
	TokenStore TokenStore
	SessionKey string
}

// Middleware tries JWT, then opaque API tokens, then session authentication.
func (g Guard) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token := bearerToken(r); token != "" {
			if g.JWT != nil {
				if claims, err := g.JWT.Validate(token); err == nil {
					ctx := WithUserID(r.Context(), claims.UserID)
					if claims.Role != "" {
						ctx = WithUserRole(ctx, claims.Role)
					}
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			if g.TokenStore != nil {
				if userID, ok := g.TokenStore.Validate(token); ok {
					next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
					return
				}
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if g.SessionKey != "" {
			if id, ok := CurrentUserID(r, g.SessionKey); ok {
				ctx := WithUserID(r.Context(), id)
				if session := middleware.SessionFromContext(r.Context()); session != nil {
					if role, ok := session.Get("current_user_role").(string); ok && role != "" {
						ctx = WithUserRole(ctx, role)
					}
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// OptionalMiddleware attaches user context when credentials are present but allows anonymous access.
func (g Guard) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token := bearerToken(r); token != "" {
			if g.JWT != nil {
				if claims, err := g.JWT.Validate(token); err == nil {
					ctx := WithUserID(r.Context(), claims.UserID)
					if claims.Role != "" {
						ctx = WithUserRole(ctx, claims.Role)
					}
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			if g.TokenStore != nil {
				if userID, ok := g.TokenStore.Validate(token); ok {
					next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
					return
				}
			}
		}
		if g.SessionKey != "" {
			if id, ok := CurrentUserID(r, g.SessionKey); ok {
				ctx := WithUserID(r.Context(), id)
				if session := middleware.SessionFromContext(r.Context()); session != nil {
					if role, ok := session.Get("current_user_role").(string); ok && role != "" {
						ctx = WithUserRole(ctx, role)
					}
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// UserIDFromRequest returns the authenticated user ID from context, JWT, API token, or session.
func UserIDFromRequest(r *http.Request, sessionKey string) (int64, bool) {
	if id, ok := UserIDFromContext(r.Context()); ok {
		return id, true
	}
	return CurrentUserID(r, sessionKey)
}

// RoleFromRequest returns the authenticated user's role from context or session.
func RoleFromRequest(r *http.Request) (string, bool) {
	if role, ok := RoleFromContext(r.Context()); ok {
		return role, true
	}
	session := middleware.SessionFromContext(r.Context())
	if session == nil {
		return "", false
	}
	role, ok := session.Get("current_user_role").(string)
	return role, ok && role != ""
}

// DefaultJWT creates a JWT manager from APP_KEY material with optional TTL.
func DefaultJWT(secret string, ttl time.Duration) *JWT {
	return NewJWT(JWTConfig{Secret: []byte(secret), TTL: ttl})
}

// JWTFromEnv builds a JWT manager using APP_KEY and optional JWT_TTL.
func JWTFromEnv(appKey string) *JWT {
	ttl := 24 * time.Hour
	if raw := os.Getenv("JWT_TTL"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			ttl = d
		}
	}
	return DefaultJWT(appKey, ttl)
}
