package integrations

/*
|--------------------------------------------------------------------------
| Providers
|--------------------------------------------------------------------------
|
| Implements Providers as part of the integrations package in the
| Gofreight framework. Key symbols: Provider, RegisterProvider,
| BootProviders.
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
| Symbols defined here include: Provider (exported type); RegisterProvider
| (RegisterProvider adds an app or package-level integration provider.);
| BootProviders (BootProviders runs Register on all providers, configures
| integrations, then Boot.).
| 
*/

import "sync"

// Provider registers and boots third-party integrations at application startup.
type Provider interface {
	// Name identifies the provider (for logging and debugging).
	Name() string
	// Register adds integration factories to the registry.
	Register(r *Registry)
	// Boot runs after ConfigureAll; optional post-configuration setup.
	Boot(env EnvReader) error
}

var (
	providerMu sync.RWMutex
	providers  []Provider
)

// RegisterProvider adds an app or package-level integration provider.
// Call from init() in your application, before app.Run().
func RegisterProvider(p Provider) {
	providerMu.Lock()
	defer providerMu.Unlock()
	providers = append(providers, p)
}

// BootProviders runs Register on all providers, configures integrations, then Boot.
func BootProviders(env EnvReader) error {
	providerMu.RLock()
	list := append([]Provider(nil), providers...)
	providerMu.RUnlock()

	r := Default()
	for _, p := range list {
		p.Register(r)
	}
	if err := ConfigureAll(env); err != nil {
		return err
	}
	for _, p := range list {
		if err := p.Boot(env); err != nil {
			return err
		}
	}
	return nil
}
