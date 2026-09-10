package i18n

/*
|--------------------------------------------------------------------------
| I18n
|--------------------------------------------------------------------------
|
| Test suite for I18n in the i18n package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| The i18n package loads JSON locale files from config/locales and
| resolves translation keys in views and controllers.
| 
| Middleware can set locale from session or Accept-Language; helpers
| mirror Laravel-style __() usage in GFT.
| 
| Run with go test ./i18n/... or go test for this package from the
| framework root.
| 
*/

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
