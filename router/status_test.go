package router_test

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
