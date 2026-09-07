package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// ColumnInfo describes a database column.
type ColumnInfo struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
	Default    sql.NullString
}

// TableInfo describes a database table.
type TableInfo struct {
	Name    string
	Columns []ColumnInfo
}

// PrimaryKey returns the primary key column name, defaulting to "id".
func (t *TableInfo) PrimaryKey() string {
	for _, c := range t.Columns {
		if c.PrimaryKey {
			return c.Name
		}
	}
	return "id"
}

// ListTables returns user tables excluding system/migration tables.
func ListTables(ctx context.Context) ([]string, error) {
	switch DriverName() {
	case SQLite:
		return listTablesSQLite(ctx)
	case MySQL, MariaDB:
		return listTablesMySQL(ctx)
	default:
		return listTablesPostgres(ctx)
	}
}

// TableSchema returns column metadata for a table.
func TableSchema(ctx context.Context, table string) (*TableInfo, error) {
	if err := validateIdent(table); err != nil {
		return nil, err
	}

	switch DriverName() {
	case SQLite:
		return tableSchemaSQLite(ctx, table)
	case MySQL, MariaDB:
		return tableSchemaMySQL(ctx, table)
	default:
		return tableSchemaPostgres(ctx, table)
	}
}

// TableRows returns paginated rows from a table.
func TableRows(ctx context.Context, table string, limit, offset int) ([]map[string]any, error) {
	return TableRowsQuery(ctx, table, limit, offset, "", nil, "")
}

// TableRowsQuery returns paginated, optionally filtered and sorted rows.
func TableRowsQuery(ctx context.Context, table string, limit, offset int, whereSQL string, whereArgs []any, orderBy string) ([]map[string]any, error) {
	if err := validateIdent(table); err != nil {
		return nil, err
	}
	query := "SELECT * FROM " + table
	if whereSQL != "" {
		query += " WHERE " + whereSQL
	}
	if orderBy != "" {
		query += " ORDER BY " + orderBy
	}
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	return queryRows(ctx, query, whereArgs...)
}

// TableCountQuery returns row count with optional WHERE clause.
func TableCountQuery(ctx context.Context, table, whereSQL string, whereArgs []any) (int64, error) {
	if err := validateIdent(table); err != nil {
		return 0, err
	}
	query := "SELECT COUNT(*) FROM " + table
	if whereSQL != "" {
		query += " WHERE " + whereSQL
	}
	var count int64
	err := DB().QueryRowContext(ctx, query, whereArgs...).Scan(&count)
	return count, err
}

// BuildSearchClause builds a safe WHERE fragment for table browsing.
func BuildSearchClause(column, op, value string) (string, []any, error) {
	if err := validateIdent(column); err != nil {
		return "", nil, err
	}
	switch op {
	case "eq":
		return column + " = " + Placeholder(1), []any{value}, nil
	case "ne":
		return column + " != " + Placeholder(1), []any{value}, nil
	case "like":
		return column + " LIKE " + Placeholder(1), []any{"%" + value + "%"}, nil
	case "gt":
		return column + " > " + Placeholder(1), []any{value}, nil
	case "lt":
		return column + " < " + Placeholder(1), []any{value}, nil
	case "null":
		return column + " IS NULL", nil, nil
	case "notnull":
		return column + " IS NOT NULL", nil, nil
	default:
		return "", nil, fmt.Errorf("unsupported operator: %s", op)
	}
}

// BuildOrderBy builds a safe ORDER BY clause.
func BuildOrderBy(column, dir string) (string, error) {
	if column == "" {
		return "", nil
	}
	if err := validateIdent(column); err != nil {
		return "", err
	}
	dir = strings.ToUpper(strings.TrimSpace(dir))
	if dir != "ASC" && dir != "DESC" {
		dir = "ASC"
	}
	return column + " " + dir, nil
}

// TruncateTable removes all rows from a table.
func TruncateTable(ctx context.Context, table string) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	switch DriverName() {
	case SQLite:
		_, err := DB().ExecContext(ctx, "DELETE FROM "+table)
		return err
	case MySQL, MariaDB:
		_, err := DB().ExecContext(ctx, "TRUNCATE TABLE "+table)
		return err
	default:
		_, err := DB().ExecContext(ctx, "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE")
		return err
	}
}

// ExportTableCSV returns CSV content for a table.
func ExportTableCSV(ctx context.Context, table string) (string, error) {
	if err := validateIdent(table); err != nil {
		return "", err
	}
	rows, err := queryRows(ctx, "SELECT * FROM "+table)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		schema, err := TableSchema(ctx, table)
		if err != nil {
			return "", err
		}
		var cols []string
		for _, c := range schema.Columns {
			cols = append(cols, c.Name)
		}
		return strings.Join(cols, ",") + "\n", nil
	}
	cols := make([]string, 0, len(rows[0]))
	for k := range rows[0] {
		cols = append(cols, k)
	}
	sortStrings(cols)
	var b strings.Builder
	b.WriteString(strings.Join(cols, ","))
	b.WriteByte('\n')
	for _, row := range rows {
		vals := make([]string, len(cols))
		for i, col := range cols {
			vals[i] = csvCell(row[col])
		}
		b.WriteString(strings.Join(vals, ","))
		b.WriteByte('\n')
	}
	return b.String(), nil
}

func csvCell(v any) string {
	if v == nil {
		return ""
	}
	s := fmt.Sprint(v)
	if strings.ContainsAny(s, ",\"\n\r") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return `"` + s + `"`
	}
	return s
}

func sortStrings(ss []string) {
	for i := 0; i < len(ss); i++ {
		for j := i + 1; j < len(ss); j++ {
			if ss[j] < ss[i] {
				ss[i], ss[j] = ss[j], ss[i]
			}
		}
	}
}

// TableCount returns row count for a table.
func TableCount(ctx context.Context, table string) (int64, error) {
	if err := validateIdent(table); err != nil {
		return 0, err
	}
	var count int64
	err := DB().QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	return count, err
}

// FindRow returns a single row by primary key.
func FindRow(ctx context.Context, table string, id string) (map[string]any, error) {
	if err := validateIdent(table); err != nil {
		return nil, err
	}
	schema, err := TableSchema(ctx, table)
	if err != nil {
		return nil, err
	}
	pk := schema.PrimaryKey()
	if err := validateIdent(pk); err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = %s", table, pk, Placeholder(1))
	rows, err := queryRows(ctx, query, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	return rows[0], nil
}

// InsertRow inserts a row from column values.
func InsertRow(ctx context.Context, table string, values map[string]any) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	cols, placeholders, args := buildInsert(values)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	_, err := DB().ExecContext(ctx, query, args...)
	return err
}

// UpdateRow updates a row by primary key.
func UpdateRow(ctx context.Context, table, id string, values map[string]any) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	schema, err := TableSchema(ctx, table)
	if err != nil {
		return err
	}
	pk := schema.PrimaryKey()
	setParts, args := buildUpdate(values)
	args = append(args, id)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = %s", table, strings.Join(setParts, ", "), pk, Placeholder(len(args)))
	_, err = DB().ExecContext(ctx, query, args...)
	return err
}

// DeleteRow deletes a row by primary key.
func DeleteRow(ctx context.Context, table, id string) error {
	if err := validateIdent(table); err != nil {
		return err
	}
	schema, err := TableSchema(ctx, table)
	if err != nil {
		return err
	}
	pk := schema.PrimaryKey()
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = %s", table, pk, Placeholder(1))
	_, err = DB().ExecContext(ctx, query, id)
	return err
}

// DeleteRows deletes multiple rows by primary key values.
func DeleteRows(ctx context.Context, table string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	schema, err := TableSchema(ctx, table)
	if err != nil {
		return err
	}
	pk := schema.PrimaryKey()
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = Placeholder(i + 1)
		args[i] = id
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)", table, pk, strings.Join(placeholders, ", "))
	_, err = DB().ExecContext(ctx, query, args...)
	return err
}

// ExecQuery runs a read-only SQL query (SELECT only).
func ExecQuery(ctx context.Context, sql string) ([]map[string]any, error) {
	trimmed := strings.TrimSpace(strings.ToUpper(sql))
	if !strings.HasPrefix(trimmed, "SELECT") && !strings.HasPrefix(trimmed, "PRAGMA") && !strings.HasPrefix(trimmed, "EXPLAIN") {
		return nil, fmt.Errorf("only SELECT, PRAGMA, and EXPLAIN queries are allowed")
	}
	return queryRows(ctx, sql)
}

func queryRows(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	rows, err := DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]any
	for rows.Next() {
		vals := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any)
		for i, col := range columns {
			row[col] = normalizeValue(vals[i])
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

func normalizeValue(v any) any {
	switch val := v.(type) {
	case []byte:
		return string(val)
	default:
		return v
	}
}

func buildInsert(values map[string]any) (cols []string, placeholders []string, args []any) {
	i := 1
	for col, val := range values {
		if col == "id" && (val == nil || val == "" || val == "0") {
			continue
		}
		if err := validateIdent(col); err != nil {
			continue
		}
		cols = append(cols, col)
		placeholders = append(placeholders, Placeholder(i))
		args = append(args, val)
		i++
	}
	return
}

func buildUpdate(values map[string]any) (setParts []string, args []any) {
	i := 1
	for col, val := range values {
		if col == "id" {
			continue
		}
		if err := validateIdent(col); err != nil {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s = %s", col, Placeholder(i)))
		args = append(args, val)
		i++
	}
	return
}

func validateIdent(name string) error {
	if name == "" || strings.ContainsAny(name, " ;'\"()\\") {
		return fmt.Errorf("invalid identifier: %s", name)
	}
	return nil
}

func listTablesPostgres(ctx context.Context) ([]string, error) {
	query := `SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename`
	rows, err := DB().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStrings(rows)
}

func listTablesSQLite(ctx context.Context) ([]string, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	rows, err := DB().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStrings(rows)
}

func listTablesMySQL(ctx context.Context) ([]string, error) {
	query := `SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() ORDER BY table_name`
	rows, err := DB().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStrings(rows)
}

func tableSchemaPostgres(ctx context.Context, table string) (*TableInfo, error) {
	query := `SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns WHERE table_name = $1 ORDER BY ordinal_position`
	rows, err := DB().QueryContext(ctx, query, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	info := &TableInfo{Name: table}
	for rows.Next() {
		var col ColumnInfo
		var nullable string
		if err := rows.Scan(&col.Name, &col.Type, &nullable, &col.Default); err != nil {
			return nil, err
		}
		col.Nullable = nullable == "YES"
		col.PrimaryKey = col.Name == "id"
		info.Columns = append(info.Columns, col)
	}
	return info, rows.Err()
}

func tableSchemaSQLite(ctx context.Context, table string) (*TableInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", table)
	rows, err := DB().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	info := &TableInfo{Name: table}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		info.Columns = append(info.Columns, ColumnInfo{
			Name: name, Type: ctype, Nullable: notnull == 0,
			PrimaryKey: pk == 1, Default: dflt,
		})
	}
	return info, rows.Err()
}

func tableSchemaMySQL(ctx context.Context, table string) (*TableInfo, error) {
	query := `SELECT column_name, data_type, is_nullable, column_default, column_key
		FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? ORDER BY ordinal_position`
	rows, err := DB().QueryContext(ctx, query, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	info := &TableInfo{Name: table}
	for rows.Next() {
		var col ColumnInfo
		var nullable, colKey string
		if err := rows.Scan(&col.Name, &col.Type, &nullable, &col.Default, &colKey); err != nil {
			return nil, err
		}
		col.Nullable = nullable == "YES"
		col.PrimaryKey = colKey == "PRI"
		info.Columns = append(info.Columns, col)
	}
	return info, rows.Err()
}

func scanStrings(rows *sql.Rows) ([]string, error) {
	var result []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		if s == "schema_migrations" {
			continue
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
