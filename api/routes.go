package api

/*
|--------------------------------------------------------------------------
| Routes
|--------------------------------------------------------------------------
|
| Implements Routes as part of the api package in the Gofreight framework.
| Key symbols: Group.
| 
| The api package implements JSON resource transformers and helpers for
| versioned HTTP APIs.
| 
| Resources map models and structs to consistent JSON shapes, pagination
| metadata, and optional link collections.
| 
| Use api.Group with the router to register /api/v1-style routes with
| shared middleware.
| 
| Symbols defined here include: Group (Group registers API routes under
| /api/{version} with optional middleware.).
| 
*/

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
