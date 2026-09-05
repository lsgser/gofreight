package faker_test

import (
	"testing"

	"github.com/lsgser/gofreight/gftest/faker"
)

func TestEmailNotEmpty(t *testing.T) {
	if faker.Email() == "" {
		t.Fatal("expected email")
	}
}

func TestDefinitionsForModelUser(t *testing.T) {
	defs := faker.DefinitionsForModel("User")
	nameFn, ok := defs["name"].(func() any)
	if !ok {
		t.Fatal("expected lazy name generator")
	}
	if nameFn() == "" {
		t.Fatal("expected name value")
	}
}

func TestForFieldTypeLazy(t *testing.T) {
	fn := faker.ForFieldType("email")
	v, ok := fn().(string)
	if !ok || v == "" {
		t.Fatalf("got %#v", fn())
	}
}

func TestLazyProducesValues(t *testing.T) {
	a := faker.Lazy(func() any { return faker.Word() })
	b := faker.Lazy(func() any { return faker.Word() })
	// Both should return non-empty strings (not guaranteed unique).
	if a() == "" || b() == "" {
		t.Fatal("expected words")
	}
}
