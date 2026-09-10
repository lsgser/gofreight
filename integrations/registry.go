package integrations

/*
|--------------------------------------------------------------------------
| Registry
|--------------------------------------------------------------------------
|
| Implements Registry as part of the integrations package in the Gofreight
| framework. Key symbols: Integration, EnvReader, OsEnv, Get, Registry,
| NewRegistry.
| 
| Integrations register pluggable drivers for mail, storage, cache, queue,
| and custom third-party APIs.
| 
| Active() resolves the configured implementation from environment
| variables; wire Application in ConfigureIntegrations.
| 
| Built-in connectors cover SMTP, SendGrid, S3-compatible storage, and
| Redis without vendor-specific SDKs in app code.
| 
| Symbols defined here include: Integration (exported type); EnvReader
| (exported type); OsEnv (exported type); Registry (exported type);
| NewRegistry (NewRegistry creates an integration registry.); Register
| (Register adds an integration factory to the registry.); Register
| (Register adds a factory to this registry.); ConfigureAll (ConfigureAll
| initializes all registered integrations from environment.).
| 
*/

import (
	"fmt"
	"os"
	"sync"
)

// Integration represents a third-party or cloud service connection.
type Integration interface {
	// Name returns the integration identifier (e.g. "storage", "email").
	Name() string
	// Configure initializes the integration from environment/config.
	Configure(env EnvReader) error
	// Enabled returns true if the integration is configured and active.
	Enabled() bool
}

// EnvReader reads configuration values (typically os.Getenv).
type EnvReader interface {
	Get(key string) string
}

// OsEnv uses environment variables.
type OsEnv struct{}

func (OsEnv) Get(key string) string { return os.Getenv(key) }

// Registry manages registered integrations.
type Registry struct {
	mu           sync.RWMutex
	integrations map[string]Integration
	factories    map[string]func() Integration
}

// Global registry instance.
var defaultRegistry = NewRegistry()

// NewRegistry creates an integration registry.
func NewRegistry() *Registry {
	return &Registry{
		integrations: make(map[string]Integration),
		factories:    make(map[string]func() Integration),
	}
}

// Register adds an integration factory to the registry.
func Register(name string, factory func() Integration) {
	defaultRegistry.Register(name, factory)
}

// Register adds a factory to this registry.
func (r *Registry) Register(name string, factory func() Integration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// ConfigureAll initializes all registered integrations from environment.
func ConfigureAll(env EnvReader) error {
	return defaultRegistry.ConfigureAll(env)
}

// ConfigureAll initializes all registered integrations.
func (r *Registry) ConfigureAll(env EnvReader) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, factory := range r.factories {
		integration := factory()
		if err := integration.Configure(env); err != nil {
			return fmt.Errorf("integration %s: %w", name, err)
		}
		r.integrations[name] = integration
	}
	return nil
}

// Get returns a configured integration by name.
func Get(name string) (Integration, bool) {
	return defaultRegistry.Get(name)
}

// Get returns an integration from this registry.
func (r *Registry) Get(name string) (Integration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	i, ok := r.integrations[name]
	return i, ok
}

// List returns all configured integration names.
func List() []string {
	return defaultRegistry.List()
}

// List returns enabled integration names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var names []string
	for name, i := range r.integrations {
		if i.Enabled() {
			names = append(names, name)
		}
	}
	return names
}

// All returns all registered integration instances.
func (r *Registry) All() map[string]Integration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]Integration, len(r.integrations))
	for k, v := range r.integrations {
		result[k] = v
	}
	return result
}

// Default returns the global registry.
func Default() *Registry {
	return defaultRegistry
}

// MustGet returns an integration or panics.
func MustGet(name string) Integration {
	i, ok := Get(name)
	if !ok || !i.Enabled() {
		panic(fmt.Sprintf("integration %q not configured", name))
	}
	return i
}

func init() {
	Register("storage", func() Integration { return &Storage{} })
	Register("email", func() Integration { return &Email{} })
	Register("redis", func() Integration { return &Redis{} })
	Register("webhook", func() Integration { return &Webhook{} })
	Register("analytics", func() Integration { return &Analytics{} })
}
