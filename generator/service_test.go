package generator

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
