package integrations

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
)

// CustomIntegration lets apps register their own services.
type CustomIntegration struct {
	name    string
	enabled bool
	config  map[string]string
	onInit  func(map[string]string) error
}

// NewCustom creates a user-defined integration.
func NewCustom(name string, onInit func(map[string]string) error) *CustomIntegration {
	return &CustomIntegration{name: name, onInit: onInit, config: make(map[string]string)}
}

func (c *CustomIntegration) Name() string { return c.name }

func (c *CustomIntegration) Configure(env EnvReader) error {
	prefix := envKeyPrefix(c.name)
	for _, key := range []string{"URL", "API_KEY", "SECRET", "ENABLED"} {
		envKey := prefix + key
		if v := env.Get(envKey); v != "" {
			c.config[key] = v
		}
	}
	if c.onInit != nil {
		if err := c.onInit(c.config); err != nil {
			return err
		}
	}
	c.enabled = c.config["ENABLED"] == "true" || c.config["API_KEY"] != "" || c.config["URL"] != ""
	return nil
}

func (c *CustomIntegration) Enabled() bool { return c.enabled }

// Config returns the integration configuration.
func (c *CustomIntegration) Config() map[string]string {
	return c.config
}

func envKeyPrefix(name string) string {
	// "my_service" -> "MY_SERVICE_"
	result := ""
	for _, c := range name {
		if c == '_' || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			if c >= 'a' && c <= 'z' {
				result += string(c - 32)
			} else {
				result += string(c)
			}
		}
	}
	return result + "_"
}

// RegisterCustom adds a custom integration to the default registry.
func RegisterCustom(name string, onInit func(map[string]string) error) {
	Register(name, func() Integration {
		return NewCustom(name, onInit)
	})
}

// EventBus provides simple in-app pub/sub for integration events.
type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]func(Event)
}

// Event represents an integration event.
type Event struct {
	Name    string
	Payload map[string]any
}

var bus = &EventBus{handlers: make(map[string][]func(Event))}

// Subscribe registers an event handler.
func Subscribe(event string, fn func(Event)) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[event] = append(bus.handlers[event], fn)
}

// Publish emits an event to subscribers.
func Publish(ctx context.Context, event Event) {
	bus.mu.RLock()
	handlers := bus.handlers[event.Name]
	bus.mu.RUnlock()
	for _, fn := range handlers {
		fn(event)
	}
}

// StatusHandler returns an HTTP handler showing integration status (for admin/dev).
func StatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := make(map[string]bool)
		for name, i := range Default().All() {
			status[name] = i.Enabled()
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"integrations": status,
			"active":         List(),
		})
	}
}
