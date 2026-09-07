package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

// ExceptionRenderer renders an error view by name.
type ExceptionRenderer func(w http.ResponseWriter, view string, data map[string]any) error

// ExceptionHandler renders consistent error pages and JSON errors.
type ExceptionHandler struct {
	Render        ExceptionRenderer
	ShowDebug     bool
	NotFoundView  string
	ServerErrView string
}

// NewExceptionHandler creates production-friendly error handling middleware.
func NewExceptionHandler(render ExceptionRenderer, debug bool) *ExceptionHandler {
	return &ExceptionHandler{
		Render:        render,
		ShowDebug:     debug,
		NotFoundView:  "errors/404",
		ServerErrView: "errors/500",
	}
}

// Middleware catches panics and renders error responses.
func (e *ExceptionHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v\n%s", err, debug.Stack())
				e.renderError(w, r, http.StatusInternalServerError, "Internal Server Error", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (e *ExceptionHandler) renderError(w http.ResponseWriter, r *http.Request, code int, message string, detail any) {
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		payload := map[string]any{"error": message}
		if e.ShowDebug && detail != nil {
			payload["debug"] = fmt.Sprint(detail)
		}
		json.NewEncoder(w).Encode(payload)
		return
	}
	if e.Render != nil {
		viewName := e.ServerErrView
		if code == http.StatusNotFound {
			viewName = e.NotFoundView
		}
		w.WriteHeader(code)
		if err := e.Render(w, viewName, map[string]any{"Message": message}); err == nil {
			return
		}
	}
	http.Error(w, message, code)
}

func wantsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return accept == "application/json" || accept == "*/*" && r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}
