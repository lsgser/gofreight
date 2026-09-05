package view

import (
	"net/http"

	"github.com/lsgser/gofreight/middleware"
	"github.com/lsgser/gofreight/validation"
)

// MergeRequestContext injects CSRF token, flash, validation errors, and old input into view data.
func MergeRequestContext(r *http.Request, data map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range data {
		out[k] = v
	}

	session := middleware.SessionFromContext(r.Context())
	if session == nil {
		return out
	}

	if _, ok := out["CSRFToken"]; !ok {
		out["CSRFToken"] = middleware.CSRFTokenFromSession(session)
	}
	if _, ok := out["Errors"]; !ok {
		out["Errors"] = session.PullValidationErrors()
	}
	if _, ok := out["Old"]; !ok {
		out["Old"] = session.PullOldInput()
	}
	if _, ok := out["Flash"]; !ok {
		flashes := session.Flashes()
		if msg, ok := flashes["message"]; ok {
			out["Flash"] = msg
		}
	}
	return out
}

// Old returns a repopulated field value (old input first, then fallback).
func Old(field string, ctx map[string]any, fallback string) string {
	if old, ok := ctx["Old"].(map[string]string); ok {
		if v, ok := old[field]; ok && v != "" {
			return v
		}
	}
	return fallback
}

// FieldErrors returns validation messages for a field.
func FieldErrors(field string, ctx map[string]any) []string {
	errs, ok := ctx["Errors"].(validation.Errors)
	if !ok || errs == nil {
		return nil
	}
	return errs[field]
}

// HasError reports whether a field has validation errors.
func HasError(field string, ctx map[string]any) bool {
	return len(FieldErrors(field, ctx)) > 0
}
