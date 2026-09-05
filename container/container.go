package container

import (
	"sync"
)

// Container is a simple service container (Laravel-style IoC).
type Container struct {
	mu          sync.RWMutex
	bindings    map[string]func() any
	singletons  map[string]bool
	instances   map[string]any
}

// New creates an empty container.
func New() *Container {
	return &Container{
		bindings:   make(map[string]func() any),
		singletons: make(map[string]bool),
		instances:  make(map[string]any),
	}
}

// Bind registers a transient factory (new instance each Resolve).
func (c *Container) Bind(name string, factory func() any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[name] = factory
	c.singletons[name] = false
	delete(c.instances, name)
}

// Singleton registers a shared instance factory.
func (c *Container) Singleton(name string, factory func() any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[name] = factory
	c.singletons[name] = true
	delete(c.instances, name)
}

// Resolve returns a service by name.
func (c *Container) Resolve(name string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	factory, ok := c.bindings[name]
	if !ok {
		return nil, false
	}
	if c.singletons[name] {
		if inst, ok := c.instances[name]; ok {
			return inst, true
		}
		inst := factory()
		c.instances[name] = inst
		return inst, true
	}
	return factory(), true
}

// MustResolve returns a service or panics.
func (c *Container) MustResolve(name string) any {
	v, ok := c.Resolve(name)
	if !ok {
		panic("container: service " + name + " not bound")
	}
	return v
}

// Has reports whether a binding exists.
func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.bindings[name]
	return ok
}
