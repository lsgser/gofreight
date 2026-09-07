package database

import (
	"fmt"
	"net/url"
	"strings"
)

// Driver represents a supported database driver.
type Driver string

const (
	Postgres Driver = "postgres"
	SQLite   Driver = "sqlite3"
	MySQL    Driver = "mysql"
	MariaDB  Driver = "mariadb"
)

var currentDriver Driver

// DetectDriver determines the database driver from a connection URL.
func DetectDriver(databaseURL string) (Driver, error) {
	if databaseURL == "" {
		return "", fmt.Errorf("empty database URL")
	}

	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		return Postgres, nil
	}
	if strings.HasPrefix(databaseURL, "sqlite://") {
		return SQLite, nil
	}
	if strings.HasPrefix(databaseURL, "file:") || strings.HasSuffix(databaseURL, ".db") || strings.HasSuffix(databaseURL, ".sqlite") {
		return SQLite, nil
	}
	if strings.HasPrefix(databaseURL, "mysql://") {
		return MySQL, nil
	}
	if strings.HasPrefix(databaseURL, "mariadb://") {
		return MariaDB, nil
	}

	u, err := url.Parse(databaseURL)
	if err == nil {
		switch u.Scheme {
		case "postgres", "postgresql":
			return Postgres, nil
		case "sqlite", "sqlite3":
			return SQLite, nil
		case "mysql":
			return MySQL, nil
		case "mariadb":
			return MariaDB, nil
		}
	}

	return Postgres, nil
}

// DriverName returns the active driver.
func DriverName() Driver {
	return currentDriver
}

// Placeholder returns the parameter placeholder for the given 1-based index.
func Placeholder(index int) string {
	switch currentDriver {
	case SQLite, MySQL, MariaDB:
		return "?"
	default:
		return fmt.Sprintf("$%d", index)
	}
}

// Placeholders returns a comma-separated list of placeholders.
func Placeholders(count int) string {
	parts := make([]string, count)
	for i := 0; i < count; i++ {
		parts[i] = Placeholder(i + 1)
	}
	return strings.Join(parts, ", ")
}

// ReturningClause returns SQL for returning inserted/updated columns.
func ReturningClause(columns ...string) string {
	switch currentDriver {
	case SQLite:
		return ""
	case MySQL, MariaDB:
		return ""
	default:
		return " RETURNING " + strings.Join(columns, ", ")
	}
}

// NowFunc returns the SQL function for current timestamp.
func NowFunc() string {
	switch currentDriver {
	case SQLite:
		return "datetime('now')"
	case MySQL, MariaDB:
		return "NOW()"
	default:
		return "NOW()"
	}
}

// AutoIncrement returns the auto-increment column definition.
func AutoIncrement() string {
	switch currentDriver {
	case SQLite:
		return "INTEGER PRIMARY KEY AUTOINCREMENT"
	case MySQL, MariaDB:
		return "INT AUTO_INCREMENT PRIMARY KEY"
	default:
		return "SERIAL PRIMARY KEY"
	}
}

// TimestampType returns the timestamp column type for migrations.
func TimestampType() string {
	switch currentDriver {
	case SQLite:
		return "TEXT DEFAULT (datetime('now'))"
	case MySQL, MariaDB:
		return "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"
	default:
		return "TIMESTAMP DEFAULT NOW()"
	}
}

// TimestampTzType returns a timezone-aware timestamp column with default (Laravel timestampsTz).
func TimestampTzType() string {
	switch currentDriver {
	case SQLite:
		return "TEXT DEFAULT (datetime('now'))"
	case MySQL, MariaDB:
		return "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"
	default:
		return "TIMESTAMPTZ DEFAULT NOW()"
	}
}

// TimestampColumnType returns the base timestamp type without default.
func TimestampColumnType(tz bool) string {
	switch currentDriver {
	case SQLite:
		return "TEXT"
	case MySQL, MariaDB:
		return "TIMESTAMP"
	default:
		if tz {
			return "TIMESTAMPTZ"
		}
		return "TIMESTAMP"
	}
}

// NullableTimestampType returns a nullable timestamp column (Laravel softDeletes).
func NullableTimestampType() string {
	switch currentDriver {
	case SQLite:
		return "TEXT"
	case MySQL, MariaDB:
		return "TIMESTAMP"
	default:
		return "TIMESTAMP"
	}
}

// NullableTimestampTzType returns a nullable timezone-aware timestamp (Laravel softDeletesTz).
func NullableTimestampTzType() string {
	switch currentDriver {
	case SQLite:
		return "TEXT"
	case MySQL, MariaDB:
		return "TIMESTAMP"
	default:
		return "TIMESTAMPTZ"
	}
}

// OpenPath extracts the file path from a SQLite URL.
func SQLitePath(databaseURL string) string {
	if strings.HasPrefix(databaseURL, "sqlite://") {
		return strings.TrimPrefix(databaseURL, "sqlite://")
	}
	if strings.HasPrefix(databaseURL, "file:") {
		return strings.TrimPrefix(databaseURL, "file:")
	}
	return databaseURL
}

// MySQLDSN converts a mysql:// or mariadb:// URL to a Go MySQL driver DSN.
// The go-sql-driver/mysql driver is wire-compatible with MariaDB.
func MySQLDSN(databaseURL string) string {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return databaseURL
	}
	user := u.User.Username()
	pass, _ := u.User.Password()
	host := u.Host
	dbName := strings.TrimPrefix(u.Path, "/")
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, dbName)
}
