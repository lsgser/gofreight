package view_test

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lsgser/gofreight/view"
)

func TestRenderLayoutExecutesTitleSlot(t *testing.T) {
	dir := t.TempDir()
	layoutDir := filepath.Join(dir, "layouts")
	homeDir := filepath.Join(dir, "home")
	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	layout := `<!DOCTYPE html>
<html>
<head><title>#place "title" "Fallback"</title></head>
<body>#place "content"</body>
</html>`
	index := `#layout "layouts.application"

#slot "title"
{= .Name } — Welcome
#endslot

#slot "content"
<h1>{= .Name }</h1>
#endslot`

	if err := os.WriteFile(filepath.Join(layoutDir, "application.gft"), []byte(layout), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(homeDir, "index.gft"), []byte(index), 0644); err != nil {
		t.Fatal(err)
	}

	engine := view.New(dir)
	if err := engine.Load(); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	if err := engine.Render(rec, "home/index", map[string]any{"Name": "demoapp"}); err != nil {
		t.Fatal(err)
	}

	body := rec.Body.String()
	if strings.Contains(body, "{{.Name}}") {
		t.Fatalf("title slot was not executed: %s", body)
	}
	if !strings.Contains(body, "<title>demoapp — Welcome</title>") {
		t.Fatalf("expected rendered title, got:\n%s", body)
	}
	if !strings.Contains(body, "<h1>demoapp</h1>") {
		t.Fatalf("expected rendered content, got:\n%s", body)
	}
}

func TestRenderStringUsesLayout(t *testing.T) {
	dir := t.TempDir()
	layoutDir := filepath.Join(dir, "layouts")
	mailDir := filepath.Join(dir, "mail")
	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatal(err)
	}

	layout := `<html><body>#place "content"</body></html>`
	viewSrc := `#layout "layouts.wrapper"

#slot "content"
<span>{= .Tag }</span>
#endslot`

	if err := os.WriteFile(filepath.Join(layoutDir, "wrapper.gft"), []byte(layout), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mailDir, "ping.gft"), []byte(viewSrc), 0644); err != nil {
		t.Fatal(err)
	}

	engine := view.New(dir)
	if err := engine.Load(); err != nil {
		t.Fatal(err)
	}

	html, err := engine.RenderString("mail/ping", map[string]any{"Tag": "pong"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "<span>pong</span>") {
		t.Fatalf("expected rendered mail string, got:\n%s", html)
	}
}
