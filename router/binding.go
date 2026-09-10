package router

/*
|--------------------------------------------------------------------------
| Binding
|--------------------------------------------------------------------------
|
| Implements Binding as part of the router package in the Gofreight
| framework. Key symbols: Bound, BoundAs, BindModelID, BindModelColumn,
| BindModel, BindModelBy.
| 
| The router matches verbs and paths, supports groups, prefixes, named
| routes, constraints, signed URLs, and domain routing.
| 
| Routes register in routes/web.go and routes/api.go; see docs/routing.md
| for middleware and model binding.
| 
| Symbols defined here include: Bound (Bound returns a model bound to the
| request by route model binding.); BoundAs (BoundAs returns a typed bound
| model.); BindModelID (BindModelID binds a route parameter to a model
| looked up by numeric ID.); BindModelColumn (BindModelColumn binds a
| route parameter to a model looked up by any column.); BindModel
| (BindModel attaches implicit route model binding by numeric ID.);
| BindModelBy (BindModelBy attaches implicit route model binding by a
| custom column.); Bind (Bind registers a custom binding function for a
| route parameter.).
| 
*/

import (
	"context"
	"net/http"
	"strconv"

	"github.com/lsgser/gofreight/model"
)

type boundModelsKey struct{}

// Bound returns a model bound to the request by route model binding.
func Bound(r *http.Request, name string) (any, bool) {
	m, _ := r.Context().Value(boundModelsKey{}).(map[string]any)
	if m == nil {
		return nil, false
	}
	v, ok := m[name]
	return v, ok
}

// BoundAs returns a typed bound model.
func BoundAs[T any](r *http.Request, name string) (*T, bool) {
	v, ok := Bound(r, name)
	if !ok {
		return nil, false
	}
	t, ok := v.(*T)
	return t, ok
}

func withBound(r *http.Request, name string, value any) *http.Request {
	m, _ := r.Context().Value(boundModelsKey{}).(map[string]any)
	next := make(map[string]any, len(m)+1)
	for k, v := range m {
		next[k] = v
	}
	next[name] = value
	return r.WithContext(context.WithValue(r.Context(), boundModelsKey{}, next))
}

// BindModelID binds a route parameter to a model looked up by numeric ID.
func BindModelID[T any](repo *model.Repository[T], param string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := strconv.ParseInt(r.PathValue(param), 10, 64)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			record, err := repo.Find(r.Context(), id)
			if err != nil || record == nil {
				http.NotFound(w, r)
				return
			}
			next.ServeHTTP(w, withBound(r, param, record))
		})
	}
}

// BindModelColumn binds a route parameter to a model looked up by any column.
func BindModelColumn[T any](repo *model.Repository[T], param, column string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value := r.PathValue(param)
			if value == "" {
				http.NotFound(w, r)
				return
			}
			record, err := repo.FindBy(r.Context(), column, value)
			if err != nil || record == nil {
				http.NotFound(w, r)
				return
			}
			next.ServeHTTP(w, withBound(r, param, record))
		})
	}
}

// BindModel attaches implicit route model binding by numeric ID.
func BindModel[T any](reg *RouteRegistrar, repo *model.Repository[T], param string) *RouteRegistrar {
	return reg.Use(BindModelID(repo, param))
}

// BindModelBy attaches implicit route model binding by a custom column.
func BindModelBy[T any](reg *RouteRegistrar, repo *model.Repository[T], param, column string) *RouteRegistrar {
	return reg.Use(BindModelColumn(repo, param, column))
}

// Bind registers a custom binding function for a route parameter.
func (reg *RouteRegistrar) Bind(param string, bind func(*http.Request) (any, error)) *RouteRegistrar {
	return reg.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value, err := bind(r)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			next.ServeHTTP(w, withBound(r, param, value))
		})
	})
}
