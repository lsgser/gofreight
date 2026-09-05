package integrations

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
