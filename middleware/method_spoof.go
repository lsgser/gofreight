package middleware

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
