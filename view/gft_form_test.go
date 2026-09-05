package view_test

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
