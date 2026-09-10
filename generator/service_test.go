package generator

/*
|--------------------------------------------------------------------------
| Service
|--------------------------------------------------------------------------
|
| Test suite for Service in the generator package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| The generator package powers gofreight new and all make:* scaffolds.
| 
| It writes idiomatic directory layouts, GFT views, migrations, tests, and
| auth stubs from templates.
| 
| CLI handlers in cmd/gofreight call into this package; templates live
| primarily in templates.go.
| 
| Run with go test ./generator/... or go test for this package from the
| framework root.
| 
*/

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServiceGenerator(t *testing.T) {
	dir := t.TempDir()
	if err := Service(dir, "Payment"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "app", "services", "payment_service.go")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected service file: %v", err)
	}
	raw, _ := os.ReadFile(path)
	if !containsBytes(raw, []byte("PaymentService")) {
		t.Fatalf("unexpected content: %s", raw)
	}
}

func containsBytes(b, sub []byte) bool {
	return len(sub) == 0 || len(b) >= len(sub) && indexBytes(b, sub) >= 0
}

func indexBytes(b, sub []byte) int {
	for i := 0; i+len(sub) <= len(b); i++ {
		if string(b[i:i+len(sub)]) == string(sub) {
			return i
		}
	}
	return -1
}
