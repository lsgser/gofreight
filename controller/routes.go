package controller

/*
|--------------------------------------------------------------------------
| Routes
|--------------------------------------------------------------------------
|
| Implements Routes as part of the controller package in the Gofreight
| framework. Key symbols: SetRouter, Router, RouteURL, RedirectBack,
| RedirectRoute.
| 
| Controllers wrap http.HandlerFunc with a Base struct that exposes
| Request, Response, validation, views, and JSON helpers.
| 
| This package defines RenderView, Redirect, status helpers, file
| downloads, and route registration utilities.
| 
| Application controllers embed these patterns; see docs/controllers.md
| for request lifecycle.
| 
| Symbols defined here include: SetRouter (SetRouter configures the
| application router for named-route URL helpers.); Router (Router returns
| the configured application router.); RouteURL (RouteURL builds a URL for
| a named route.); RedirectBack (RedirectBack redirects to the Referer
| header, or fallback when absent.); RedirectRoute (RedirectRoute
| redirects to a named route.).
| 
*/

import (
	"fmt"
	"net/http"

	"github.com/lsgser/gofreight/router"
)

var defaultRouter *router.Router

// SetRouter configures the application router for named-route URL helpers.
func SetRouter(r *router.Router) {
	defaultRouter = r
}

// Router returns the configured application router.
func Router() *router.Router {
	return defaultRouter
}

// RouteURL builds a URL for a named route.
func RouteURL(name string, params map[string]string) (string, error) {
	if defaultRouter == nil {
		return "", fmt.Errorf("router not configured — call controller.SetRouter in bootstrap")
	}
	return defaultRouter.URL(name, params)
}

// RedirectBack redirects to the Referer header, or fallback when absent.
func (c *Base) RedirectBack(fallback string, code int) {
	target := c.Request.Header.Get("Referer")
	if target == "" {
		target = fallback
	}
	if target == "" {
		target = "/"
	}
	if code == 0 {
		code = http.StatusFound
	}
	c.Redirect(target, code)
}

// RedirectRoute redirects to a named route.
func (c *Base) RedirectRoute(name string, params map[string]string, code int) error {
	url, err := RouteURL(name, params)
	if err != nil {
		return err
	}
	if code == 0 {
		code = http.StatusFound
	}
	c.Redirect(url, code)
	return nil
}
