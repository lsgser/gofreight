package validation_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lsgser/gofreight/validation"
)

func TestValidatorRequired(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader("email=test@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	v, err := validation.New(req)
	if err != nil {
		t.Fatal(err)
	}
	v.Required("name")
	if !v.Fails() {
		t.Fatal("expected failure for missing name")
	}
}

func TestValidatorEmail(t *testing.T) {
	v := validation.FromMap(map[string]string{"email": "not-an-email"})
	v.Email("email")
	if !v.Fails() {
		t.Fatal("expected invalid email")
	}
}
