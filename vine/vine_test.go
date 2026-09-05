package vine_test

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
