package api

import (
	"encoding/json"
	"net/http"
)

// Resource transforms a model into a JSON-serializable map (Laravel API Resource).
type Resource interface {
	ToMap() map[string]any
}

// ResourceFunc adapts a function to Resource.
type ResourceFunc func() map[string]any

func (f ResourceFunc) ToMap() map[string]any { return f() }

// Collection wraps multiple resources for JSON output.
type Collection struct {
	Data  []map[string]any `json:"data"`
	Meta  map[string]any   `json:"meta,omitempty"`
	Links map[string]any   `json:"links,omitempty"`
}

// PaginatedMeta builds standard pagination meta.
func PaginatedMeta(page, perPage, total int) map[string]any {
	lastPage := total / perPage
	if total%perPage != 0 {
		lastPage++
	}
	return map[string]any{
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    lastPage,
	}
}

// Render sends a single resource as JSON.
func Render(w http.ResponseWriter, status int, r Resource) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"data": r.ToMap()})
}

// RenderCollection sends a collection as JSON.
func RenderCollection(w http.ResponseWriter, status int, col Collection) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(col)
}

// RenderMany maps resources and renders a collection.
func RenderMany(w http.ResponseWriter, status int, items []Resource) {
	data := make([]map[string]any, len(items))
	for i, item := range items {
		data[i] = item.ToMap()
	}
	RenderCollection(w, status, Collection{Data: data})
}

// Error sends a JSON error response.
func Error(w http.ResponseWriter, status int, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body := map[string]any{"error": message}
	for k, v := range details {
		body[k] = v
	}
	json.NewEncoder(w).Encode(body)
}

// VersionPrefix returns a path prefix like /api/v1.
func VersionPrefix(version string) string {
	if version == "" {
		version = "v1"
	}
	return "/api/" + version
}
