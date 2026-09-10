package router_test

/*
|--------------------------------------------------------------------------
| Status
|--------------------------------------------------------------------------
|
| Test suite for Status in the router package.
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

func TestRouteStatus(t *testing.T) {
	r := router.New()
	r.Get("/gone", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("should be overridden"))
	}).Status(http.StatusGone)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/gone", nil))
	if rec.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d", rec.Code)
	}
}

func TestRouteStatusName(t *testing.T) {
	r := router.New()
	r.Get("/ping", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("pong"))
	}).StatusName("no_content")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
