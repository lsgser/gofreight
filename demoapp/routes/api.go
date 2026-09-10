package routes

/*
|--------------------------------------------------------------------------
| Api
|--------------------------------------------------------------------------
|
| Implements Api as part of the routes package in the Gofreight framework.
| Key symbols: API.
| 
| Symbols defined here include: API (/*
| |--------------------------------------------------------------------------
| | API
| |--------------------------------------------------------------------------
| | | Registers JSON API routes (prefix applied by Register). | */).
| 
*/

import (
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
)

/*
|--------------------------------------------------------------------------
| API
|--------------------------------------------------------------------------
|
| Registers JSON API routes (prefix applied by Register).
|
*/
func API(r *router.Router) {
	r.Get("/health", controller.Handler(func(base controller.Base) error {
		base.RenderJSON(map[string]string{"status": "ok"})
		return nil
	}), "health")
}
