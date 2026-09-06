package routes

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
|
| Stateless JSON API routes. Prefix /api/v1 is applied in routes/register.go.
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
