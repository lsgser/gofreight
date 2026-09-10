package auth

/*
|--------------------------------------------------------------------------
| Verification
|--------------------------------------------------------------------------
|
| Test suite for Verification in the auth package.
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
	"time"
)

func TestVerificationStore(t *testing.T) {
	store := NewMemoryVerificationStore()
	token, err := store.Create(1, "a@example.com", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	id, email, ok := store.Consume(token)
	if !ok || id != 1 || email != "a@example.com" {
		t.Fatalf("consume failed id=%d email=%s ok=%v", id, email, ok)
	}
}
