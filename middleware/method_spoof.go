package middleware

/*
|--------------------------------------------------------------------------
| Method Spoof
|--------------------------------------------------------------------------
|
| Implements Method Spoof as part of the middleware package in the
| Gofreight framework. Key symbols: MethodSpoof.
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
| Symbols defined here include: MethodSpoof (MethodSpoof reads _method
| from POST bodies and rewrites the request method (PUT/PATCH/DELETE).).
| 
*/

import (
	"net/http"
	"strings"
)

// MethodSpoof reads _method from POST bodies and rewrites the request method (PUT/PATCH/DELETE).
func MethodSpoof(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = r.ParseForm()
			if method := strings.ToUpper(strings.TrimSpace(r.FormValue("_method"))); method != "" {
				switch method {
				case http.MethodPut, http.MethodPatch, http.MethodDelete:
					r.Method = method
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
