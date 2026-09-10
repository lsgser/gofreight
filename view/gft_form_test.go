package view_test

/*
|--------------------------------------------------------------------------
| Gft Form
|--------------------------------------------------------------------------
|
| Test suite for Gft Form in the view package.
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
	"strings"
	"testing"

	"github.com/lsgser/gofreight/view"
)

func TestCompileGFTFormField(t *testing.T) {
	src := `#form action="/posts" method="POST"
#field "title" label="Title" type="text" value=".Item.Title"
<button>Save</button>
#endform`

	out, err := view.CompileGFT(src)
	if err != nil {
		t.Fatal(err)
	}
	body := out.Source
	if !strings.Contains(body, `name="authenticity_token"`) {
		t.Fatalf("missing csrf token: %s", body)
	}
	if !strings.Contains(body, `name="title"`) {
		t.Fatalf("missing title field: %s", body)
	}
	if !strings.Contains(body, `fieldErrors "title"`) {
		t.Fatalf("missing field errors: %s", body)
	}
}

func TestCompileGFTFormMethodSpoof(t *testing.T) {
	src := `#form action="/posts/1" method="PUT"
#field "title" label="Title"
#endform`

	out, err := view.CompileGFT(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.Source, `name="_method" value="PUT"`) {
		t.Fatalf("missing method spoof: %s", out.Source)
	}
}
