package middleware

import (
	"net/http"
	"strings"
)

// CORS adds Cross-Origin Resource Sharing headers.
func CORS(origins ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool)
	for _, o := range origins {
		allowed[o] = true
	}
	allowAll := len(origins) == 0 || (len(origins) == 1 && origins[0] == "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Version")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// APIVersion validates the X-API-Version header or URL prefix.
func APIVersion(version string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := r.Header.Get("X-API-Version")
			if v == "" {
				parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
				if len(parts) > 0 && strings.HasPrefix(parts[0], "v") {
					v = parts[0]
				}
			}
			if v != "" && v != version {
				http.Error(w, "Unsupported API version", http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adds common security headers.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
