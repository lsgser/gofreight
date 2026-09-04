package database_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gofreight/gofreight/database"
)

func setupSchema(t *testing.T) {
	t.Helper()
	database.Reset()
	_, err := database.Connect("sqlite://:memory:")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateTable(t *testing.T) {
	setupSchema(t)
	ctx := context.Background()

	err := database.CreateTable(ctx, "widgets", []database.ColumnDef{
		{Name: "name", Type: "TEXT", Nullable: false},
		{Name: "price", Type: "FLOAT", Nullable: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	schema, err := database.TableSchema(ctx, "widgets")
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Columns) < 3 {
		t.Fatalf("expected at least 3 columns (id, name, price, timestamps), got %d", len(schema.Columns))
	}
}

func TestAddColumn(t *testing.T) {
	setupSchema(t)
	ctx := context.Background()

	if err := database.CreateTable(ctx, "items", []database.ColumnDef{
		{Name: "title", Type: "TEXT"},
	}); err != nil {
		t.Fatal(err)
	}

	if err := database.AddColumn(ctx, "items", database.ColumnDef{
		Name: "sku", Type: "VARCHAR(50)", Nullable: true,
	}); err != nil {
		t.Fatal(err)
	}

	schema, _ := database.TableSchema(ctx, "items")
	found := false
	for _, c := range schema.Columns {
		if c.Name == "sku" {
			found = true
		}
	}
	if !found {
		t.Fatal("sku column not found")
	}
}

func TestDropTable(t *testing.T) {
	setupSchema(t)
	ctx := context.Background()

	if err := database.CreateTable(ctx, "temp_table", []database.ColumnDef{
		{Name: "x", Type: "INTEGER"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := database.DropTable(ctx, "temp_table"); err != nil {
		t.Fatal(err)
	}
	tables, _ := database.ListTables(ctx)
	for _, name := range tables {
		if name == "temp_table" {
			t.Fatal("table should be dropped")
		}
	}
}

func TestDropSchemaMigrationsBlocked(t *testing.T) {
	setupSchema(t)
	err := database.DropTable(context.Background(), "schema_migrations")
	if err == nil {
		t.Fatal("expected error dropping schema_migrations")
	}
}

func TestImportExportSQL(t *testing.T) {
	setupSchema(t)
	ctx := context.Background()

	if err := database.CreateTable(ctx, "notes", []database.ColumnDef{
		{Name: "body", Type: "TEXT"},
	}); err != nil {
		t.Fatal(err)
	}

	if err := database.ImportSQL(ctx, "INSERT INTO notes (body) VALUES ('hello');"); err != nil {
		t.Fatal(err)
	}

	sqlText, err := database.ExportTableSQL(ctx, "notes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sqlText, "CREATE TABLE") {
		t.Fatal("export should include CREATE TABLE")
	}
	if !strings.Contains(sqlText, "hello") {
		t.Fatal("export should include inserted data")
	}
}
