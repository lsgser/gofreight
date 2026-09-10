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
