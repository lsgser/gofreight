package plugins

import "sync"

// Hook is a lifecycle callback registered by plugins.
type Hook func(args ...any) error

// Registry manages plugin hooks (like Rails engines).
type Registry struct {
	mu    sync.RWMutex
	hooks map[string][]Hook
}

var defaultRegistry = &Registry{hooks: make(map[string][]Hook)}

// Register adds a hook for an event (boot, route, migrate, etc.).
func Register(event string, hook Hook) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()
	defaultRegistry.hooks[event] = append(defaultRegistry.hooks[event], hook)
}

// Run executes all hooks for an event.
func Run(event string, args ...any) error {
	defaultRegistry.mu.RLock()
	hooks := defaultRegistry.hooks[event]
	defaultRegistry.mu.RUnlock()
	for _, h := range hooks {
		if err := h(args...); err != nil {
			return err
		}
	}
	return nil
}

// Events lists registered hook names.
func Events() []string {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	var names []string
	for k := range defaultRegistry.hooks {
		names = append(names, k)
	}
	return names
}
