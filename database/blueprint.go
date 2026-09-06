package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Blueprint builds migration SQL programmatically.
type Blueprint struct {
	table   string
	creates []ColumnDef
	adds    []ColumnDef
	drops   []string
	indexes []string
}

// NewBlueprint creates a migration blueprint for a table.
func NewBlueprint(table string) *Blueprint {
	return &Blueprint{table: table}
}

type colOpt func(*ColumnDef)

func colNotNull() colOpt   { return func(c *ColumnDef) { c.Nullable = false } }
func colNullable() colOpt  { return func(c *ColumnDef) { c.Nullable = true } }
func colDefault(v string) colOpt { return func(c *ColumnDef) { c.Default = v } }

func applyColOpts(col ColumnDef, opts ...colOpt) ColumnDef {
	col.Nullable = true
	for _, o := range opts {
		o(&col)
	}
	return col
}

// StringColumn adds a text column.
func (b *Blueprint) StringColumn(name string, opts ...colOpt) {
	b.adds = append(b.adds, applyColOpts(ColumnDef{Name: name, Type: "TEXT"}, opts...))
}

// IntegerColumn adds an integer column.
func (b *Blueprint) IntegerColumn(name string, opts ...colOpt) {
	b.adds = append(b.adds, applyColOpts(ColumnDef{Name: name, Type: "INTEGER"}, opts...))
}

// BooleanColumn adds a boolean column.
func (b *Blueprint) BooleanColumn(name string, opts ...colOpt) {
	b.adds = append(b.adds, applyColOpts(ColumnDef{Name: name, Type: "BOOLEAN"}, opts...))
}

// DateTimeColumn adds a datetime column.
func (b *Blueprint) DateTimeColumn(name string, opts ...colOpt) {
	b.adds = append(b.adds, applyColOpts(ColumnDef{Name: name, Type: "TEXT"}, opts...))
}

// DropColumn marks a column for rollback.
func (b *Blueprint) DropColumn(name string) {
	b.drops = append(b.drops, name)
}

// Index adds an index statement.
func (b *Blueprint) Index(columns ...string) {
	b.indexes = append(b.indexes, fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_%s ON %s (%s);",
		b.table, strings.Join(columns, "_"), b.table, strings.Join(columns, ", ")))
}

// ToUpSQL generates the UP migration SQL.
func (b *Blueprint) ToUpSQL() string {
	if len(b.creates) > 0 {
		return buildCreateTableSQL(b.table, b.creates)
	}
	var parts []string
	for _, col := range b.adds {
		parts = append(parts, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", b.table, blueprintColumnSQL(col)))
	}
	parts = append(parts, b.indexes...)
	return strings.Join(parts, "\n")
}

// ToDownSQL generates the DOWN migration SQL.
func (b *Blueprint) ToDownSQL() string {
	if len(b.creates) > 0 {
		return fmt.Sprintf("DROP TABLE IF EXISTS %s;", b.table)
	}
	var parts []string
	for _, col := range b.drops {
		if col != "" {
			parts = append(parts, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", b.table, col))
		}
	}
	for i := len(b.adds) - 1; i >= 0; i-- {
		parts = append(parts, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", b.table, b.adds[i].Name))
	}
	return strings.Join(parts, "\n")
}

func blueprintColumnSQL(c ColumnDef) string {
	s := c.Name + " " + c.Type
	if !c.Nullable {
		s += " NOT NULL"
	}
	if c.Default != "" {
		s += " DEFAULT " + c.Default
	}
	return s
}

// CreateTableBlueprint builds a create-table migration pair.
func CreateTableBlueprint(table string, fn func(*Blueprint)) (up, down string) {
	b := NewBlueprint(table)
	fn(b)
	cols := []ColumnDef{
		{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true, Nullable: false},
	}
	cols = append(cols, b.adds...)
	cols = append(cols,
		ColumnDef{Name: "created_at", Type: "TEXT", Default: "(datetime('now'))", Nullable: true},
		ColumnDef{Name: "updated_at", Type: "TEXT", Default: "(datetime('now'))", Nullable: true},
	)
	b.creates = cols
	return b.ToUpSQL(), b.ToDownSQL()
}

// AlterTableBlueprint builds an alter-table migration with auto-generated down.
func AlterTableBlueprint(table string, fn func(*Blueprint)) (up, down string) {
	b := NewBlueprint(table)
	fn(b)
	for _, col := range b.adds {
		b.drops = append(b.drops, col.Name)
	}
	return b.ToUpSQL(), b.ToDownSQL()
}

// WriteMigrationPair writes up and down SQL files.
func WriteMigrationPair(dir, version, up, down string) error {
	if err := os.WriteFile(filepath.Join(dir, version+".sql"), []byte(up), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, version+"_down.sql"), []byte(down), 0644)
}
