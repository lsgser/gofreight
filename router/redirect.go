package router

import (
	"fmt"
	"net/http"
	"strings"
)

// Redirect registers a GET route that redirects to another URL (Laravel Route::redirect).
func (r *Router) Redirect(from, to string, code int) {
	if code == 0 {
		code = http.StatusFound
	}
	status := code
	r.Get(from, func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, to, status)
	})
}

// PermanentRedirect registers a 301 redirect from one path to another.
func (r *Router) PermanentRedirect(from, to string) {
	r.Redirect(from, to, http.StatusMovedPermanently)
}

// URL builds a path for a named route, substituting :param placeholders.
func (r *Router) URL(name string, params map[string]string) (string, error) {
	path, err := r.lookupPath(name)
	if err != nil {
		return "", err
	}
	path = normalizeRoutePath(path)
	for key, val := range params {
		if strings.Contains(path, ":"+key+"*") {
			path = strings.Replace(path, ":"+key+"*", val, 1)
			continue
		}
		if strings.Contains(path, ":"+key+"?") {
			if val == "" {
				path = strings.Replace(path, "/:"+key+"?", "", 1)
				path = strings.Replace(path, ":"+key+"?", "", 1)
			} else {
				path = strings.Replace(path, ":"+key+"?", val, 1)
			}
			continue
		}
		path = strings.Replace(path, ":"+key, val, 1)
	}
	if strings.Contains(path, ":") {
		return "", fmt.Errorf("route %q: missing parameters for %q", name, path)
	}
	return path, nil
}

// URLParams builds a named route URL from alternating key/value pairs.
//
//	r.URLParams("posts.show", "id", "42")
func (r *Router) URLParams(name string, params ...string) (string, error) {
	if len(params)%2 != 0 {
		return "", fmt.Errorf("route %q: URLParams requires key/value pairs", name)
	}
	m := make(map[string]string, len(params)/2)
	for i := 0; i+1 < len(params); i += 2 {
		m[params[i]] = params[i+1]
	}
	return r.URL(name, m)
}

func (r *Router) lookupPath(name string) (string, error) {
	for _, route := range r.routes {
		if route.Name != name || route.PathPrefix != "" {
			continue
		}
		return route.Path, nil
	}
	return "", fmt.Errorf("route %q not found", name)
}
