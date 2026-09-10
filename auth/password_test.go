package auth_test

/*
|--------------------------------------------------------------------------
| Password
|--------------------------------------------------------------------------
|
| Test suite for Password in the auth package.
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
	"testing"

	"github.com/lsgser/gofreight/auth"
)

func TestHashPassword(t *testing.T) {
	hash, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatal(err)
	}
	if !auth.CheckPassword(hash, "secret123") {
		t.Fatal("password should match")
	}
	if auth.CheckPassword(hash, "wrong") {
		t.Fatal("wrong password should not match")
	}
}

func TestAuthenticate(t *testing.T) {
	hash, _ := auth.HashPassword("pass")
	user := &auth.User{ID: 1, Email: "a@b.com", PasswordHash: hash}
	if err := auth.Authenticate(user, "pass"); err != nil {
		t.Fatal(err)
	}
	if err := auth.Authenticate(user, "bad"); err == nil {
		t.Fatal("expected error")
	}
}
