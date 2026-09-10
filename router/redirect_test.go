package router_test

/*
|--------------------------------------------------------------------------
| Redirect
|--------------------------------------------------------------------------
|
| Test suite for Redirect in the router package.
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

func TestRedirectRoute(t *testing.T) {
	r := router.New()
	r.Redirect("/old", "/new", http.StatusMovedPermanently)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/old", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/new" {
		t.Fatalf("expected Location /new, got %q", loc)
	}
}

func TestPermanentRedirect(t *testing.T) {
	r := router.New()
	r.PermanentRedirect("/legacy", "/current")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/legacy", nil))
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rec.Code)
	}
}

func TestURLNamedRoute(t *testing.T) {
	r := router.New()
	r.Get("/posts/:id", func(w http.ResponseWriter, req *http.Request) {}, "posts.show")

	url, err := r.URL("posts.show", map[string]string{"id": "42"})
	if err != nil {
		t.Fatal(err)
	}
	if url != "/posts/42" {
		t.Fatalf("expected /posts/42, got %q", url)
	}
}

func TestURLParams(t *testing.T) {
	r := router.New()
	r.Get("/users/:id/posts/:post", func(w http.ResponseWriter, req *http.Request) {}, "users.posts.show")

	url, err := r.URLParams("users.posts.show", "id", "1", "post", "9")
	if err != nil {
		t.Fatal(err)
	}
	if url != "/users/1/posts/9" {
		t.Fatalf("unexpected url %q", url)
	}
}

func TestFallback(t *testing.T) {
	r := router.New()
	r.Fallback(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("fallback"))
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected fallback, got %d", rec.Code)
	}
}

func TestAny(t *testing.T) {
	r := router.New()
	method := ""
	r.Any("/webhook", func(w http.ResponseWriter, req *http.Request) {
		method = req.Method
	})

	for _, m := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(m, "/webhook", nil))
		if rec.Code == http.StatusNotFound {
			t.Fatalf("Any did not register %s", m)
		}
		if method != m {
			t.Fatalf("expected method %s, got %s", m, method)
		}
	}
}
