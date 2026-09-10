package view_test

/*
|--------------------------------------------------------------------------
| Gft
|--------------------------------------------------------------------------
|
| Test suite for Gft in the view package.
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

func TestCompileGFTLayoutAndSlot(t *testing.T) {
	src := `#layout "layouts.application"

#slot "content"
<h1>Hello</h1>
#endslot`

	out, err := view.CompileGFT(src)
	if err != nil {
		t.Fatal(err)
	}
	if out.Layout != "layouts/application" {
		t.Fatalf("layout: %s", out.Layout)
	}
	if out.Sections["content"] != "<h1>Hello</h1>" {
		t.Fatalf("slot: %q", out.Sections["content"])
	}
}

func TestCompileGFTWhenEach(t *testing.T) {
	src := `#when .Posts
#each .Posts as post
  <h2>{= .Title }</h2>
#endeach
#otherwise
  <p>Empty</p>
#endwhen`

	out, err := view.CompileGFT(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.Source, `{{if .Posts}}`) {
		t.Fatalf("expected when: %s", out.Source)
	}
	if !strings.Contains(out.Source, `{{.Title}}`) {
		t.Fatalf("expected each output: %s", out.Source)
	}
}

func TestCompileGFTEachOr(t *testing.T) {
	src := `#eachor .Items as item
  <li>{= .Name }</li>
#otherwise
  <li>None</li>
#endeach`

	out, err := view.CompileGFT(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.Source, `{{else}}`) {
		t.Fatalf("expected otherwise: %s", out.Source)
	}
}

func TestCompileGFTPartialToken(t *testing.T) {
	src := `#token
#partial "partials.flash"`

	out, err := view.CompileGFT(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.Source, `authenticity_token`) {
		t.Fatal("expected token field")
	}
	if !strings.Contains(out.Source, `partials.flash`) {
		t.Fatalf("expected partial: %s", out.Source)
	}
}

func TestCompileGFTOutputSyntax(t *testing.T) {
	src := `{= .Title }
{! .HTML !}`

	out, err := view.CompileGFT(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.Source, `{{.Title}}`) {
		t.Fatalf("expected escaped output: %s", out.Source)
	}
	if !strings.Contains(out.Source, `safeHTML`) {
		t.Fatalf("expected raw output: %s", out.Source)
	}
}
