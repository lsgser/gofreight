package request

/*
|--------------------------------------------------------------------------
| Form
|--------------------------------------------------------------------------
|
| Test suite for Form in the request package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| Form requests bind and validate HTTP input using vine schemas or
| validation tags before controller actions run.
| 
| Integrates with controller Base for 422 Unprocessable responses and GFT
| error display helpers.
| 
| Run with go test ./request/... or go test for this package from the
| framework root.
| 
*/

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
