package validation_test

/*
|--------------------------------------------------------------------------
| Validation
|--------------------------------------------------------------------------
|
| Test suite for Validation in the validation package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| Validation provides rule structs and runners used by models, form
| requests, and the vine DSL.
| 
| Errors map to field names for JSON and GFT #error directives.
| 
| Run with go test ./validation/... or go test for this package from the
| framework root.
| 
*/

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
