package routes

/*
|--------------------------------------------------------------------------
| Web
|--------------------------------------------------------------------------
|
| Implements Web as part of the routes package in the Gofreight framework.
| Key symbols: Web.
| 
| Symbols defined here include: Web (/*
| |--------------------------------------------------------------------------
| | Web
| |--------------------------------------------------------------------------
| | | Registers browser-facing HTTP routes. | */).
| 
*/

import (
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
	"github.com/lsgser/gofreight/version"
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
			"FrameworkVersion": version.Module(),
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
