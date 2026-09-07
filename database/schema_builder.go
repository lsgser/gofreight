package database

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// SchemaCreate creates a table using the blueprint DSL (Laravel Schema::create).
func SchemaCreate(ctx context.Context, table string, fn func(*Blueprint)) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	b := NewBlueprint(table)
	fn(b)
	b.flushPendingColumn()

	sql := buildCreateTableSQL(table, b.adds)
	if _, err := DB().ExecContext(ctx, sql); err != nil {
		return fmt.Errorf("create table %s: %w", table, err)
	}
	for _, idx := range b.indexes {
		if _, err := DB().ExecContext(ctx, idx); err != nil {
			return fmt.Errorf("create index on %s: %w", table, err)
		}
	}
	return nil
}

// SchemaTable alters a table using the blueprint DSL (Laravel Schema::table).
func SchemaTable(ctx context.Context, table string, fn func(*Blueprint)) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	b := NewBlueprint(table)
	fn(b)
	b.flushPendingColumn()

	for _, col := range b.adds {
		if err := AddColumn(ctx, table, col); err != nil {
			return fmt.Errorf("add column %s to %s: %w", col.Name, table, err)
		}
	}
	for _, col := range b.drops {
		if err := DropColumn(ctx, table, col); err != nil {
			return fmt.Errorf("drop column %s from %s: %w", col, table, err)
		}
	}
	for _, idx := range b.indexes {
		if _, err := DB().ExecContext(ctx, idx); err != nil {
			return fmt.Errorf("create index on %s: %w", table, err)
		}
	}
	return nil
}

// SchemaDrop drops a table (Laravel Schema::drop / dropIfExists).
func SchemaDrop(ctx context.Context, table string) error {
	return DropTable(ctx, table)
}

// SchemaDropIfExists drops a table when it exists.
func SchemaDropIfExists(ctx context.Context, table string) error {
	return DropTable(ctx, table)
}

// SchemaRename renames a table (Laravel Schema::rename).
func SchemaRename(ctx context.Context, from, to string) error {
	return RenameTable(ctx, from, to)
}

// HasTable reports whether a table exists (Laravel Schema::hasTable).
func HasTable(ctx context.Context, table string) (bool, error) {
	if err := validateIdent(table); err != nil {
		return false, err
	}
	tables, err := ListTables(ctx)
	if err != nil {
		return false, err
	}
	for _, t := range tables {
		if t == table {
			return true, nil
		}
	}
	return false, nil
}

// HasColumn reports whether a table has a column (Laravel Schema::hasColumn).
func HasColumn(ctx context.Context, table, column string) (bool, error) {
	if err := validateIdent(table); err != nil {
		return false, err
	}
	if err := validateIdent(column); err != nil {
		return false, err
	}
	schema, err := TableSchema(ctx, table)
	if err != nil {
		return false, err
	}
	for _, c := range schema.Columns {
		if c.Name == column {
			return true, nil
		}
	}
	return false, nil
}

// HasGoMigrations reports whether any Go migrations are registered.
func HasGoMigrations() bool {
	return len(RegisteredMigrations()) > 0
}

// HasGoMigrationFiles reports whether dir contains .go migration sources.
func HasGoMigrationFiles(dir string) bool {
	entries, err := readDirNames(dir)
	if err != nil {
		return false
	}
	for _, name := range entries {
		if strings.HasSuffix(name, ".go") && name != "doc.go" {
			return true
		}
	}
	return false
}

func readDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
