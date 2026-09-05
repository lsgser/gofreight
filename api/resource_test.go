package api

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
