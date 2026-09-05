package routes

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
|
| Register stateless JSON API routes here. These routes are prefixed with
| /api/v1 and are intended for mobile clients, SPAs, and third-party
| consumers.
|
| Generate an API controller:
|   gofreight make:api Post title:string body:text
|
*/

import (
	"github.com/lsgser/gofreight/router"
)

// API registers JSON API routes under /api/v1.
func API(r *router.Router) {
	// Example:
	// r.Get("/posts", controller.Handler((&controllers.PostAPIController{}).Index))
	_ = r
}
