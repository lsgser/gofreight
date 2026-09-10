package database_test

/*
|--------------------------------------------------------------------------
| Introspect
|--------------------------------------------------------------------------
|
| Test suite for Introspect in the database package.
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
	"context"
	"testing"

	"github.com/lsgser/gofreight/database"
)

func TestIntrospectSQLite(t *testing.T) {
	database.Reset()
	_, err := database.Connect("sqlite://:memory:")
	if err != nil {
		t.Fatal(err)
	}
	database.Migrate(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT
	)`)

	ctx := context.Background()
	tables, err := database.ListTables(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 1 || tables[0] != "users" {
		t.Fatalf("tables: %v", tables)
	}

	schema, err := database.TableSchema(ctx, "users")
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Columns) < 3 {
		t.Fatalf("expected 3 columns, got %d", len(schema.Columns))
	}

	database.InsertRow(ctx, "users", map[string]any{"name": "Alice", "email": "a@test.com"})
	count, _ := database.TableCount(ctx, "users")
	if count != 1 {
		t.Fatalf("count: %d", count)
	}

	rows, err := database.TableRows(ctx, "users", 10, 0)
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows: %v, err: %v", rows, err)
	}
}

func TestExecQueryReadOnly(t *testing.T) {
	database.Reset()
	_, _ = database.Connect("sqlite://:memory:")
	ctx := context.Background()

	_, err := database.ExecQuery(ctx, "DELETE FROM users")
	if err == nil {
		t.Fatal("expected error for non-SELECT query")
	}
}
