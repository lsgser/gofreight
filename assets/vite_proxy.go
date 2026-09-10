package assets

/*
|--------------------------------------------------------------------------
| Vite Proxy
|--------------------------------------------------------------------------
|
| Implements Vite Proxy as part of the assets package in the Gofreight
| framework. Key symbols: ViteProxyMiddleware.
| 
| The assets package connects your public/ directory and optional Vite dev
| server to HTTP handlers.
| 
| In development, Vite can proxy hot module replacement; in production, a
| manifest maps entry points to built files.
| 
| GFT templates use #vite and /assets/ paths documented in the templating
| guide.
| 
| Symbols defined here include: ViteProxyMiddleware (ViteProxyMiddleware
| forwards Vite dev-server paths when public/hot exists.).
| 
*/

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// ViteProxyMiddleware forwards Vite dev-server paths when public/hot exists.
func ViteProxyMiddleware(v *Vite) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if v == nil || !v.IsRunning() {
				next.ServeHTTP(w, r)
				return
			}
			if !shouldProxyToVite(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			target, err := url.Parse(strings.TrimSuffix(v.DevServerURL, "/"))
			if err != nil {
				http.Error(w, "vite proxy misconfigured", http.StatusBadGateway)
				return
			}
			proxy := httputil.NewSingleHostReverseProxy(target)
			r.Host = target.Host
			proxy.ServeHTTP(w, r)
		})
	}
}

func shouldProxyToVite(path string) bool {
	return strings.HasPrefix(path, "/@vite") ||
		strings.HasPrefix(path, "/@fs/") ||
		strings.HasPrefix(path, "/resources/") ||
		strings.HasPrefix(path, "/node_modules/")
}
