package auth_test

/*
|--------------------------------------------------------------------------
| Jwt
|--------------------------------------------------------------------------
|
| Test suite for Jwt in the auth package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
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
| Run with go test ./auth/... or go test for this package from the
| framework root.
| 
*/

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lsgser/gofreight/auth"
)

func TestJWTIssueAndValidate(t *testing.T) {
	jwtMgr := auth.NewJWT(auth.JWTConfig{
		Secret: []byte("test-secret-key"),
		TTL:    time.Hour,
		Issuer: "gofreight",
	})

	token, _, err := jwtMgr.Issue(42, "editor")
	if err != nil {
		t.Fatal(err)
	}

	claims, err := jwtMgr.Validate(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 {
		t.Fatalf("uid: got %d", claims.UserID)
	}
	if claims.Role != "editor" {
		t.Fatalf("role: got %q", claims.Role)
	}
}

func TestJWTRejectsTamperedToken(t *testing.T) {
	jwtMgr := auth.NewJWT(auth.JWTConfig{Secret: []byte("secret-one"), TTL: time.Hour})
	other := auth.NewJWT(auth.JWTConfig{Secret: []byte("secret-two"), TTL: time.Hour})

	token, _, err := jwtMgr.Issue(1, "user")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.Validate(token); err == nil {
		t.Fatal("expected invalid signature")
	}
}

func TestJWTMiddleware(t *testing.T) {
	jwtMgr := auth.NewJWT(auth.JWTConfig{Secret: []byte("jwt-secret"), TTL: time.Hour})
	token, _, err := jwtMgr.Issue(7, "admin")
	if err != nil {
		t.Fatal(err)
	}

	called := false
	handler := auth.JWTMiddleware(jwtMgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		id, ok := auth.UserIDFromContext(r.Context())
		if !ok || id != 7 {
			t.Fatalf("expected user 7, got %d ok=%v", id, ok)
		}
		role, ok := auth.RoleFromContext(r.Context())
		if !ok || role != "admin" {
			t.Fatalf("expected admin role, got %q ok=%v", role, ok)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	if !called {
		t.Fatal("handler not called")
	}
}

func TestGuardUsesJWT(t *testing.T) {
	jwtMgr := auth.NewJWT(auth.JWTConfig{Secret: []byte("jwt-secret"), TTL: time.Hour})
	token, _, err := jwtMgr.Issue(12, "editor")
	if err != nil {
		t.Fatal(err)
	}

	guard := auth.Guard{JWT: jwtMgr}
	called := false
	handler := guard.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || !called {
		t.Fatalf("status=%d called=%v", rec.Code, called)
	}
}
