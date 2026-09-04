package controllers

import (
	"github.com/gofreight/gofreight/controller"
	"github.com/gofreight/gofreight/router"
)

type HomeController struct{}

func RegisterHomeRoutes(r *router.Router) {
	c := HomeController{}
	r.Get("/", controller.Handler(c.Index))
}

func (c HomeController) Index(base controller.Base) error {
	return base.RenderView("home/index", nil)
}
