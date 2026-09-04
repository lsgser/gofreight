package config

import (
	"blog/app/controllers"
	"github.com/gofreight/gofreight/router"
)

func Routes(r *router.Router) {
	controllers.RegisterHomeRoutes(r)
	controllers.RegisterPostRoutes(r)
}
