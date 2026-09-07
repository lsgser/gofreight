package router

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

// RouteEntry is serializable route metadata for caching and URL generation.
type RouteEntry struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Name   string `json:"name"`
	Domain string `json:"domain,omitempty"`
}

// ResourceOptions limits which REST actions are registered.
type ResourceOptions struct {
	Only   []string
	Except []string
}

func (o ResourceOptions) allows(action string) bool {
	if len(o.Only) > 0 {
		for _, a := range o.Only {
			if a == action {
				return true
			}
		}
		return false
	}
	for _, a := range o.Except {
		if a == action {
			return false
		}
	}
	return true
}

// ExportRoutes returns serializable route metadata.
func (r *Router) ExportRoutes() []RouteEntry {
	var out []RouteEntry
	for _, route := range r.routes {
		if route.PathPrefix != "" {
			continue
		}
		out = append(out, RouteEntry{
			Method: route.Method,
			Path:   route.Path,
			Name:   route.Name,
			Domain: route.Domain,
		})
	}
	return out
}

// SaveCache writes route metadata to a JSON file for faster URL lookups.
func (r *Router) SaveCache(path string) error {
	if path == "" {
		path = filepath.Join("bootstrap", "cache", "routes.json")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(r.ExportRoutes(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0644)
}

// LoadCache merges cached route names for URL generation.
func (r *Router) LoadCache(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var entries []RouteEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name == "" {
			continue
		}
		found := false
		for _, route := range r.routes {
			if route.Name == e.Name {
				found = true
				break
			}
		}
		if !found {
			r.routes = append(r.routes, Route{
				Method: e.Method,
				Path:   e.Path,
				Name:   e.Name,
				Domain: e.Domain,
				Handler: func(w http.ResponseWriter, req *http.Request) {
					http.NotFound(w, req)
				},
			})
		}
	}
	return nil
}
