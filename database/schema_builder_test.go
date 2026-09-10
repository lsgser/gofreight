package database_test

/*
|--------------------------------------------------------------------------
| Schema Builder
|--------------------------------------------------------------------------
|
| Test suite for Schema Builder in the database package.
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
	"strings"
	"testing"

	"github.com/lsgser/gofreight/database"
)

func TestSchemaCreateBlueprint(t *testing.T) {
	setupSchema(t)
	ctx := context.Background()

	err := database.SchemaCreate(ctx, "users", func(b *database.Blueprint) {
		b.Id()
		b.String("email").NotNull().Unique()
		b.Timestamps()
	})
	if err != nil {
		t.Fatal(err)
	}

	schema, err := database.TableSchema(ctx, "users")
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Columns) < 4 {
		t.Fatalf("expected id, email, timestamps, got %d columns", len(schema.Columns))
	}
}

func TestGoMigratorUpDown(t *testing.T) {
	setupSchema(t)
	database.ResetRegisteredMigrations()

	database.RegisterMigration("0001_test", func(ctx context.Context) error {
		return database.SchemaCreate(ctx, "items", func(b *database.Blueprint) {
			b.Id()
			b.String("name").NotNull()
			b.Timestamps()
		})
	}, func(ctx context.Context) error {
		return database.SchemaDropIfExists(ctx, "items")
	})

	m := database.NewGoMigrator()
	if err := m.Up(); err != nil {
		t.Fatal(err)
	}
	if err := m.Down(); err != nil {
		t.Fatal(err)
	}
}

func TestSchemaCreateSoftDeletes(t *testing.T) {
	setupSchema(t)
	ctx := context.Background()

	err := database.SchemaCreate(ctx, "posts", func(b *database.Blueprint) {
		b.Id()
		b.String("title").NotNull()
		b.SoftDeletes()
		b.Timestamps()
	})
	if err != nil {
		t.Fatal(err)
	}

	schema, err := database.TableSchema(ctx, "posts")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range schema.Columns {
		if c.Name == "deleted_at" {
			found = true
			if !c.Nullable {
				t.Fatal("deleted_at should be nullable")
			}
		}
	}
	if !found {
		t.Fatal("deleted_at column missing")
	}
}

func TestBlueprintTimestampsTz(t *testing.T) {
	up, _ := database.CreateTableBlueprint("events", func(b *database.Blueprint) {
		b.Id()
		b.String("name").NotNull()
		b.TimestampsTz()
	})
	if !strings.Contains(up, "created_at") || !strings.Contains(up, "updated_at") {
		t.Fatalf("missing timestamps: %s", up)
	}
}

func TestBlueprintDropSoftDeletes(t *testing.T) {
	up, down := database.AlterTableBlueprint("posts", func(b *database.Blueprint) {
		b.DropSoftDeletes()
	})
	if !strings.Contains(up, "DROP COLUMN") || !strings.Contains(up, "deleted_at") {
		t.Fatalf("unexpected up: %s", up)
	}
	_ = down
}

func TestHasTable(t *testing.T) {
	setupSchema(t)
	ctx := context.Background()
	if err := database.SchemaCreate(ctx, "widgets", func(b *database.Blueprint) {
		b.Id()
	}); err != nil {
		t.Fatal(err)
	}
	ok, err := database.HasTable(ctx, "widgets")
	if err != nil || !ok {
		t.Fatalf("HasTable: ok=%v err=%v", ok, err)
	}
	ok, err = database.HasColumn(ctx, "widgets", "id")
	if err != nil || !ok {
		t.Fatalf("HasColumn: ok=%v err=%v", ok, err)
	}
}

func TestBlueprintIdTimestamps(t *testing.T) {
	up, _ := database.CreateTableBlueprint("posts", func(b *database.Blueprint) {
		b.Id()
		b.String("title").NotNull()
		b.Timestamps()
	})
	if !strings.Contains(up, "id INTEGER PRIMARY KEY") {
		t.Fatalf("missing id: %s", up)
	}
	if !strings.Contains(up, "created_at") || !strings.Contains(up, "updated_at") {
		t.Fatalf("missing timestamps: %s", up)
	}
}
