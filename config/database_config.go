package config

/*
|--------------------------------------------------------------------------
| Database Config
|--------------------------------------------------------------------------
|
| Implements Database Config as part of the config package in the
| Gofreight framework. Key symbols: DatabaseConfig, LoadDatabaseConfig,
| ResolveDatabaseURL, URL.
| 
| The config package loads .env files, resolves MAIL_DRIVER and database
| URLs, and reads config/app.yaml.
| 
| Database drivers and app keys are validated early so misconfiguration
| fails fast at boot.
| 
| Application code reads config through helpers rather than os.Getenv
| scattered across the codebase.
| 
| Symbols defined here include: DatabaseConfig (exported type);
| LoadDatabaseConfig (LoadDatabaseConfig reads database settings from file
| config and environment.); ResolveDatabaseURL (URL returns a connection
| URL for database.Connect.); URL (URL builds a driver connection URL from
| discrete settings.).
| 
*/

import (
	"fmt"
	"net/url"
	"strings"
)

// DatabaseConfig holds discrete database connection settings (DB_* env vars).
type DatabaseConfig struct {
	Connection string
	Host       string
	Port       string
	Database   string
	Username   string
	Password   string
	SSLMode    string
}

// LoadDatabaseConfig reads database settings from file config and environment.
func LoadDatabaseConfig(files *FileConfig) DatabaseConfig {
	conn := strings.ToLower(firstNonEmpty(
		osGet("DB_CONNECTION"),
		filesGetString(files, "db_connection", ""),
	))
	if conn == "" {
		conn = "sqlite"
	}

	return DatabaseConfig{
		Connection: conn,
		Host:       firstNonEmpty(osGet("DB_HOST"), filesGetString(files, "db_host", "127.0.0.1")),
		Port:       firstNonEmpty(osGet("DB_PORT"), filesGetString(files, "db_port", "")),
		Database:   firstNonEmpty(osGet("DB_DATABASE"), filesGetString(files, "db_database", "db/development.db")),
		Username:   firstNonEmpty(osGet("DB_USERNAME"), filesGetString(files, "db_username", "")),
		Password:   firstNonEmpty(osGet("DB_PASSWORD"), filesGetString(files, "db_password", "")),
		SSLMode:    firstNonEmpty(osGet("DB_SSLMODE"), filesGetString(files, "db_sslmode", "disable")),
	}
}

// URL returns a connection URL for database.Connect.
//
// Resolution order:
//  1. DATABASE_URL — full URL (backward compatible)
//  2. DB_URL — single URL override
//  3. DB_CONNECTION + DB_HOST, DB_PORT, DB_DATABASE, DB_USERNAME, DB_PASSWORD
func ResolveDatabaseURL(files *FileConfig) string {
	if v := osGet("DATABASE_URL"); v != "" {
		return v
	}
	if v := osGet("DB_URL"); v != "" {
		return v
	}
	return LoadDatabaseConfig(files).URL()
}

// URL builds a driver connection URL from discrete settings.
func (d DatabaseConfig) URL() string {
	switch normalizeConnection(d.Connection) {
	case "sqlite", "sqlite3":
		path := d.Database
		if path == "" {
			path = "db/development.db"
		}
		if strings.HasPrefix(path, "sqlite://") {
			return path
		}
		return "sqlite://" + path
	case "pgsql", "postgres", "postgresql":
		return d.buildNetworkURL("postgres")
	case "mysql":
		return d.buildNetworkURL("mysql")
	case "mariadb":
		return d.buildNetworkURL("mariadb")
	default:
		if d.Connection != "" {
			return d.buildNetworkURL(d.Connection)
		}
		return "sqlite://db/development.db"
	}
}

func (d DatabaseConfig) buildNetworkURL(scheme string) string {
	host := d.Host
	if host == "" {
		host = "127.0.0.1"
	}

	port := d.Port
	if port == "" {
		switch scheme {
		case "postgres":
			port = "5432"
		case "mysql", "mariadb":
			port = "3306"
		}
	}

	user := url.UserPassword(d.Username, d.Password)
	if d.Username == "" && d.Password == "" {
		user = url.User("")
	} else if d.Password == "" {
		user = url.User(d.Username)
	}

	dbName := strings.TrimPrefix(d.Database, "/")
	u := &url.URL{
		Scheme: scheme,
		User:   user,
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   "/" + dbName,
	}

	if scheme == "postgres" {
		q := u.Query()
		if d.SSLMode != "" {
			q.Set("sslmode", d.SSLMode)
		}
		u.RawQuery = q.Encode()
	}

	return u.String()
}

func normalizeConnection(conn string) string {
	return strings.ToLower(strings.TrimSpace(conn))
}

func osGet(key string) string {
	return strings.TrimSpace(getEnv(key, ""))
}

func filesGetString(files *FileConfig, key, fallback string) string {
	if files == nil {
		return fallback
	}
	return files.GetString(key, fallback)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
