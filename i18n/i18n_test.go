package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTranslate(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "en.json"), []byte(`{"hello":"Hello"}`), 0644)
	tr := New("en", "en")
	if err := tr.LoadDir(dir); err != nil {
		t.Fatal(err)
	}
	if tr.T("hello", nil) != "Hello" {
		t.Fatal("translation failed")
	}
	if tr.T("missing", nil) != "missing" {
		t.Fatal("expected key fallback")
	}
}
