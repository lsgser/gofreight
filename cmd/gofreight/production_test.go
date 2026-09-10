package main

/*
|--------------------------------------------------------------------------
| Production
|--------------------------------------------------------------------------
|
| Test suite for Production in the gofreight package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| This directory contains the gofreight CLI binary: command registration,
| terminal UI, and handlers for make:*, migrate, serve, test, and
| mail:preview.
| 
| Each subcommand lives in its own source file; commands.go registers the
| catalog shown by gofreight list.
| 
| Install locally with go install ./cmd/gofreight from the framework
| repository root.
| 
| Run with go test ./cmd/gofreight/... or go test for this package from
| the framework root.
| 
*/

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/lsgser/gofreight/config"
)

func TestNeedsProductionGuard(t *testing.T) {
	guarded := []string{
		"migrate", "migrate:fresh", "db:wipe", "db:seed", "down", "queue:clear", "make:model", "key:generate",
	}
	for _, name := range guarded {
		if !needsProductionGuard(name) {
			t.Fatalf("expected %q to require production guard", name)
		}
	}

	safe := []string{"list", "help", "migrate:status", "db:show", "queue:failed", "config:show", "about"}
	for _, name := range safe {
		if needsProductionGuard(name) {
			t.Fatalf("expected %q to skip production guard", name)
		}
	}
}

func TestGuardProductionSkipsOutsideProduction(t *testing.T) {
	t.Setenv("GOFREIGHT_ENV", "development")
	if !guardProduction("migrate", nil) {
		t.Fatal("expected guard to allow command in development")
	}
}

func TestGuardProductionSkipsWithForce(t *testing.T) {
	t.Setenv("GOFREIGHT_ENV", "production")
	if !guardProduction("migrate", []string{"--force"}) {
		t.Fatal("expected --force to bypass production guard")
	}
}

func TestGuardProductionRequiresYes(t *testing.T) {
	t.Setenv("GOFREIGHT_ENV", "production")
	t.Setenv("APP_NAME", "demo")
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("DB_DATABASE", "demo.sqlite3")

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	go func() {
		_, _ = w.WriteString("yes\n")
		_ = w.Close()
	}()

	if !guardProduction("migrate", nil) {
		t.Fatal("expected confirmation to allow command")
	}
}

func TestGuardProductionAbortsWithoutYes(t *testing.T) {
	t.Setenv("GOFREIGHT_ENV", "production")

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	go func() {
		_, _ = w.WriteString("no\n")
		_ = w.Close()
	}()

	if guardProduction("migrate", nil) {
		t.Fatal("expected guard to reject non-yes response")
	}
}

func TestPrintProductionWarning(t *testing.T) {
	cfg := configForTest()
	var buf bytes.Buffer
	printProductionWarning(&buf, "migrate:fresh", cfg)
	out := buf.String()
	for _, part := range []string{"PRODUCTION ENVIRONMENT", "migrate:fresh", "sqlite/demo.sqlite3", "Type \"yes\" to continue"} {
		if !strings.Contains(out, part) {
			t.Fatalf("expected %q in output:\n%s", part, out)
		}
	}
}

func configForTest() *config.Config {
	loadEnv()
	cfg := config.Load()
	cfg.Environment = "production"
	cfg.Database.Connection = "sqlite"
	cfg.Database.Database = "demo.sqlite3"
	return cfg
}
