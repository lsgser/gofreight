package api

import "github.com/lsgser/gofreight/router"

// Group registers API routes under /api/{version} with optional middleware.
//
//	API.Group(r, "v1", func(api *router.Router) {
//	    api.ApiResource("posts", handlers)
//	}, auth.JWTMiddleware(jwtMgr))
// Or: auth.Guard{JWT: jwtMgr, TokenStore: store, SessionKey: "current_user_id"}.Middleware
func Group(r *router.Router, version string, fn func(*router.Router), middleware ...router.MiddlewareFunc) {
	g := r.Group(fn).Prefix(VersionPrefix(version)).Name("api.")
	if len(middleware) > 0 {
		g.Use(middleware...)
	}
	g.Apply()
}
