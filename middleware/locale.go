package middleware

/*
|--------------------------------------------------------------------------
| Locale
|--------------------------------------------------------------------------
|
| Implements Locale as part of the middleware package in the Gofreight
| framework. Key symbols: LocaleFromContext, LocaleMiddleware.
| 
| HTTP middleware implements sessions, CSRF, CORS, locale, structured
| logging, rate limiting, and maintenance mode.
| 
| Register global middleware in bootstrap or attach to route groups for
| API-specific stacks.
| 
| Session drivers include file, cookie, and Redis variants selected by
| SESSION_DRIVER.
| 
| Symbols defined here include: LocaleFromContext (LocaleFromContext
| returns the request locale.); LocaleMiddleware (LocaleMiddleware sets
| locale from query, cookie, or Accept-Language.).
| 
*/

import (
	"context"
	"net/http"
	"strings"

	"github.com/lsgser/gofreight/i18n"
)

type localeKey struct{}

// LocaleFromContext returns the request locale.
func LocaleFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(localeKey{}).(string); ok {
		return v
	}
	return "en"
}

// LocaleMiddleware sets locale from query, cookie, or Accept-Language.
func LocaleMiddleware(translator *i18n.Translator, defaultLocale string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			locale := defaultLocale
			if q := r.URL.Query().Get("locale"); q != "" {
				locale = q
			} else if c, err := r.Cookie("locale"); err == nil && c.Value != "" {
				locale = c.Value
			} else if al := r.Header.Get("Accept-Language"); al != "" {
				locale = parseAcceptLanguage(al, defaultLocale)
			}
			if translator != nil {
				translator.SetLocale(locale)
			}
			ctx := context.WithValue(r.Context(), localeKey{}, locale)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func parseAcceptLanguage(header, fallback string) string {
	parts := strings.Split(header, ",")
	if len(parts) == 0 {
		return fallback
	}
	lang := strings.TrimSpace(strings.Split(parts[0], ";")[0])
	if lang == "" {
		return fallback
	}
	return lang
}
