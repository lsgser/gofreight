package routes

/*
|--------------------------------------------------------------------------
| Route Registration
|--------------------------------------------------------------------------
|
| This file wires together all route groups for the application. Web routes
| (HTML) and API routes (JSON) are defined in separate files and loaded
| from here.
|
*/

import "github.com/lsgser/gofreight/router"

/*
|--------------------------------------------------------------------------
| Register
|--------------------------------------------------------------------------
|
| Loads web and API route groups onto the router.
|
*/
func Register(r *router.Router) {
	Web(r)

	r.Group(func(api *router.Router) {
		API(api)
	}).Prefix("/api/v1").Name("api.").Apply()
}
