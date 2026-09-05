package routes

/*
|--------------------------------------------------------------------------
| Web Routes
|--------------------------------------------------------------------------
|
| Register routes for browser requests here. These routes receive session
| state, CSRF protection, and typically return GFT HTML views.
|
| Generate a full CRUD resource:
|   gofreight make:scaffold Post title:string body:text
|
| Then register the generated routes below, e.g.:
|   controllers.RegisterPostRoutes(r)
|
*/

import (
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
)

// Web registers browser-facing HTTP routes.
func Web(r *router.Router) {
	r.Get("/", controller.Handler(func(base controller.Base) error {
		return base.RenderView("home/index", map[string]any{"Name": "demoapp"})
	}))

	// Generated resources register here, e.g.:
	// controllers.RegisterPostRoutes(r)
}
