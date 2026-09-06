package integrations

import "strings"

// Category groups third-party adapters by service type (mail, cache, queue, storage, etc.).
// Each category can be backed by any registered driver.
type Category string

const (
	CategoryMail      Category = "mail"
	CategoryStorage   Category = "storage"
	CategoryCache     Category = "cache"
	CategoryPayment   Category = "payment"
	CategoryAnalytics Category = "analytics"
	CategoryWebhook   Category = "webhook"
)

// CategorizedDriver marks an integration with its service category and driver id.
type CategorizedDriver interface {
	Integration
	Category() Category
	DriverID() string
}

// ProviderEnvKey returns the env var that selects the active driver for a category.
func ProviderEnvKey(c Category) string {
	switch c {
	case CategoryMail:
		return "EMAIL_PROVIDER"
	case CategoryStorage:
		return "STORAGE_PROVIDER"
	case CategoryCache:
		return "CACHE_DRIVER"
	case CategoryPayment:
		return "PAYMENT_PROVIDER"
	case CategoryAnalytics:
		return "ANALYTICS_PROVIDER"
	case CategoryWebhook:
		return "WEBHOOK_DRIVER"
	default:
		return strings.ToUpper(string(c)) + "_PROVIDER"
	}
}

// AsCategorizedDriver returns a driver if the integration implements CategorizedDriver.
func AsCategorizedDriver(i Integration) (CategorizedDriver, bool) {
	cd, ok := i.(CategorizedDriver)
	return cd, ok && cd.Enabled()
}

// Active returns the configured integration for a service category.
// Resolution order: explicit provider env → first enabled driver in category.
func Active(category Category, env EnvReader) (Integration, bool) {
	_ = ConfigureAll(env)
	provider := strings.ToLower(strings.TrimSpace(providerForCategory(env, category)))
	if provider == "" {
		provider = strings.ToLower(strings.TrimSpace(env.Get(ProviderEnvKey(category))))
	}
	if provider != "" {
		if i, ok := Get(provider); ok {
			if cd, ok := AsCategorizedDriver(i); ok && cd.Category() == category {
				return i, true
			}
		}
		for _, i := range Default().All() {
			cd, ok := AsCategorizedDriver(i)
			if !ok || cd.Category() != category {
				continue
			}
			if cd.DriverID() == provider {
				return i, true
			}
		}
		return nil, false
	}
	for _, i := range Default().All() {
		cd, ok := AsCategorizedDriver(i)
		if ok && cd.Category() == category {
			return i, true
		}
	}
	return nil, false
}

// ActiveStorage returns the configured storage driver.
func ActiveStorage(env EnvReader) (*Storage, bool) {
	i, ok := Active(CategoryStorage, env)
	if !ok {
		return nil, false
	}
	return AsStorage(i)
}

// ActiveEmail returns the configured mail driver.
func ActiveEmail(env EnvReader) (*Email, bool) {
	i, ok := Active(CategoryMail, env)
	if !ok {
		return nil, false
	}
	return AsEmail(i)
}

// ActiveRedis returns the configured cache driver when CACHE_DRIVER=redis.
func ActiveRedis(env EnvReader) (*Redis, bool) {
	i, ok := Active(CategoryCache, env)
	if !ok {
		return nil, false
	}
	return AsRedis(i)
}

// ActiveAnalytics returns the configured analytics driver.
func ActiveAnalytics(env EnvReader) (*Analytics, bool) {
	i, ok := Active(CategoryAnalytics, env)
	if !ok {
		return nil, false
	}
	a, ok := i.(*Analytics)
	return a, ok && a.Enabled()
}

// ActiveWebhook returns the configured webhook driver.
func ActiveWebhook(env EnvReader) (*Webhook, bool) {
	i, ok := Active(CategoryWebhook, env)
	if !ok {
		return nil, false
	}
	w, ok := i.(*Webhook)
	return w, ok && w.Enabled()
}

// ActivePayment returns the configured payment driver registered by your app.
func ActivePayment(env EnvReader) (Integration, bool) {
	return Active(CategoryPayment, env)
}

// DriversInCategory lists enabled driver ids for a category.
func DriversInCategory(category Category) []string {
	var ids []string
	for _, i := range Default().All() {
		cd, ok := AsCategorizedDriver(i)
		if ok && cd.Category() == category {
			ids = append(ids, cd.DriverID())
		}
	}
	return ids
}
