package controllers

/*
|--------------------------------------------------------------------------
| Home Controller
|--------------------------------------------------------------------------
|
| Implements Home Controller as part of the controllers package in the
| Gofreight framework. Key symbols: HomeController, RegisterHomeRoutes,
| Index.
| 
| Symbols defined here include: HomeController (exported type).
| 
*/

import (
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
)

type HomeController struct{}

func RegisterHomeRoutes(r *router.Router) {
	c := HomeController{}
	r.Get("/", controller.Handler(c.Index))
}

func (c HomeController) Index(base controller.Base) error {
	return base.RenderView("home/index", nil)
}
