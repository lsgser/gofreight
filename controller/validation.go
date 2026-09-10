package controller

/*
|--------------------------------------------------------------------------
| Validation
|--------------------------------------------------------------------------
|
| Implements Validation as part of the controller package in the Gofreight
| framework. Key symbols: ValidateUsing, HandleValidationFailure,
| WantsJSON, RedirectBackWithErrors, ViewData.
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
| Symbols defined here include: ErrValidationFailed (exported value);
| ValidateUsing (ValidateUsing validates the request with a Vine schema.);
| HandleValidationFailure (HandleValidationFailure stores errors for HTML
| flows or returns JSON 422.); WantsJSON (WantsJSON returns true when the
| client expects a JSON response.); RedirectBackWithErrors
| (RedirectBackWithErrors flashes validation errors and old input, then
| redirects to Referer.); ViewData (ViewData merges request-scoped view
| context (CSRF, errors, old input) with extra data.).
| 
*/

import (
	"errors"
	"net/http"
	"strings"

	"github.com/lsgser/gofreight/middleware"
	"github.com/lsgser/gofreight/request"
	"github.com/lsgser/gofreight/validation"
	"github.com/lsgser/gofreight/view"
	"github.com/lsgser/gofreight/vine"
)

// ErrValidationFailed is returned when ValidateUsing fails.
var ErrValidationFailed = errors.New("validation failed")

// ValidateUsing validates the request with a Vine schema.
// On failure, responds with JSON 422 or redirects back with errors and old input.
func (c *Base) ValidateUsing(schema *vine.ObjectSchema) (map[string]string, error) {
	data, errs, err := request.ValidateUsing(c.Request, schema)
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		c.HandleValidationFailure(errs, data)
		return nil, ErrValidationFailed
	}
	return data, nil
}

// HandleValidationFailure stores errors for HTML flows or returns JSON 422.
func (c *Base) HandleValidationFailure(errs validation.Errors, old map[string]string) {
	if c.WantsJSON() {
		c.Unprocessable(errs)
		return
	}
	c.RedirectBackWithErrors(errs, old)
}

// WantsJSON returns true when the client expects a JSON response.
func (c *Base) WantsJSON() bool {
	accept := c.Request.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		return false
	}
	return strings.Contains(accept, "application/json") ||
		c.Request.Header.Get("Content-Type") == "application/json"
}

// RedirectBackWithErrors flashes validation errors and old input, then redirects to Referer.
func (c *Base) RedirectBackWithErrors(errs validation.Errors, old map[string]string) {
	session := middleware.SessionFromContext(c.Request.Context())
	if session != nil {
		session.FlashValidationErrors(errs)
		session.FlashOldInput(old)
	}
	target := c.Request.Header.Get("Referer")
	if target == "" {
		target = "/"
	}
	c.Redirect(target, http.StatusSeeOther)
}

// ViewData merges request-scoped view context (CSRF, errors, old input) with extra data.
func (c *Base) ViewData(extra map[string]any) map[string]any {
	if extra == nil {
		extra = map[string]any{}
	}
	return view.MergeRequestContext(c.Request, extra)
}
