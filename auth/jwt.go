package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims are signed into access tokens.
type JWTClaims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// JWT issues and validates HS256 bearer tokens signed with APP_KEY.
type JWT struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

// JWTConfig configures token signing.
type JWTConfig struct {
	Secret []byte
	TTL    time.Duration
	Issuer string
}

// NewJWT creates a JWT manager. TTL defaults to 24 hours.
func NewJWT(cfg JWTConfig) *JWT {
	ttl := cfg.TTL
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &JWT{
		secret: cfg.Secret,
		ttl:    ttl,
		issuer: cfg.Issuer,
	}
}

// Issue creates a signed access token for a user.
func (j *JWT) Issue(userID int64, role string) (string, time.Time, error) {
	if len(j.secret) == 0 {
		return "", time.Time{}, errors.New("jwt secret is required")
	}
	expiresAt := time.Now().Add(j.ttl)
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

// Validate parses and verifies a JWT access token.
func (j *JWT) Validate(tokenString string) (JWTClaims, error) {
	var claims JWTClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return JWTClaims{}, err
	}
	if !token.Valid {
		return JWTClaims{}, errors.New("invalid token")
	}
	return claims, nil
}

// JWTMiddleware authenticates requests with a Bearer JWT.
func JWTMiddleware(jwtMgr *JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			claims, err := jwtMgr.Validate(token)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := WithUserID(r.Context(), claims.UserID)
			if claims.Role != "" {
				ctx = WithUserRole(ctx, claims.Role)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type ctxUserRole struct{}

// WithUserRole stores a role on the request context.
func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, ctxUserRole{}, role)
}

// RoleFromContext returns the authenticated user's role from JWT/API context.
func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(ctxUserRole{}).(string)
	return role, ok && role != ""
}

// bearerToken reads a Bearer token from Authorization or api_token query param.
func bearerToken(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(header) > 7 && strings.EqualFold(header[:7], "Bearer ") {
		return strings.TrimSpace(header[7:])
	}
	return r.URL.Query().Get("api_token")
}
