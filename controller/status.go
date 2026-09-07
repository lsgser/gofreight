package controller

import (
	"fmt"
	"net/http"
	"strings"
)

// Named HTTP status codes (Rails-style symbols).
var namedStatuses = map[string]int{
	"ok":                     http.StatusOK,
	"created":                http.StatusCreated,
	"accepted":               http.StatusAccepted,
	"no_content":             http.StatusNoContent,
	"reset_content":          http.StatusResetContent,
	"partial_content":        http.StatusPartialContent,
	"moved_permanently":      http.StatusMovedPermanently,
	"found":                  http.StatusFound,
	"see_other":              http.StatusSeeOther,
	"not_modified":           http.StatusNotModified,
	"bad_request":            http.StatusBadRequest,
	"unauthorized":           http.StatusUnauthorized,
	"payment_required":       http.StatusPaymentRequired,
	"forbidden":              http.StatusForbidden,
	"not_found":              http.StatusNotFound,
	"method_not_allowed":     http.StatusMethodNotAllowed,
	"not_acceptable":         http.StatusNotAcceptable,
	"conflict":               http.StatusConflict,
	"gone":                   http.StatusGone,
	"unprocessable_entity":   http.StatusUnprocessableEntity,
	"too_many_requests":      http.StatusTooManyRequests,
	"internal_server_error":  http.StatusInternalServerError,
	"not_implemented":        http.StatusNotImplemented,
	"bad_gateway":            http.StatusBadGateway,
	"service_unavailable":    http.StatusServiceUnavailable,
	"gateway_timeout":        http.StatusGatewayTimeout,
}

// StatusFromName resolves a Rails-style status name to an HTTP code.
//
//	StatusFromName("created")  // 201
//	StatusFromName("no-content") // 204
func StatusFromName(name string) (int, bool) {
	key := strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	code, ok := namedStatuses[key]
	return code, ok
}

// SetStatus sets the response status used by Render* helpers.
func (c *Base) SetStatus(code int) {
	c.Status = code
}

// Head sends a response with only a status line (Rails head :no_content).
func (c *Base) Head(code int) {
	if code == 0 {
		code = http.StatusOK
	}
	c.Status = code
	c.Response.WriteHeader(code)
}

// OK sends a 200 JSON response.
func (c *Base) OK(data any) {
	c.Status = http.StatusOK
	c.RenderJSON(data)
}

// Created sends a 201 JSON response (Rails render status: :created).
func (c *Base) Created(data any) {
	c.Status = http.StatusCreated
	if data == nil {
		c.Head(http.StatusCreated)
		return
	}
	c.RenderJSON(data)
}

// Accepted sends a 202 JSON response.
func (c *Base) Accepted(data any) {
	c.Status = http.StatusAccepted
	if data == nil {
		c.Head(http.StatusAccepted)
		return
	}
	c.RenderJSON(data)
}

// NoContent sends a 204 response with no body (Rails head :no_content).
func (c *Base) NoContent() {
	c.Head(http.StatusNoContent)
}

// Forbidden sends a 403 response.
func (c *Base) Forbidden(message string) {
	c.Abort(http.StatusForbidden, message)
}

// Gone sends a 410 response.
func (c *Base) Gone(message string) {
	c.Abort(http.StatusGone, message)
}

// Conflict sends a 409 response.
func (c *Base) Conflict(message string) {
	c.Abort(http.StatusConflict, message)
}

// TooManyRequests sends a 429 response.
func (c *Base) TooManyRequests(message string) {
	c.Abort(http.StatusTooManyRequests, message)
}

// Abort stops the request with an HTTP error (Laravel abort()).
func (c *Base) Abort(code int, message string) {
	if message == "" {
		message = http.StatusText(code)
	}
	if c.WantsJSON() {
		c.Status = code
		c.RenderJSON(map[string]string{"error": message})
		return
	}
	http.Error(c.Response, message, code)
}

// AbortNamed aborts using a Rails-style status name.
func (c *Base) AbortNamed(name, message string) error {
	code, ok := StatusFromName(name)
	if !ok {
		return fmt.Errorf("unknown HTTP status %q", name)
	}
	c.Abort(code, message)
	return nil
}

// AbortIf aborts when condition is true.
func (c *Base) AbortIf(condition bool, code int, message string) {
	if condition {
		c.Abort(code, message)
	}
}

// AbortUnless aborts when condition is false.
func (c *Base) AbortUnless(condition bool, code int, message string) {
	if !condition {
		c.Abort(code, message)
	}
}

// RedirectNamed redirects using a Rails-style status name.
func (c *Base) RedirectNamed(url, statusName string) error {
	code, ok := StatusFromName(statusName)
	if !ok {
		return fmt.Errorf("unknown HTTP status %q", statusName)
	}
	c.Redirect(url, code)
	return nil
}
