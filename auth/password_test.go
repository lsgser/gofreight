package auth_test

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
