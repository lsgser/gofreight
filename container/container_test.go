package container

/*
|--------------------------------------------------------------------------
| Container
|--------------------------------------------------------------------------
|
| Test suite for Container in the container package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| The container package implements a lightweight service container for
| bindings and singletons.
| 
| Register factories in bootstrap/app.go; resolve services from
| controllers or jobs by name or type.
| 
| Pattern mirrors Laravel's container at a smaller scale for explicit
| wiring in Go.
| 
| Run with go test ./container/... or go test for this package from the
| framework root.
| 
*/

import "testing"

type greeter struct{ msg string }

func TestContainerSingleton(t *testing.T) {
	c := New()
	c.Singleton("greeter", func() any { return &greeter{msg: "hi"} })
	a := c.MustResolve("greeter").(*greeter)
	b := c.MustResolve("greeter").(*greeter)
	if a != b {
		t.Fatal("expected same singleton instance")
	}
}

func TestContainerBindTransient(t *testing.T) {
	c := New()
	c.Bind("greeter", func() any { return &greeter{msg: "hi"} })
	a := c.MustResolve("greeter").(*greeter)
	b := c.MustResolve("greeter").(*greeter)
	if a == b {
		t.Fatal("expected different transient instances")
	}
}
