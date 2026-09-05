package request

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFormRequestValidation(t *testing.T) {
	body := strings.NewReader("email=test@example.com&name=")
	r := httptest.NewRequest("POST", "/", body)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	fr, err := NewFormRequest(r)
	if err != nil {
		t.Fatal(err)
	}
	fr.Required("name")
	if fr.Validate() {
		t.Fatal("expected validation failure")
	}
	if _, ok := fr.Errors["name"]; !ok {
		t.Fatal("expected name error")
	}
}
