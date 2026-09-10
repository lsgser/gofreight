package routes

/*
|--------------------------------------------------------------------------
| Register
|--------------------------------------------------------------------------
|
| Implements Register as part of the routes package in the Gofreight
| framework. Key symbols: Register.
| 
| Symbols defined here include: Register (/*
| |--------------------------------------------------------------------------
| | Register
| |--------------------------------------------------------------------------
| | | Loads web and API route groups onto the router. | */).
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
