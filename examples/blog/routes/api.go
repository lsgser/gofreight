package routes

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
|
| Stateless JSON API routes. Registered inside a route group in
| routes/register.go with prefix /api/v1.
|
| Example:
|
|   api.ApiResource("posts", router.ApiResourceHandlers{
|       Index:   controller.Handler(PostAPIController{}.Index),
|       Store:   controller.Handler(PostAPIController{}.Store),
|       Show:    controller.Handler(PostAPIController{}.Show),
|       Update:  controller.Handler(PostAPIController{}.Update),
|       Destroy: controller.Handler(PostAPIController{}.Destroy),
|   })
|
| Group middleware (auth, rate limit):
|
|   r.Group(func(api *router.Router) { ... }).
|       Prefix("/api/v1").
|       Use(auth.APITokenMiddleware(store)).
|       Apply()
|
*/

import (
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
)

// API registers JSON API routes (prefix applied by Register).
func API(r *router.Router) {
	r.Get("/health", controller.Handler(func(base controller.Base) error {
		base.RenderJSON(map[string]string{"status": "ok"})
		return nil
	}), "health")
}
