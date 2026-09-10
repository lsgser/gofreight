package database

/*
|--------------------------------------------------------------------------
| Schema
|--------------------------------------------------------------------------
|
| Implements Schema as part of the database package in the Gofreight
| framework. Key symbols: ColumnDef, CreateTable, DropTable, AddColumn,
| ExportTableSQL, ExecDDL.
| 
| The database package manages connections, fluent schema blueprints, Go
| and SQL migrations, seeding, and introspection.
| 
| Migrations run via gofreight migrate; blueprints generate portable DDL
| across SQLite, PostgreSQL, and MySQL.
| 
| Lower-level query helpers complement the model ORM in the model package.
| 
| Symbols defined here include: ColumnDef (exported type);
| CommonColumnTypes (exported value); CreateTable (CreateTable creates a
| new table from column definitions.); DropTable (DropTable removes a
| table.); AddColumn (AddColumn adds a column to an existing table.);
| ExportTableSQL (ExportTableSQL returns CREATE TABLE + INSERT statements
| for a table.); ExecDDL (ExecDDL runs a DDL statement (CREATE, ALTER,
| DROP) in development admin.); ImportSQL (ImportSQL runs multiple SQL
| statements separated by semicolons.).
| 
*/

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// ColumnDef defines a column for CREATE TABLE / ADD COLUMN.
type ColumnDef struct {
	Name          string
	Type          string
	Nullable      bool
	PrimaryKey    bool
	AutoIncrement bool
	Unique        bool
	Default       string
}

// Common column types for the admin UI.
var CommonColumnTypes = []string{
	"INTEGER", "BIGINT", "VARCHAR(255)", "TEXT", "BOOLEAN",
	"TIMESTAMP", "DATE", "FLOAT", "JSON", "UUID",
}

// CreateTable creates a new table from column definitions.
func CreateTable(ctx context.Context, table string, columns []ColumnDef) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	if len(columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}

	sql := buildCreateTableSQL(table, columns)
	_, err := DB().ExecContext(ctx, sql)
	return err
}

// DropTable removes a table.
func DropTable(ctx context.Context, table string) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	if table == "schema_migrations" {
		return fmt.Errorf("cannot drop schema_migrations")
	}
	_, err := DB().ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
	return err
}

// AddColumn adds a column to an existing table.
func AddColumn(ctx context.Context, table string, col ColumnDef) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	if err := validateIdent(col.Name); err != nil {
		return err
	}

	colType := normalizeColumnType(col.Type)
	def := fmt.Sprintf("%s %s", col.Name, colType)
	if !col.Nullable {
		def += " NOT NULL"
	}
	if col.Default != "" {
		def += " DEFAULT " + col.Default
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", table, def)
	_, err := DB().ExecContext(ctx, sql)
	return err
}

// ExportTableSQL returns CREATE TABLE + INSERT statements for a table.
func ExportTableSQL(ctx context.Context, table string) (string, error) {
	if err := validateIdent(table); err != nil {
		return "", err
	}

	schema, err := TableSchema(ctx, table)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("-- Export: %s\n", table))

	cols := make([]ColumnDef, len(schema.Columns))
	for i, c := range schema.Columns {
		cols[i] = ColumnDef{
			Name: c.Name, Type: c.Type, Nullable: c.Nullable, PrimaryKey: c.PrimaryKey,
		}
	}
	sb.WriteString(buildCreateTableSQL(table, cols))
	sb.WriteString(";\n\n")

	rows, err := TableRows(ctx, table, 10000, 0)
	if err != nil {
		return sb.String(), err
	}

	colNames := make([]string, len(schema.Columns))
	for i, c := range schema.Columns {
		colNames[i] = c.Name
	}

	for _, row := range rows {
		vals := make([]string, len(colNames))
		for i, col := range colNames {
			v := row[col]
			if v == nil {
				vals[i] = "NULL"
			} else {
				vals[i] = fmt.Sprintf("'%v'", escapeSQLString(fmt.Sprintf("%v", v)))
			}
		}
		sb.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n",
			table, strings.Join(colNames, ", "), strings.Join(vals, ", ")))
	}

	return sb.String(), nil
}

// ExecDDL runs a DDL statement (CREATE, ALTER, DROP) in development admin.
func ExecDDL(ctx context.Context, sql string) error {
	trimmed := strings.TrimSpace(strings.ToUpper(sql))
	allowed := strings.HasPrefix(trimmed, "CREATE") ||
		strings.HasPrefix(trimmed, "ALTER") ||
		strings.HasPrefix(trimmed, "DROP") ||
		strings.HasPrefix(trimmed, "INSERT") ||
		strings.HasPrefix(trimmed, "UPDATE") ||
		strings.HasPrefix(trimmed, "DELETE")

	if !allowed {
		return fmt.Errorf("statement type not allowed")
	}
	_, err := DB().ExecContext(ctx, sql)
	return err
}

// ImportSQL runs multiple SQL statements separated by semicolons.
func ImportSQL(ctx context.Context, sql string) error {
	statements := splitSQLStatements(sql)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if err := ExecDDL(ctx, stmt); err != nil {
			return fmt.Errorf("failed on: %s...: %w", truncate(stmt, 60), err)
		}
	}
	return nil
}

func buildCreateTableSQL(table string, columns []ColumnDef) string {
	var parts []string
	hasPK := false
	hasCreatedAt := false
	hasUpdatedAt := false

	for _, col := range columns {
		if err := validateIdent(col.Name); err != nil {
			continue
		}
		switch col.Name {
		case "created_at":
			hasCreatedAt = true
		case "updated_at":
			hasUpdatedAt = true
		}
		colType := normalizeColumnType(col.Type)

		if col.PrimaryKey && col.AutoIncrement {
			parts = append(parts, fmt.Sprintf("%s %s", col.Name, AutoIncrement()))
			hasPK = true
			continue
		}

		def := fmt.Sprintf("%s %s", col.Name, colType)
		if col.PrimaryKey {
			def += " PRIMARY KEY"
			hasPK = true
		}
		if !col.Nullable && !col.PrimaryKey {
			def += " NOT NULL"
		}
		if col.Default != "" {
			def += " DEFAULT " + col.Default
		}
		if col.Unique {
			def += " UNIQUE"
		}
		parts = append(parts, def)
	}

	if !hasPK {
		parts = append([]string{fmt.Sprintf("id %s", AutoIncrement())}, parts...)
	}

	ts := TimestampType()
	if !hasCreatedAt {
		parts = append(parts, fmt.Sprintf("created_at %s", ts))
	}
	if !hasUpdatedAt {
		parts = append(parts, fmt.Sprintf("updated_at %s", ts))
	}

	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n\t%s\n)", table, strings.Join(parts, ",\n\t"))
}

func normalizeColumnType(t string) string {
	t = strings.ToUpper(strings.TrimSpace(t))
	switch DriverName() {
	case SQLite:
		switch {
		case strings.Contains(t, "VARCHAR"), strings.Contains(t, "CHAR"):
			return "TEXT"
		case t == "BOOLEAN", t == "BOOL":
			return "INTEGER"
		case t == "JSON", t == "JSONB":
			return "TEXT"
		case t == "UUID", t == "TIMESTAMPTZ":
			return "TEXT"
		case t == "TIMESTAMP", t == "DATE", t == "TIME":
			return "TEXT"
		case t == "BIGINT":
			return "INTEGER"
		case strings.HasPrefix(t, "DECIMAL"):
			return "REAL"
		}
	case MySQL, MariaDB:
		if t == "JSONB" {
			return "JSON"
		}
	}
	return t
}

func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func splitSQLStatements(sql string) []string {
	var stmts []string
	var current strings.Builder
	inString := false

	for i := 0; i < len(sql); i++ {
		c := sql[i]
		if c == '\'' {
			inString = !inString
		}
		if c == ';' && !inString {
			stmts = append(stmts, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(c)
	}
	if s := strings.TrimSpace(current.String()); s != "" {
		stmts = append(stmts, s)
	}
	return stmts
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// DropColumn removes a column from a table.
func DropColumn(ctx context.Context, table, column string) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	if err := validateIdent(column); err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", table, column)
	if DriverName() == MySQL || DriverName() == MariaDB {
		sql = fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", table, column)
	}
	_, err := DB().ExecContext(ctx, sql)
	return err
}

// RenameTable renames a database table.
func RenameTable(ctx context.Context, from, to string) error {
	if err := validateIdent(from); err != nil {
		return err
	}
	if err := validateIdent(to); err != nil {
		return err
	}
	var sql string
	switch DriverName() {
	case MySQL, MariaDB:
		sql = fmt.Sprintf("RENAME TABLE %s TO %s", from, to)
	default:
		sql = fmt.Sprintf("ALTER TABLE %s RENAME TO %s", from, to)
	}
	_, err := DB().ExecContext(ctx, sql)
	return err
}

// RenameColumn renames a column.
func RenameColumn(ctx context.Context, table, from, to string) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	if err := validateIdent(from); err != nil {
		return err
	}
	if err := validateIdent(to); err != nil {
		return err
	}
	var sql string
	switch DriverName() {
	case SQLite:
		return fmt.Errorf("sqlite requires table rebuild to rename columns; use import/export")
	case MySQL, MariaDB:
		sql = fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", table, from, to)
	default:
		sql = fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", table, from, to)
	}
	_, err := DB().ExecContext(ctx, sql)
	return err
}

// AddIndex creates an index on a table column.
func AddIndex(ctx context.Context, table, indexName, column string, unique bool) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	if err := validateIdent(indexName); err != nil {
		return err
	}
	if err := validateIdent(column); err != nil {
		return err
	}
	kind := "INDEX"
	if unique {
		kind = "UNIQUE INDEX"
	}
	sql := fmt.Sprintf("CREATE %s IF NOT EXISTS %s ON %s (%s)", kind, indexName, table, column)
	_, err := DB().ExecContext(ctx, sql)
	return err
}

// SaveTableAsMigration writes CREATE TABLE SQL to db/migrate as a new migration file.
func SaveTableAsMigration(ctx context.Context, dir, table string) (string, error) {
	sqlText, err := ExportTableSQL(ctx, table)
	if err != nil {
		return "", err
	}
	return CreateMigrationFromSQL(dir, "create_"+table, sqlText)
}

// CreateMigrationFromSQL creates a timestamped migration file with SQL content.
func CreateMigrationFromSQL(dir, name, sql string) (string, error) {
	path, err := CreateMigration(dir, name)
	if err != nil {
		return "", err
	}
	content := []byte("-- Auto-generated from admin\n\n" + sql)
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", err
	}
	return path, nil
}
