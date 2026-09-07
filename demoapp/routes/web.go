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

/*
|--------------------------------------------------------------------------
| Web
|--------------------------------------------------------------------------
|
| Registers browser-facing HTTP routes.
|
*/
func Web(r *router.Router) {
	r.Get("/", controller.Handler(func(base controller.Base) error {
		return base.RenderView("home/index", base.ViewData(map[string]any{
			"Name":             "demoapp",
			"DocsURL":          "https://lsgser.github.io/gofreight-web/",
			"FrameworkVersion": "v0.4.0",
		}))
	}), "home")

	/*
	|--------------------------------------------------------------------------
	| Generated Resources
	|--------------------------------------------------------------------------
	|
	| Register scaffold routes below, e.g.:
	|   controllers.RegisterPostRoutes(r)
	|
	*/

	/* gofreight:generated-resources */
}
