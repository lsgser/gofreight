package config

import (
	"blog/app/controllers"
	"github.com/lsgser/gofreight/router"
)

func Routes(r *router.Router) {
	controllers.RegisterHomeRoutes(r)
	controllers.RegisterPostRoutes(r)
}
