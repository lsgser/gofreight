package routes

/*
|--------------------------------------------------------------------------
| Web Routes
|--------------------------------------------------------------------------
|
| Browser-facing routes — sessions, CSRF, and GFT HTML responses.
|
*/

import (
	"blog/app/controllers"

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
	controllers.RegisterHomeRoutes(r)
	controllers.RegisterPostRoutes(r)
}
