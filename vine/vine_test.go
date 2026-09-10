package vine_test

/*
|--------------------------------------------------------------------------
| Vine
|--------------------------------------------------------------------------
|
| Test suite for Vine in the vine package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| Vine is a fluent validation DSL: vine.String().Required().Email() builds
| schemas validated against form maps or JSON.
| 
| Used in form requests, API payloads, and anywhere you want Laravel-like
| validation chains in Go.
| 
| Run with go test ./vine/... or go test for this package from the
| framework root.
| 
*/

import (
	"testing"

	"github.com/lsgser/gofreight/vine"
)

func TestObjectSchema(t *testing.T) {
	schema := vine.Object(map[string]vine.Rule{
		"title": vine.String().Required().MinLength(3),
		"email": vine.String().Required().Email(),
	})

	_, errs := schema.Validate(map[string]string{
		"title": "ab",
		"email": "bad",
	})
	if len(errs) != 2 {
		t.Fatalf("expected 2 fields with errors, got %d", len(errs))
	}

	out, errs := schema.Validate(map[string]string{
		"title": "Hello",
		"email": "user@example.com",
	})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if out["title"] != "Hello" {
		t.Fatalf("expected title preserved")
	}
}
