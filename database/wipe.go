package database

import (
	"fmt"
	"strings"
)

// Wipe drops all user tables in the connected database.
func Wipe() error {
	switch currentDriver {
	case SQLite:
		return wipeSQLite()
	case Postgres:
		return wipePostgreSQL()
	case MySQL, MariaDB:
		return wipeMySQL()
	default:
		return fmt.Errorf("db:wipe not supported for driver %q", currentDriver)
	}
}

func wipeSQLite() error {
	rows, err := DB().Query(`
		SELECT name FROM sqlite_master
		WHERE type = 'table' AND name NOT LIKE 'sqlite_%'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, table := range tables {
		if _, err := DB().Exec("DROP TABLE IF EXISTS " + quoteIdentSQLite(table)); err != nil {
			return fmt.Errorf("drop %s: %w", table, err)
		}
		fmt.Printf("  dropped: %s\n", table)
	}
	return nil
}

func wipePostgreSQL() error {
	rows, err := DB().Query(`
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, table := range tables {
		stmt := fmt.Sprintf(`DROP TABLE IF EXISTS "%s" CASCADE`, strings.ReplaceAll(table, `"`, `""`))
		if _, err := DB().Exec(stmt); err != nil {
			return fmt.Errorf("drop %s: %w", table, err)
		}
		fmt.Printf("  dropped: %s\n", table)
	}
	return nil
}

func wipeMySQL() error {
	rows, err := DB().Query("SHOW TABLES")
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	DB().Exec("SET FOREIGN_KEY_CHECKS = 0")
	defer DB().Exec("SET FOREIGN_KEY_CHECKS = 1")

	for _, table := range tables {
		stmt := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", strings.ReplaceAll(table, "`", "``"))
		if _, err := DB().Exec(stmt); err != nil {
			return fmt.Errorf("drop %s: %w", table, err)
		}
		fmt.Printf("  dropped: %s\n", table)
	}
	return nil
}

func quoteIdentSQLite(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
