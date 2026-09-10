package view_test

/*
|--------------------------------------------------------------------------
| Gft Integration
|--------------------------------------------------------------------------
|
| Test suite for Gft Integration in the view package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| The view engine compiles Gofreight Templates (.gft) to html/template,
| supports layouts, slots, partials, and RenderString for mail.
| 
| Views live under app/views; hot reload in development reloads templates
| on each request when enabled.
| 
| Run with go test ./view/... or go test for this package from the
| framework root.
| 
*/

import (
	"os"
	"strings"
	"testing"

	"github.com/lsgser/gofreight/view"
)

func TestCompileBlogIndexGFT(t *testing.T) {
	b, err := os.ReadFile("../examples/blog/app/views/posts/index.gft")
	if err != nil {
		t.Skip(err)
	}
	out, err := view.CompileGFT(string(b))
	if err != nil {
		t.Fatal(err)
	}
	sec := out.Sections["content"]
	if !strings.Contains(sec, `{{range $post := .Posts}}`) {
		t.Fatalf("expected eachor range, got:\n%s", sec)
	}
	if strings.Contains(sec, "Posts as post") {
		t.Fatalf("eachor header leaked into body: %s", sec)
	}
}
