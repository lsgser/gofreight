package router_test

/*
|--------------------------------------------------------------------------
| Group
|--------------------------------------------------------------------------
|
| Test suite for Group in the router package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| The router matches verbs and paths, supports groups, prefixes, named
| routes, constraints, signed URLs, and domain routing.
| 
| Routes register in routes/web.go and routes/api.go; see docs/routing.md
| for middleware and model binding.
| 
| Run with go test ./router/... or go test for this package from the
| framework root.
| 
*/

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lsgser/gofreight/router"
)

func TestRouteGroup(t *testing.T) {
	r := router.New()
	r.Group(func(api *router.Router) {
		api.Get("/users", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("users"))
		})
	}).Prefix("/api/v1").Name("api.").Apply()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "users" {
		t.Fatalf("got %q", rec.Body.String())
	}
}

func TestNestedGroups(t *testing.T) {
	r := router.New()
	r.Group(func(api *router.Router) {
		api.Group(func(v1 *router.Router) {
			v1.Get("/users", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("ok"))
			})
		}).Prefix("/v1").Apply()
	}).Prefix("/api").Apply()

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/users", nil))
	if rec.Body.String() != "ok" {
		t.Fatalf("got %q", rec.Body.String())
	}
}

func TestGroupMiddleware(t *testing.T) {
	r := router.New()
	order := ""
	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			order += "auth"
			next.ServeHTTP(w, req)
		})
	}
	r.Group(func(api *router.Router) {
		api.Get("/secure", func(w http.ResponseWriter, req *http.Request) {
			order += "handler"
			w.WriteHeader(http.StatusOK)
		})
	}).Prefix("/api").Use(auth).Apply()

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/secure", nil))
	if order != "authhandler" {
		t.Fatalf("got order %q", order)
	}
}

func TestApiResource(t *testing.T) {
	r := router.New()
	called := ""
	r.Group(func(api *router.Router) {
		api.ApiResource("posts", router.ApiResourceHandlers{
			Index: func(w http.ResponseWriter, req *http.Request) { called = "index" },
			Show:  func(w http.ResponseWriter, req *http.Request) { called = "show" },
		})
	}).Prefix("/api/v1").Apply()

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil))
	if called != "index" {
		t.Fatalf("expected index, got %q", called)
	}

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/posts/5", nil))
	if called != "show" {
		t.Fatalf("expected show, got %q", called)
	}

	// no /posts/new for API resource
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/posts/new", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for new, got %d", rec.Code)
	}
}

func TestRouteMiddleware(t *testing.T) {
	r := router.New()
	order := ""
	groupMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			order += "group"
			next.ServeHTTP(w, req)
		})
	}
	routeMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			order += "route"
			next.ServeHTTP(w, req)
		})
	}
	r.Group(func(api *router.Router) {
		api.Get("/posts", func(w http.ResponseWriter, req *http.Request) {
			order += "handler"
			w.WriteHeader(http.StatusOK)
		}).Use(routeMw)
	}).Prefix("/api").Use(groupMw).Apply()

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/posts", nil))
	if order != "grouproutehandler" {
		t.Fatalf("got order %q", order)
	}
}

func TestGroupPrefixLegacy(t *testing.T) {
	r := router.New()
	r.GroupPrefix("/api/v1", func(api *router.Router) {
		api.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("ok"))
		})
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	if rec.Body.String() != "ok" {
		t.Fatalf("got %q", rec.Body.String())
	}
}
