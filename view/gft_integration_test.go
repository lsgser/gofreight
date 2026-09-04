package view_test

import (
	"os"
	"strings"
	"testing"

	"github.com/gofreight/gofreight/view"
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
