package routes

/*
|--------------------------------------------------------------------------
| Route Registration
|--------------------------------------------------------------------------
|
| Wires web (HTML) and API (JSON) route groups for this application.
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
