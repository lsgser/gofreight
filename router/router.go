package router

import (
	"fmt"
	"net/http"
	"strings"
)

// HandlerFunc is an alias for the standard HTTP handler signature.
type HandlerFunc = http.HandlerFunc

// Route represents a single HTTP route.
type Route struct {
	Method     string
	Path       string
	Name       string
	Handler    HandlerFunc
	PathPrefix string // non-empty means prefix match
}

// Router maps HTTP requests to handlers. Inspired by Rails' routes.rb.
type Router struct {
	routes     []Route
	middleware []func(http.Handler) http.Handler
	prefix     string
}

// New creates an empty router.
func New() *Router {
	return &Router{}
}

// Use adds middleware to the router stack.
func (r *Router) Use(mw ...func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, mw...)
}

// Group creates a sub-router with a path prefix (like Rails namespaces/scopes).
func (r *Router) Group(prefix string, fn func(*Router)) {
	sub := &Router{
		prefix:     r.prefix + prefix,
		middleware: r.middleware,
	}
	fn(sub)
	r.routes = append(r.routes, sub.routes...)
}

// Get registers a GET route.
func (r *Router) Get(path string, handler HandlerFunc, name ...string) {
	r.add("GET", path, handler, name...)
}

// Post registers a POST route.
func (r *Router) Post(path string, handler HandlerFunc, name ...string) {
	r.add("POST", path, handler, name...)
}

// Put registers a PUT route.
func (r *Router) Put(path string, handler HandlerFunc, name ...string) {
	r.add("PUT", path, handler, name...)
}

// Patch registers a PATCH route.
func (r *Router) Patch(path string, handler HandlerFunc, name ...string) {
	r.add("PATCH", path, handler, name...)
}

// Delete registers a DELETE route.
func (r *Router) Delete(path string, handler HandlerFunc, name ...string) {
	r.add("DELETE", path, handler, name...)
}

// Resources registers RESTful routes for a resource, like Rails `resources :posts`.
//
//	GET    /posts          -> index
//	GET    /posts/new      -> new
//	POST   /posts          -> create
//	GET    /posts/:id      -> show
//	GET    /posts/:id/edit -> edit
//	PUT    /posts/:id      -> update
//	PATCH  /posts/:id      -> update
//	DELETE /posts/:id      -> destroy
func (r *Router) Resources(name string, handlers ResourceHandlers) {
	base := "/" + name
	idPath := base + "/:id"

	if handlers.Index != nil {
		r.Get(base, handlers.Index, name+".index")
	}
	if handlers.New != nil {
		r.Get(base+"/new", handlers.New, name+".new")
	}
	if handlers.Create != nil {
		r.Post(base, handlers.Create, name+".create")
	}
	if handlers.Show != nil {
		r.Get(idPath, handlers.Show, name+".show")
	}
	if handlers.Edit != nil {
		r.Get(idPath+"/edit", handlers.Edit, name+".edit")
	}
	if handlers.Update != nil {
		r.Put(idPath, handlers.Update, name+".update")
		r.Patch(idPath, handlers.Update, name+".update")
	}
	if handlers.Destroy != nil {
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

// Mount registers a handler for all requests under a path prefix.
func (r *Router) Mount(prefix string, handler http.Handler) {
	r.routes = append(r.routes, Route{
		Method:     "*",
		PathPrefix: r.prefix + prefix,
		Handler:    handler.ServeHTTP,
	})
}

func (r *Router) add(method, path string, handler HandlerFunc, name ...string) {
	routeName := ""
	if len(name) > 0 {
		routeName = name[0]
	}
	r.routes = append(r.routes, Route{
		Method:  method,
		Path:    r.prefix + path,
		Name:    routeName,
		Handler: handler,
	})
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := r.match(req)
	if handler == nil {
		http.NotFound(w, req)
		return
	}

	var h http.Handler = http.HandlerFunc(handler)
	for i := len(r.middleware) - 1; i >= 0; i-- {
		h = r.middleware[i](h)
	}
	h.ServeHTTP(w, req)
}

func (r *Router) match(req *http.Request) HandlerFunc {
	for _, route := range r.routes {
		if route.Method != "*" && route.Method != req.Method {
			continue
		}
		if route.PathPrefix != "" {
			if strings.HasPrefix(req.URL.Path, route.PathPrefix) {
				return route.Handler
			}
			continue
		}
		if params := matchPath(route.Path, req.URL.Path); params != nil {
			for k, v := range params {
				req.SetPathValue(k, v)
			}
			return route.Handler
		}
	}
	return nil
}

// matchPath compares a route pattern against a request path.
// Supports :param segments (e.g. /posts/:id).
func matchPath(pattern, path string) map[string]string {
	pParts := strings.Split(strings.Trim(pattern, "/"), "/")
	rParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(pParts) != len(rParts) {
		return nil
	}

	params := make(map[string]string)
	for i, pp := range pParts {
		if strings.HasPrefix(pp, ":") {
			params[strings.TrimPrefix(pp, ":")] = rParts[i]
		} else if pp != rParts[i] {
			return nil
		}
	}
	return params
}

// Routes returns a human-readable list of registered routes (for debugging).
func (r *Router) Routes() []string {
	var lines []string
	for _, route := range r.routes {
		name := route.Name
		if name == "" {
			name = "-"
		}
		lines = append(lines, fmt.Sprintf("%-7s %s  %s", route.Method, route.Path, name))
	}
	return lines
}
