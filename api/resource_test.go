package api

/*
|--------------------------------------------------------------------------
| Resource
|--------------------------------------------------------------------------
|
| Test suite for Resource in the api package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| The api package implements JSON resource transformers and helpers for
| versioned HTTP APIs.
| 
| Resources map models and structs to consistent JSON shapes, pagination
| metadata, and optional link collections.
| 
| Use api.Group with the router to register /api/v1-style routes with
| shared middleware.
| 
| Run with go test ./api/... or go test for this package from the
| framework root.
| 
*/

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type sampleResource struct{ id int64 }

func (s sampleResource) ToMap() map[string]any {
	return map[string]any{"id": s.id}
}

func TestRender(t *testing.T) {
	w := httptest.NewRecorder()
	Render(w, http.StatusOK, sampleResource{id: 1})
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	data := body["data"].(map[string]any)
	if data["id"].(float64) != 1 {
		t.Fatalf("unexpected body %v", body)
	}
}

func TestVersionPrefix(t *testing.T) {
	if VersionPrefix("v2") != "/api/v2" {
		t.Fatal("prefix mismatch")
	}
}
