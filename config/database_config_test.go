package config_test

/*
|--------------------------------------------------------------------------
| Database Config
|--------------------------------------------------------------------------
|
| Test suite for Database Config in the config package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| The config package loads .env files, resolves MAIL_DRIVER and database
| URLs, and reads config/app.yaml.
| 
| Database drivers and app keys are validated early so misconfiguration
| fails fast at boot.
| 
| Application code reads config through helpers rather than os.Getenv
| scattered across the codebase.
| 
| Run with go test ./config/... or go test for this package from the
| framework root.
| 
*/

import (
	"strings"
	"testing"

	"github.com/lsgser/gofreight/config"
)

func TestDatabaseConfigSQLite(t *testing.T) {
	cfg := config.DatabaseConfig{Connection: "sqlite", Database: "db/test.db"}
	got := cfg.URL()
	if got != "sqlite://db/test.db" {
		t.Fatalf("got %q", got)
	}
}

func TestDatabaseConfigPostgres(t *testing.T) {
	cfg := config.DatabaseConfig{
		Connection: "pgsql",
		Host:       "127.0.0.1",
		Port:       "5432",
		Database:   "myapp",
		Username:   "postgres",
		Password:   "secret",
		SSLMode:    "disable",
	}
	got := cfg.URL()
	for _, part := range []string{"postgres://", "127.0.0.1:5432", "/myapp", "sslmode=disable"} {
		if !strings.Contains(got, part) {
			t.Fatalf("missing %q in %q", part, got)
		}
	}
}

func TestDatabaseConfigMySQL(t *testing.T) {
	cfg := config.DatabaseConfig{
		Connection: "mysql",
		Host:       "localhost",
		Database:   "blog",
		Username:   "root",
		Password:   "pass",
	}
	got := cfg.URL()
	for _, part := range []string{"mysql://", "root:pass@", "localhost:3306", "/blog"} {
		if !strings.Contains(got, part) {
			t.Fatalf("missing %q in %q", part, got)
		}
	}
}

func TestResolveDatabaseURLPrefersDATABASEURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://override/db")
	t.Setenv("DB_CONNECTION", "sqlite")
	got := config.ResolveDatabaseURL(nil)
	if got != "postgres://override/db" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveDatabaseURLSQLiteDefault(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_URL", "")
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("DB_DATABASE", "")
	got := config.ResolveDatabaseURL(nil)
	if got != "sqlite://db/development.db" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveDatabaseURLFromComponents(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_URL", "")
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("DB_DATABASE", "db/custom.db")
	got := config.ResolveDatabaseURL(nil)
	if got != "sqlite://db/custom.db" {
		t.Fatalf("got %q", got)
	}
}
