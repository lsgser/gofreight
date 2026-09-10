package generator

/*
|--------------------------------------------------------------------------
| Graphql
|--------------------------------------------------------------------------
|
| Test suite for Graphql in the generator package.
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
	"strings"
	"testing"
)

func TestGraphQLInstaller(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "bootstrap"), 0755); err != nil {
		t.Fatal(err)
	}
	bootstrap := `package bootstrap

import "github.com/lsgser/gofreight/application"

func Application() *application.Application {
	app := application.New()
	return app
}
`
	if err := os.WriteFile(filepath.Join(dir, "bootstrap", "app.go"), []byte(bootstrap), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testapp\n\ngo 1.26.0\n\nrequire github.com/lsgser/gofreight v0.5.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := GraphQL(dir); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{
		"graphql/register.go",
		"graphql/modules.go",
		"bootstrap/app.go",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
	appBootstrap, _ := os.ReadFile(filepath.Join(dir, "bootstrap", "app.go"))
	if !strings.Contains(string(appBootstrap), "graphql.Mount(app)") {
		t.Fatalf("bootstrap should mount graphql: %s", appBootstrap)
	}
	mod, _ := os.ReadFile(filepath.Join(dir, "go.mod"))
	if !strings.Contains(string(mod), "graph-gophers/dataloader/v7") {
		t.Fatalf("go.mod should include dataloader: %s", mod)
	}
}

func TestGraphQLModuleGenerator(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "bootstrap"), 0755); err != nil {
		t.Fatal(err)
	}
	bootstrap := `package bootstrap

import "github.com/lsgser/gofreight/application"

func Application() *application.Application {
	app := application.New()
	return app
}
`
	if err := os.WriteFile(filepath.Join(dir, "bootstrap", "app.go"), []byte(bootstrap), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testapp\n\ngo 1.26.0\n\nrequire github.com/lsgser/gofreight v0.5.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := GraphQLModule(dir, "Post", map[string]string{
		"title": "string",
		"body":  "text",
	}); err != nil {
		t.Fatal(err)
	}
	modulePath := filepath.Join(dir, "graphql", "post_module.go")
	raw, err := os.ReadFile(modulePath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	for _, want := range []string{"PostModule", "createPost", "registerPostLoaders", `"title"`} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected %q in generated module", want)
		}
	}
	modules, _ := os.ReadFile(filepath.Join(dir, "graphql", "modules.go"))
	if !strings.Contains(string(modules), "PostModule()") {
		t.Fatalf("modules.go should register PostModule: %s", modules)
	}
}
