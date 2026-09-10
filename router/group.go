package router

/*
|--------------------------------------------------------------------------
| Group
|--------------------------------------------------------------------------
|
| Implements Group as part of the router package in the Gofreight
| framework. Key symbols: MiddlewareFunc, RouteGroup, Group, GroupPrefix,
| Prefix, Use.
| 
| The router matches verbs and paths, supports groups, prefixes, named
| routes, constraints, signed URLs, and domain routing.
| 
| Routes register in routes/web.go and routes/api.go; see docs/routing.md
| for middleware and model binding.
| 
| Symbols defined here include: MiddlewareFunc (exported type); RouteGroup
| (exported type); Group (Group starts a route group. Call Prefix, Use,
| Name, then Apply.); GroupPrefix (GroupPrefix registers routes under a
| prefix (legacy sugar for Group(...).Prefix(p).Apply()).); Prefix (Prefix
| sets the URI prefix for all routes in the group.); Use (Use attaches
| middleware to all routes in the group (runs before route handlers).);
| Middleware (Middleware is an alias for Use.); Name (Name sets a prefix
| for route names inside the group (e.g. "api." → "api.users.index").).
| 
*/

import (
	"net/http"
	"strings"
)

// MiddlewareFunc is standard HTTP middleware.
type MiddlewareFunc = func(http.Handler) http.Handler

// RouteGroup builds route groups: callback first, then Prefix/Use/Name, then Apply.
//
//	r.Group(func(api *Router) {
//	    api.Get("/users", handler)
//	}).Prefix("/api/v1").Use(authMw).Name("api.").Apply()
type RouteGroup struct {
	parent     *Router
	fn         func(*Router)
	prefix     string
	namePrefix string
	middleware []MiddlewareFunc
	domain     string
	subdomain  string
}

// Group starts a route group. Call Prefix, Use, Name, then Apply.
func (r *Router) Group(fn func(*Router)) *RouteGroup {
	return &RouteGroup{parent: r, fn: fn}
}

// GroupPrefix registers routes under a prefix (legacy sugar for Group(...).Prefix(p).Apply()).
func (r *Router) GroupPrefix(prefix string, fn func(*Router)) {
	r.Group(fn).Prefix(prefix).Apply()
}

// Prefix sets the URI prefix for all routes in the group.
func (g *RouteGroup) Prefix(prefix string) *RouteGroup {
	g.prefix = joinPaths(g.prefix, prefix)
	return g
}

// Use attaches middleware to all routes in the group (runs before route handlers).
func (g *RouteGroup) Use(middleware ...MiddlewareFunc) *RouteGroup {
	g.middleware = append(g.middleware, middleware...)
	return g
}

// Middleware is an alias for Use.
func (g *RouteGroup) Middleware(middleware ...MiddlewareFunc) *RouteGroup {
	return g.Use(middleware...)
}

// Name sets a prefix for route names inside the group (e.g. "api." → "api.users.index").
func (g *RouteGroup) Name(prefix string) *RouteGroup {
	g.namePrefix = prefix
	return g
}

// Domain restricts routes in the group to a host pattern (e.g. "api.example.com", "{tenant}.example.com").
func (g *RouteGroup) Domain(domain string) *RouteGroup {
	g.domain = domain
	return g
}

// Subdomain prefixes the default domain (e.g. Subdomain("api") → api.example.com).
func (g *RouteGroup) Subdomain(subdomain string) *RouteGroup {
	g.subdomain = subdomain
	return g
}

// Apply registers the group routes on the parent router.
func (g *RouteGroup) Apply() {
	sub := g.parent.childRouter(g.prefix, g.namePrefix, g.middleware, g.domain, g.subdomain)
	g.fn(sub)
	g.parent.routes = append(g.parent.routes, sub.routes...)
}

func (r *Router) childRouter(prefix, namePrefix string, groupMiddleware []MiddlewareFunc, domain, subdomain string) *Router {
	resolvedDomain := resolveGroupDomain(subdomain, domain, r.defaultDomain)
	if resolvedDomain == "" {
		resolvedDomain = r.domain
	}
	return &Router{
		prefix:          joinPaths(r.prefix, prefix),
		namePrefix:      r.namePrefix + namePrefix,
		middleware:      r.middleware,
		groupMiddleware: append(append([]MiddlewareFunc{}, r.groupMiddleware...), groupMiddleware...),
		domain:          resolvedDomain,
		defaultDomain:   r.defaultDomain,
	}
}

func joinPaths(parts ...string) string {
	out := ""
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		out = strings.TrimSuffix(out, "/") + p
	}
	if out == "" {
		return ""
	}
	return out
}
