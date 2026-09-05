package container

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
