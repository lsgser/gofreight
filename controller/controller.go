package controller

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/lsgser/gofreight/view"
)

var defaultViews *view.Engine

// SetViews configures the global view engine for controllers.
func SetViews(v *view.Engine) {
	defaultViews = v
}

// Views returns the configured view engine.
func Views() *view.Engine {
	return defaultViews
}

// Base is the foundation for all controllers, similar to ActionController::Base.
type Base struct {
	Request  *http.Request
	Response http.ResponseWriter
	Status   int
}

// NewBase creates a controller instance bound to the current request.
func NewBase(w http.ResponseWriter, r *http.Request) Base {
	return Base{Request: r, Response: w, Status: http.StatusOK}
}

// RenderJSON sends a JSON response.
func (c *Base) RenderJSON(data any) {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(c.Status)
	json.NewEncoder(c.Response).Encode(data)
}

// RenderView renders a template using the global view engine.
func (c *Base) RenderView(name string, data any) error {
	if defaultViews == nil {
		return ErrNoViewEngine
	}
	c.Response.WriteHeader(c.Status)
	if m, ok := data.(map[string]any); ok {
		data = view.MergeRequestContext(c.Request, m)
	}
	return defaultViews.Render(c.Response, name, data)
}

// RenderPartial renders a partial template (no layout).
func (c *Base) RenderPartial(name string, data any) error {
	if defaultViews == nil {
		return ErrNoViewEngine
	}
	return defaultViews.Partial(c.Response, name, data)
}

// RenderHTML renders an HTML template with the given data.
func (c *Base) RenderHTML(tmpl *template.Template, name string, data any) error {
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.WriteHeader(c.Status)
	return tmpl.ExecuteTemplate(c.Response, name, data)
}

// RenderText sends a plain text response.
func (c *Base) RenderText(text string) {
	c.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Response.WriteHeader(c.Status)
	c.Response.Write([]byte(text))
}

// Redirect sends an HTTP redirect.
func (c *Base) Redirect(url string, code int) {
	if code == 0 {
		code = http.StatusFound
	}
	http.Redirect(c.Response, c.Request, url, code)
}

// Param returns a route parameter (e.g. :id from /posts/:id).
func (c *Base) Param(key string) string {
	return c.Request.PathValue(key)
}

// Query returns a query string parameter.
func (c *Base) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// NotFound renders a 404 response.
func (c *Base) NotFound(message string) {
	c.Status = http.StatusNotFound
	c.RenderJSON(map[string]string{"error": message})
}

// Unprocessable renders a 422 response with validation errors.
func (c *Base) Unprocessable(errors map[string][]string) {
	c.Status = http.StatusUnprocessableEntity
	c.RenderJSON(map[string]any{"errors": errors})
}

// Unauthorized renders a 401 response.
func (c *Base) Unauthorized(message string) {
	c.Status = http.StatusUnauthorized
	c.RenderJSON(map[string]string{"error": message})
}

// Handler wraps a controller action method into an http.HandlerFunc.
func Handler(action func(Base) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base := NewBase(w, r)
		if err := action(base); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// ErrNoViewEngine is returned when RenderView is called without a view engine.
var ErrNoViewEngine = &viewEngineError{}

type viewEngineError struct{}

func (e *viewEngineError) Error() string {
	return "view engine not configured — call controller.SetViews()"
}
