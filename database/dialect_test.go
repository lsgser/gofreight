package database_test

import (
	"testing"

	"github.com/gofreight/gofreight/database"
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
