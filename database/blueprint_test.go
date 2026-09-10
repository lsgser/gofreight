package database

/*
|--------------------------------------------------------------------------
| Blueprint
|--------------------------------------------------------------------------
|
| Test suite for Blueprint in the database package.
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

import "testing"

func TestCreateTableBlueprint(t *testing.T) {
	up, down := CreateTableBlueprint("articles", func(b *Blueprint) {
		b.StringColumn("title", ColNotNull())
		b.BooleanColumn("published")
	})
	if up == "" || down == "" {
		t.Fatal("expected up and down SQL")
	}
	if !contains(up, "CREATE TABLE") || !contains(up, "articles") {
		t.Fatalf("unexpected up: %s", up)
	}
	if !contains(down, "DROP TABLE") {
		t.Fatalf("unexpected down: %s", down)
	}
}

func TestCreateTableBlueprintUniqueColumn(t *testing.T) {
	up, _ := CreateTableBlueprint("users", func(b *Blueprint) {
		b.String("email").NotNull().Unique()
	})
	if !contains(up, "email TEXT NOT NULL UNIQUE") {
		t.Fatalf("expected unique email column, got: %s", up)
	}
}

func TestCreateTableBlueprintUniqueIndex(t *testing.T) {
	up, _ := CreateTableBlueprint("teams", func(b *Blueprint) {
		b.String("name").NotNull()
		b.UniqueIndex("name")
	})
	if !contains(up, "CREATE UNIQUE INDEX") {
		t.Fatalf("expected unique index, got: %s", up)
	}
}

func TestCreateTableBlueprintSoftDeletes(t *testing.T) {
	up, _ := CreateTableBlueprint("posts", func(b *Blueprint) {
		b.String("title").NotNull()
		b.SoftDeletes()
	})
	if !contains(up, "deleted_at") {
		t.Fatalf("expected deleted_at column, got: %s", up)
	}
}

func TestCreateTableBlueprintSoftDeletesTz(t *testing.T) {
	up, _ := CreateTableBlueprint("posts", func(b *Blueprint) {
		b.String("title").NotNull()
		b.SoftDeletesTz()
	})
	if !contains(up, "deleted_at") {
		t.Fatalf("expected deleted_at column, got: %s", up)
	}
}

func TestAlterTableBlueprint(t *testing.T) {
	up, down := AlterTableBlueprint("posts", func(b *Blueprint) {
		b.StringColumn("slug")
	})
	if !contains(up, "ADD COLUMN") || !contains(down, "DROP COLUMN") {
		t.Fatalf("alter mismatch up=%s down=%s", up, down)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
