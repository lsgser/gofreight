package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// HandlerFunc is an alias for the standard HTTP handler signature.
type HandlerFunc = http.HandlerFunc

// Route represents a single HTTP route.
type Route struct {
	Method      string
	Path        string
	Name        string
	Handler     HandlerFunc
	PathPrefix  string
	Domain      string
	middleware  []MiddlewareFunc
	constraints map[string]*regexp.Regexp
}

// RouteRegistrar chains options (such as middleware) onto a single route.
type RouteRegistrar struct {
	router *Router
	index  int
}

// Use attaches middleware to this route (runs after group middleware, before the handler).
func (reg *RouteRegistrar) Use(middleware ...MiddlewareFunc) *RouteRegistrar {
	if reg == nil || reg.index < 0 || reg.index >= len(reg.router.routes) {
		return reg
	}
	route := &reg.router.routes[reg.index]
	route.middleware = append(route.middleware, middleware...)
	return reg
}

// Router maps HTTP requests to handlers. Supports route groups and REST resources.
type Router struct {
	routes          []Route
	middleware      []MiddlewareFunc
	prefix          string
	namePrefix      string
	groupMiddleware []MiddlewareFunc
	fallback        HandlerFunc
	domain          string
	defaultDomain   string
}

// New creates an empty router.
func New() *Router {
	return &Router{}
}

// Use adds global middleware to the router stack.
func (r *Router) Use(mw ...MiddlewareFunc) {
	r.middleware = append(r.middleware, mw...)
}

// Get registers a GET route.
func (r *Router) Get(path string, handler HandlerFunc, name ...string) *RouteRegistrar {
	return r.add("GET", path, handler, name...)
}

// Post registers a POST route.
func (r *Router) Post(path string, handler HandlerFunc, name ...string) *RouteRegistrar {
	return r.add("POST", path, handler, name...)
}

// Put registers a PUT route.
func (r *Router) Put(path string, handler HandlerFunc, name ...string) *RouteRegistrar {
	return r.add("PUT", path, handler, name...)
}

// Patch registers a PATCH route.
func (r *Router) Patch(path string, handler HandlerFunc, name ...string) *RouteRegistrar {
	return r.add("PATCH", path, handler, name...)
}

// Delete registers a DELETE route.
func (r *Router) Delete(path string, handler HandlerFunc, name ...string) *RouteRegistrar {
	return r.add("DELETE", path, handler, name...)
}

// Resources registers RESTful HTML routes (includes new/edit).
func (r *Router) Resources(name string, handlers ResourceHandlers, opts ...ResourceOptions) {
	var opt ResourceOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	base := "/" + strings.Trim(name, "/")
	idPath := base + "/:id"

	if opt.allows("index") && handlers.Index != nil {
		r.Get(base, handlers.Index, name+".index")
	}
	if opt.allows("new") && handlers.New != nil {
		r.Get(base+"/new", handlers.New, name+".new")
	}
	if opt.allows("create") && handlers.Create != nil {
		r.Post(base, handlers.Create, name+".create")
	}
	if opt.allows("show") && handlers.Show != nil {
		r.Get(idPath, handlers.Show, name+".show")
	}
	if opt.allows("edit") && handlers.Edit != nil {
		r.Get(idPath+"/edit", handlers.Edit, name+".edit")
	}
	if opt.allows("update") && handlers.Update != nil {
		r.Put(idPath, handlers.Update, name+".update")
		r.Patch(idPath, handlers.Update, name+".update")
	}
	if opt.allows("destroy") && handlers.Destroy != nil {
		r.Delete(idPath, handlers.Destroy, name+".destroy")
	}
}

// ApiResource registers JSON API REST routes (no new/edit).
//
//	GET    /posts      -> index
//	POST   /posts      -> store
//	GET    /posts/:id  -> show
//	PUT    /posts/:id  -> update
//	PATCH  /posts/:id  -> update
//	DELETE /posts/:id  -> destroy
func (r *Router) ApiResource(name string, handlers ApiResourceHandlers, opts ...ResourceOptions) {
	var opt ResourceOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	base := "/" + strings.Trim(name, "/")
	idPath := base + "/:id"

	if opt.allows("index") && handlers.Index != nil {
		r.Get(base, handlers.Index, name+".index")
	}
	if opt.allows("store") && handlers.Store != nil {
		r.Post(base, handlers.Store, name+".store")
	}
	r.Get(base+"/new", apiNotFound)
	r.Get(base+"/:id/edit", apiNotFound)
	if opt.allows("show") && handlers.Show != nil {
		r.Get(idPath, handlers.Show, name+".show")
	}
	if opt.allows("update") && handlers.Update != nil {
		r.Put(idPath, handlers.Update, name+".update")
		r.Patch(idPath, handlers.Update, name+".update")
	}
	if opt.allows("destroy") && handlers.Destroy != nil {
		r.Delete(idPath, handlers.Destroy, name+".destroy")
	}
}

// ResourceHandlers holds RESTful action handlers for Resources().
type ResourceHandlers struct {
	Index   HandlerFunc
	New     HandlerFunc
	Create  HandlerFunc
	Show    HandlerFunc
	Edit    HandlerFunc
	Update  HandlerFunc
	Destroy HandlerFunc
}

// ApiResourceHandlers holds API REST action handlers for ApiResource().
type ApiResourceHandlers struct {
	Index   HandlerFunc
	Store   HandlerFunc
	Show    HandlerFunc
	Update  HandlerFunc
	Destroy HandlerFunc
}

// Mount registers a handler for all requests under a path prefix.
func (r *Router) Mount(prefix string, handler http.Handler) {
	r.routes = append(r.routes, Route{
		Method:     "*",
		PathPrefix: joinPaths(r.prefix, prefix),
		Handler:    handler.ServeHTTP,
	})
}

func (r *Router) add(method, path string, handler HandlerFunc, name ...string) *RouteRegistrar {
	routeName := ""
	if len(name) > 0 {
		routeName = r.namePrefix + name[0]
	}
	fullPath := joinPaths(r.prefix, path)
	fullPath = normalizeRoutePath(fullPath)
	if fullPath == "" {
		fullPath = "/"
	}
	r.routes = append(r.routes, Route{
		Method:     method,
		Path:       fullPath,
		Name:       routeName,
		Handler:    handler,
		Domain:     r.domain,
		middleware: append([]MiddlewareFunc{}, r.groupMiddleware...),
	})
	return &RouteRegistrar{router: r, index: len(r.routes) - 1}
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	route := r.match(req)
	if route == nil {
		if r.fallback != nil {
			r.fallback(w, req)
			return
		}
		http.NotFound(w, req)
		return
	}

	var h http.Handler = http.HandlerFunc(route.Handler)
	for i := len(route.middleware) - 1; i >= 0; i-- {
		h = route.middleware[i](h)
	}
	for i := len(r.middleware) - 1; i >= 0; i-- {
		h = r.middleware[i](h)
	}
	h.ServeHTTP(w, req)
}

func (r *Router) match(req *http.Request) *Route {
	var best *Route
	bestScore := -1
	var bestParams map[string]string

	for i := range r.routes {
		route := &r.routes[i]
		if route.Method != "*" && route.Method != req.Method {
			continue
		}
		if route.PathPrefix != "" {
			if strings.HasPrefix(req.URL.Path, route.PathPrefix) {
				return route
			}
			continue
		}
		hostParams, ok := matchHost(req.Host, route.Domain)
		if !ok {
			continue
		}
		if params := matchPathPattern(route.Path, req.URL.Path); params != nil {
			if !paramsMatchConstraints(params, route.constraints) {
				continue
			}
			score := routeSpecificity(route.Path)
			if score > bestScore {
				merged := make(map[string]string, len(params)+len(hostParams))
				for k, v := range params {
					merged[k] = v
				}
				for k, v := range hostParams {
					merged["host_"+k] = v
				}
				best = route
				bestScore = score
				bestParams = merged
			}
		}
	}
	if best != nil {
		for k, v := range bestParams {
			req.SetPathValue(k, v)
		}
	}
	return best
}

func apiNotFound(w http.ResponseWriter, req *http.Request) {
	http.NotFound(w, req)
}

// Routes returns a human-readable list of registered routes (for debugging).
func (r *Router) Routes() []string {
	var lines []string
	for _, route := range r.routes {
		name := route.Name
		if name == "" {
			name = "-"
		}
		path := route.Path
		if route.PathPrefix != "" {
			path = route.PathPrefix + "/*"
		}
		lines = append(lines, fmt.Sprintf("%-7s %s  %s", route.Method, path, name))
	}
	return lines
}
