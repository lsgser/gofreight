package integrations

/*
|--------------------------------------------------------------------------
| Envresolve
|--------------------------------------------------------------------------
|
| Implements Envresolve as part of the integrations package in the
| Gofreight framework.
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
*/

import "strings"

func envGet(env EnvReader, keys ...string) string {
	for _, key := range keys {
		v := strings.TrimSpace(env.Get(key))
		if v != "" && v != "null" {
			return strings.Trim(v, `"`)
		}
	}
	return ""
}

func providerForCategory(env EnvReader, category Category) string {
	switch category {
	case CategoryMail:
		return envGet(env, "MAIL_MAILER", "MAIL_DRIVER", "EMAIL_PROVIDER")
	case CategoryCache:
		return envGet(env, "CACHE_STORE", "CACHE_DRIVER")
	case CategoryStorage:
		return envGet(env, "FILESYSTEM_DISK", "STORAGE_PROVIDER")
	default:
		return envGet(env, ProviderEnvKey(category))
	}
}
