package build_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lsgser/gofreight/build"
)

func TestModuleNameAndDefaultOutput(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/myapp\n\ngo 1.26.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	mod, err := build.ModuleName(dir)
	if err != nil {
		t.Fatal(err)
	}
	if mod != "example.com/myapp" {
		t.Fatalf("module: %q", mod)
	}

	out, err := build.DefaultOutput(dir)
	if err != nil {
		t.Fatal(err)
	}
	if out != filepath.Join("bin", "myapp") {
		t.Fatalf("output: %q", out)
	}
}

func TestValidateAppRoot(t *testing.T) {
	dir := t.TempDir()
	if err := build.ValidateAppRoot(dir); err == nil {
		t.Fatal("expected error without main.go")
	}
}

func TestBuildProducesBinary(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module buildtest\n\ngo 1.26.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := build.Build(build.Options{
		Dir:    dir,
		Output: filepath.Join(dir, "bin", "app"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(result.Output); err != nil {
		t.Fatalf("binary missing: %v", err)
	}
}
