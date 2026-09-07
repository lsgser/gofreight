package controller_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lsgser/gofreight/controller"
)

func TestStatusHelpers(t *testing.T) {
	t.Run("created", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Accept", "application/json")
		base := controller.NewBase(rec, req)
		base.Created(map[string]string{"id": "1"})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("no_content", func(t *testing.T) {
		rec := httptest.NewRecorder()
		base := controller.NewBase(rec, httptest.NewRequest(http.MethodDelete, "/", nil))
		base.NoContent()
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
		if rec.Body.Len() != 0 {
			t.Fatalf("expected empty body")
		}
	})

	t.Run("head", func(t *testing.T) {
		rec := httptest.NewRecorder()
		base := controller.NewBase(rec, httptest.NewRequest(http.MethodHead, "/", nil))
		base.Head(http.StatusOK)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("abort json", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept", "application/json")
		base := controller.NewBase(rec, req)
		base.Abort(http.StatusNotFound, "missing")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("abort_named", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept", "application/json")
		base := controller.NewBase(rec, req)
		if err := base.AbortNamed("forbidden", "nope"); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("status_from_name", func(t *testing.T) {
		code, ok := controller.StatusFromName("no-content")
		if !ok || code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d ok=%v", code, ok)
		}
	})
}

func TestRedirectNamed(t *testing.T) {
	rec := httptest.NewRecorder()
	base := controller.NewBase(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if err := base.RedirectNamed("/home", "see_other"); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rec.Code)
	}
}
