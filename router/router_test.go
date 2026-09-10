package router_test

/*
|--------------------------------------------------------------------------
| Router
|--------------------------------------------------------------------------
|
| Test suite for Router in the router package.
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

func TestRouterGet(t *testing.T) {
	r := router.New()
	r.Get("/hello", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Hello"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "Hello" {
		t.Fatalf("expected Hello, got %s", rec.Body.String())
	}
}

func TestRouterParams(t *testing.T) {
	r := router.New()
	r.Get("/posts/:id", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(req.PathValue("id")))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/42", nil)
	r.ServeHTTP(rec, req)

	if rec.Body.String() != "42" {
		t.Fatalf("expected 42, got %s", rec.Body.String())
	}
}

func TestResources(t *testing.T) {
	r := router.New()
	var called string
	r.Resources("posts", router.ResourceHandlers{
		Index: func(w http.ResponseWriter, req *http.Request) { called = "index" },
		Show:  func(w http.ResponseWriter, req *http.Request) { called = "show" },
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts", nil))
	if called != "index" {
		t.Fatalf("expected index, got %s", called)
	}

	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts/1", nil))
	if called != "show" {
		t.Fatalf("expected show, got %s", called)
	}
}

func TestMount(t *testing.T) {
	r := router.New()
	called := false
	r.Mount("/assets", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/assets/app.css", nil)
	r.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected prefix mount to match")
	}
}
