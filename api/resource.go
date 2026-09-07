package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// Resource transforms a model into a JSON-serializable map for API responses.
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
	if lastPage == 0 {
		lastPage = 1
	}
	return map[string]any{
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    lastPage,
	}
}

// PaginatedLinks builds Laravel-style pagination links for JSON responses.
func PaginatedLinks(baseURL string, page, lastPage int) map[string]any {
	baseURL = strings.TrimRight(baseURL, "/")
	links := map[string]any{
		"first": paginatedPageURL(baseURL, 1),
		"last":  paginatedPageURL(baseURL, lastPage),
	}
	if page > 1 {
		links["prev"] = paginatedPageURL(baseURL, page-1)
	}
	if page < lastPage {
		links["next"] = paginatedPageURL(baseURL, page+1)
	}
	return links
}

func paginatedPageURL(base string, page int) string {
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + "page=" + strconv.Itoa(page)
}

// PaginatedResponse builds a full Laravel-style paginated JSON collection.
func PaginatedResponse(baseURL string, page, perPage, total int, data []map[string]any) Collection {
	lastPage := total / perPage
	if total%perPage != 0 {
		lastPage++
	}
	if lastPage == 0 {
		lastPage = 1
	}
	return Collection{
		Data:  data,
		Meta:  PaginatedMeta(page, perPage, total),
		Links: PaginatedLinks(baseURL, page, lastPage),
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
