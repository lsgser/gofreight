package router

/*
|--------------------------------------------------------------------------
| Extra
|--------------------------------------------------------------------------
|
| Implements Extra as part of the router package in the Gofreight
| framework. Key symbols: Any, Match, Head, Options, Fallback.
| 
| The router matches verbs and paths, supports groups, prefixes, named
| routes, constraints, signed URLs, and domain routing.
| 
| Routes register in routes/web.go and routes/api.go; see docs/routing.md
| for middleware and model binding.
| 
| Symbols defined here include: Any (Any registers a route for all common
| HTTP verbs on the same path.); Match (Match registers the same handler
| for multiple HTTP methods.); Head (Head registers a HEAD route.);
| Options (Options registers an OPTIONS route.); Fallback (Fallback
| registers a handler for requests that match no other route.).
| 
*/

// Any registers a route for all common HTTP verbs on the same path.
func (r *Router) Any(path string, handler HandlerFunc, name ...string) *RouteRegistrar {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	return r.Match(methods, path, handler, name...)
}

// Match registers the same handler for multiple HTTP methods.
func (r *Router) Match(methods []string, path string, handler HandlerFunc, name ...string) *RouteRegistrar {
	var reg *RouteRegistrar
	for i, method := range methods {
		n := name
		if i > 0 {
			n = nil
		}
		reg = r.add(method, path, handler, n...)
	}
	return reg
}

// Head registers a HEAD route.
func (r *Router) Head(path string, handler HandlerFunc, name ...string) *RouteRegistrar {
	return r.add("HEAD", path, handler, name...)
}

// Options registers an OPTIONS route.
func (r *Router) Options(path string, handler HandlerFunc, name ...string) *RouteRegistrar {
	return r.add("OPTIONS", path, handler, name...)
}

// Fallback registers a handler for requests that match no other route.
func (r *Router) Fallback(handler HandlerFunc) {
	r.fallback = handler
}
