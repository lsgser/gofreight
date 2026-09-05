package auth

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
