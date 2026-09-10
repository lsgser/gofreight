package database_test

/*
|--------------------------------------------------------------------------
| Dialect
|--------------------------------------------------------------------------
|
| Test suite for Dialect in the database package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| The database package manages connections, fluent schema blueprints, Go
| and SQL migrations, seeding, and introspection.
| 
| Migrations run via gofreight migrate; blueprints generate portable DDL
| across SQLite, PostgreSQL, and MySQL.
| 
| Lower-level query helpers complement the model ORM in the model package.
| 
| Run with go test ./database/... or go test for this package from the
| framework root.
| 
*/

import (
	"testing"

	"github.com/lsgser/gofreight/database"
)

func TestDetectDriver(t *testing.T) {
	tests := []struct {
		url    string
		driver database.Driver
	}{
		{"postgres://localhost/mydb", database.Postgres},
		{"postgresql://localhost/mydb", database.Postgres},
		{"sqlite://data/app.db", database.SQLite},
		{"file:blog.db", database.SQLite},
		{"mysql://user:pass@localhost/mydb", database.MySQL},
		{"mariadb://user:pass@localhost/mydb", database.MariaDB},
	}

	for _, tt := range tests {
		driver, err := database.DetectDriver(tt.url)
		if err != nil {
			t.Fatalf("url %s: %v", tt.url, err)
		}
		if driver != tt.driver {
			t.Fatalf("url %s: expected %s, got %s", tt.url, tt.driver, driver)
		}
	}
}

func TestPlaceholder(t *testing.T) {
	// Default is postgres before connect
	database.Reset()
	if p := database.Placeholder(1); p != "$1" {
		t.Fatalf("expected $1, got %s", p)
	}
}

func TestSQLiteConnect(t *testing.T) {
	database.Reset()
	db, err := database.Connect("sqlite://:memory:")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if database.Placeholder(1) != "?" {
		t.Fatalf("expected ?, got %s", database.Placeholder(1))
	}
}

func TestMigrate(t *testing.T) {
	database.Reset()
	_, err := database.Connect("sqlite://:memory:")
	if err != nil {
		t.Fatal(err)
	}

	err = database.Migrate(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)
	if err != nil {
		t.Fatal(err)
	}

	migrator := database.NewMigrator(t.TempDir())
	// empty dir should not error
	if err := migrator.Up(); err != nil {
		t.Fatal(err)
	}
}
