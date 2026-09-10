package database

/*
|--------------------------------------------------------------------------
| Blueprint
|--------------------------------------------------------------------------
|
| Fluent schema builder for migrations: columns, indexes, timestamps, soft
| deletes, and dialect-specific SQL.
| 
| Used by db/migrate/*.go files and gofreight make:migration output.
| 
| The database package manages connections, fluent schema blueprints, Go
| and SQL migrations, seeding, and introspection.
| 
| Migrations run via gofreight migrate; blueprints generate portable DDL
| across SQLite, PostgreSQL, and MySQL.
| 
| Lower-level query helpers complement the model ORM in the model package.
| 
| Symbols defined here include: Blueprint (exported type); NewBlueprint
| (NewBlueprint creates a migration blueprint for a table.); ColNotNull
| (ColNotNull marks a column NOT NULL (for StringColumn, IntegerColumn,
| etc.).); ColNullable (ColNullable marks a column nullable.); ColDefault
| (ColDefault sets a column default value.); ColUnique (ColUnique marks a
| column UNIQUE.); ColumnBuilder (exported type); String (String adds a
| string/text column. Chain NotNull(), Unique(), Default(), etc.).
| 
*/

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
	pending *ColumnBuilder
}

// NewBlueprint creates a migration blueprint for a table.
func NewBlueprint(table string) *Blueprint {
	return &Blueprint{table: table}
}

type colOpt func(*ColumnDef)

func colNotNull() colOpt   { return func(c *ColumnDef) { c.Nullable = false } }
func colNullable() colOpt  { return func(c *ColumnDef) { c.Nullable = true } }
func colDefault(v string) colOpt { return func(c *ColumnDef) { c.Default = v } }
func colUnique() colOpt    { return func(c *ColumnDef) { c.Unique = true } }

// ColNotNull marks a column NOT NULL (for StringColumn, IntegerColumn, etc.).
func ColNotNull() colOpt { return colNotNull() }

// ColNullable marks a column nullable.
func ColNullable() colOpt { return colNullable() }

// ColDefault sets a column default value.
func ColDefault(v string) colOpt { return colDefault(v) }

// ColUnique marks a column UNIQUE.
func ColUnique() colOpt { return colUnique() }

func applyColOpts(col ColumnDef, opts ...colOpt) ColumnDef {
	col.Nullable = true
	for _, o := range opts {
		o(&col)
	}
	return col
}

// ColumnBuilder provides a fluent API for defining columns (Laravel-style chaining).
type ColumnBuilder struct {
	bp  *Blueprint
	col ColumnDef
}

func (b *Blueprint) flushPendingColumn() {
	if b.pending == nil {
		return
	}
	b.adds = append(b.adds, b.pending.col)
	b.pending = nil
}

func (b *Blueprint) beginColumn(name, colType string) *ColumnBuilder {
	b.flushPendingColumn()
	b.pending = &ColumnBuilder{
		bp:  b,
		col: ColumnDef{Name: name, Type: colType, Nullable: true},
	}
	return b.pending
}

// String adds a string/text column. Chain NotNull(), Unique(), Default(), etc.
func (b *Blueprint) String(name string) *ColumnBuilder {
	return b.beginColumn(name, "TEXT")
}

// Text adds a long-text column.
func (b *Blueprint) Text(name string) *ColumnBuilder {
	return b.beginColumn(name, "TEXT")
}

// Integer adds an integer column.
func (b *Blueprint) Integer(name string) *ColumnBuilder {
	return b.beginColumn(name, "INTEGER")
}

// Boolean adds a boolean column.
func (b *Blueprint) Boolean(name string) *ColumnBuilder {
	return b.beginColumn(name, "BOOLEAN")
}

// DateTime adds a datetime/timestamp column.
func (b *Blueprint) DateTime(name string) *ColumnBuilder {
	return b.beginColumn(name, "TEXT")
}

// Id adds an auto-increment primary key column (Laravel $table->id()).
func (b *Blueprint) Id() {
	b.flushPendingColumn()
	b.adds = append(b.adds, ColumnDef{
		Name:          "id",
		Type:          "INTEGER",
		PrimaryKey:    true,
		AutoIncrement: true,
		Nullable:      false,
	})
}

// Timestamps adds created_at and updated_at columns (Laravel $table->timestamps()).
func (b *Blueprint) Timestamps() {
	b.addTimestampColumns(false)
}

// TimestampsTz adds timezone-aware created_at and updated_at (Laravel $table->timestampsTz()).
func (b *Blueprint) TimestampsTz() {
	b.addTimestampColumns(true)
}

func (b *Blueprint) addTimestampColumns(tz bool) {
	b.flushPendingColumn()
	now := "(" + NowFunc() + ")"
	b.adds = append(b.adds,
		ColumnDef{Name: "created_at", Type: TimestampColumnType(tz), Default: now, Nullable: true},
		ColumnDef{Name: "updated_at", Type: TimestampColumnType(tz), Default: now, Nullable: true},
	)
}

// SoftDeletes adds a nullable deleted_at column (Laravel $table->softDeletes()).
func (b *Blueprint) SoftDeletes(column ...string) {
	b.addSoftDeleteColumn(false, column...)
}

// SoftDeletesTz adds a nullable timezone-aware deleted_at column (Laravel $table->softDeletesTz()).
func (b *Blueprint) SoftDeletesTz(column ...string) {
	b.addSoftDeleteColumn(true, column...)
}

func (b *Blueprint) addSoftDeleteColumn(tz bool, column ...string) {
	b.flushPendingColumn()
	name := "deleted_at"
	if len(column) > 0 && column[0] != "" {
		name = column[0]
	}
	colType := NullableTimestampType()
	if tz {
		colType = NullableTimestampTzType()
	}
	b.adds = append(b.adds, ColumnDef{Name: name, Type: colType, Nullable: true})
}

// DropSoftDeletes drops the deleted_at column (Laravel $table->dropSoftDeletes()).
func (b *Blueprint) DropSoftDeletes(column ...string) {
	name := "deleted_at"
	if len(column) > 0 && column[0] != "" {
		name = column[0]
	}
	b.DropColumn(name)
}

// DropSoftDeletesTz drops the deleted_at column (alias of DropSoftDeletes).
func (b *Blueprint) DropSoftDeletesTz(column ...string) {
	b.DropSoftDeletes(column...)
}

// DropTimestamps drops created_at and updated_at (Laravel $table->dropTimestamps()).
func (b *Blueprint) DropTimestamps() {
	b.DropColumn("created_at")
	b.DropColumn("updated_at")
}

// DropTimestampsTz drops created_at and updated_at (alias of DropTimestamps).
func (b *Blueprint) DropTimestampsTz() {
	b.DropTimestamps()
}

// RememberToken adds remember_token column (Laravel $table->rememberToken()).
func (b *Blueprint) RememberToken() {
	b.String("remember_token").Nullable()
}

// DropRememberToken drops remember_token (Laravel $table->dropRememberToken()).
func (b *Blueprint) DropRememberToken() {
	b.DropColumn("remember_token")
}

// Char adds a fixed-length string column (Laravel $table->char()).
func (b *Blueprint) Char(name string, length ...int) *ColumnBuilder {
	cb := b.beginColumn(name, "CHAR")
	if len(length) > 0 && length[0] > 0 {
		cb.col.Type = fmt.Sprintf("CHAR(%d)", length[0])
	}
	return cb
}

// BigInteger adds a BIGINT column (Laravel $table->bigInteger() / foreignId base type).
func (b *Blueprint) BigInteger(name string) *ColumnBuilder {
	return b.beginColumn(name, "BIGINT")
}

// ForeignId adds an unsigned bigint foreign key column (Laravel $table->foreignId()).
func (b *Blueprint) ForeignId(name string) *ColumnBuilder {
	return b.BigInteger(name)
}

// Float adds a floating-point column (Laravel $table->float()).
func (b *Blueprint) Float(name string) *ColumnBuilder {
	return b.beginColumn(name, "REAL")
}

// Decimal adds a decimal column (Laravel $table->decimal()).
func (b *Blueprint) Decimal(name string, total, places int) *ColumnBuilder {
	cb := b.beginColumn(name, "DECIMAL")
	if total > 0 {
		if places < 0 {
			places = 0
		}
		cb.col.Type = fmt.Sprintf("DECIMAL(%d,%d)", total, places)
	}
	return cb
}

// Json adds a JSON column (Laravel $table->json()).
func (b *Blueprint) Json(name string) *ColumnBuilder {
	return b.beginColumn(name, "JSON")
}

// Uuid adds a UUID column (Laravel $table->uuid()).
func (b *Blueprint) Uuid(name string) *ColumnBuilder {
	return b.beginColumn(name, "UUID")
}

// Date adds a date column (Laravel $table->date()).
func (b *Blueprint) Date(name string) *ColumnBuilder {
	return b.beginColumn(name, "DATE")
}

// Time adds a time column (Laravel $table->time()).
func (b *Blueprint) Time(name string) *ColumnBuilder {
	return b.beginColumn(name, "TIME")
}

// Timestamp adds a timestamp column (Laravel $table->timestamp()).
func (b *Blueprint) Timestamp(name string) *ColumnBuilder {
	return b.beginColumn(name, TimestampColumnType(false))
}

// NotNull marks the column NOT NULL.
func (cb *ColumnBuilder) NotNull() *ColumnBuilder {
	cb.col.Nullable = false
	return cb
}

// Nullable marks the column nullable.
func (cb *ColumnBuilder) Nullable() *ColumnBuilder {
	cb.col.Nullable = true
	return cb
}

// Default sets a SQL default expression or literal.
func (cb *ColumnBuilder) Default(v string) *ColumnBuilder {
	cb.col.Default = v
	return cb
}

// Unique marks the column UNIQUE.
func (cb *ColumnBuilder) Unique() *ColumnBuilder {
	cb.col.Unique = true
	return cb
}

// StringColumn adds a text column using functional options.
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

// Index adds a non-unique index.
func (b *Blueprint) Index(columns ...string) {
	b.indexes = append(b.indexes, fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_%s ON %s (%s);",
		b.table, strings.Join(columns, "_"), b.table, strings.Join(columns, ", ")))
}

// UniqueIndex adds a unique index on one or more columns.
func (b *Blueprint) UniqueIndex(columns ...string) {
	b.indexes = append(b.indexes, fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_%s ON %s (%s);",
		b.table, strings.Join(columns, "_"), b.table, strings.Join(columns, ", ")))
}

// Unique adds a unique index on one or more columns (Laravel $table->unique()).
func (b *Blueprint) Unique(columns ...string) {
	b.UniqueIndex(columns...)
}

// ToUpSQL generates the UP migration SQL.
func (b *Blueprint) ToUpSQL() string {
	b.flushPendingColumn()
	if len(b.creates) > 0 {
		sql := buildCreateTableSQL(b.table, b.creates)
		if len(b.indexes) > 0 {
			sql += "\n" + strings.Join(b.indexes, "\n")
		}
		return sql
	}
	var parts []string
	for _, col := range b.adds {
		parts = append(parts, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", b.table, blueprintColumnSQL(col)))
	}
	for _, col := range b.drops {
		if col != "" {
			parts = append(parts, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", b.table, col))
		}
	}
	parts = append(parts, b.indexes...)
	return strings.Join(parts, "\n")
}

// ToDownSQL generates the DOWN migration SQL.
func (b *Blueprint) ToDownSQL() string {
	b.flushPendingColumn()
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
	s := c.Name + " " + normalizeColumnType(c.Type)
	if !c.Nullable {
		s += " NOT NULL"
	}
	if c.Default != "" {
		s += " DEFAULT " + c.Default
	}
	if c.Unique {
		s += " UNIQUE"
	}
	return s
}

// CreateTableBlueprint builds a create-table migration pair.
func CreateTableBlueprint(table string, fn func(*Blueprint)) (up, down string) {
	b := NewBlueprint(table)
	fn(b)
	b.flushPendingColumn()
	var cols []ColumnDef
	if !hasColumnDef(b.adds, "id") {
		cols = append(cols, ColumnDef{
			Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true, Nullable: false,
		})
	}
	cols = append(cols, b.adds...)
	if !hasColumnDef(cols, "created_at") {
		cols = append(cols, ColumnDef{Name: "created_at", Type: "TEXT", Default: "(datetime('now'))", Nullable: true})
	}
	if !hasColumnDef(cols, "updated_at") {
		cols = append(cols, ColumnDef{Name: "updated_at", Type: "TEXT", Default: "(datetime('now'))", Nullable: true})
	}
	b.creates = cols
	return b.ToUpSQL(), b.ToDownSQL()
}

func hasColumnDef(cols []ColumnDef, name string) bool {
	for _, c := range cols {
		if c.Name == name {
			return true
		}
	}
	return false
}

// AlterTableBlueprint builds an alter-table migration with auto-generated down.
func AlterTableBlueprint(table string, fn func(*Blueprint)) (up, down string) {
	b := NewBlueprint(table)
	fn(b)
	b.flushPendingColumn()
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
