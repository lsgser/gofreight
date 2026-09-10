package plugins

/*
|--------------------------------------------------------------------------
| Plugins
|--------------------------------------------------------------------------
|
| Implements Plugins as part of the plugins package in the Gofreight
| framework. Key symbols: Hook, Registry, Register, Run, Events.
| 
| Plugins register extension hooks so packages can subscribe to framework
| events without modifying core code.
| 
| Use for optional packages or internal modules that need boot-time
| registration.
| 
| Symbols defined here include: Hook (exported type); Registry (exported
| type); Register (Register adds a hook for an event (boot, route,
| migrate, etc.).); Run (Run executes all hooks for an event.); Events
| (Events lists registered hook names.).
| 
*/

import "sync"

// Hook is a lifecycle callback registered by plugins.
type Hook func(args ...any) error

// Registry manages plugin lifecycle hooks.
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
