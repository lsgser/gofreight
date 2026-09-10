package database

/*
|--------------------------------------------------------------------------
| Database
|--------------------------------------------------------------------------
|
| Implements Database as part of the database package in the Gofreight
| framework. Key symbols: Connect, DB, Close, Reset, Migrate.
| 
| The database package manages connections, fluent schema blueprints, Go
| and SQL migrations, seeding, and introspection.
| 
| Migrations run via gofreight migrate; blueprints generate portable DDL
| across SQLite, PostgreSQL, and MySQL.
| 
| Lower-level query helpers complement the model ORM in the model package.
| 
| Symbols defined here include: Connect (Connect establishes a database
| connection pool, auto-detecting the driver.); DB (DB returns the active
| database connection.); Close (Close closes the database connection
| pool.); Reset (Reset closes and clears the connection (for testing).);
| Migrate (Migrate runs SQL migration statements in order.).
| 
*/

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var (
	db   *sql.DB
	once sync.Once
)

// Connect establishes a database connection pool, auto-detecting the driver.
func Connect(databaseURL string) (*sql.DB, error) {
	var err error
	once.Do(func() {
		driver, derr := DetectDriver(databaseURL)
		if derr != nil {
			err = derr
			return
		}
		currentDriver = driver

		dsn := databaseURL
		driverName := string(driver)

		switch driver {
		case SQLite:
			driverName = "sqlite"
			dsn = SQLitePath(databaseURL)
		case MySQL, MariaDB:
			driverName = "mysql" // go-sql-driver/mysql works with MariaDB
			dsn = MySQLDSN(databaseURL)
		}

		db, err = sql.Open(driverName, dsn)
		if err != nil {
			return
		}
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		err = db.Ping()
	})
	return db, err
}

// DB returns the active database connection.
func DB() *sql.DB {
	if db == nil {
		panic("database not connected — call database.Connect() first")
	}
	return db
}

// Close closes the database connection pool.
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// Reset closes and clears the connection (for testing).
func Reset() error {
	once = sync.Once{}
	if db != nil {
		err := db.Close()
		db = nil
		currentDriver = ""
		return err
	}
	return nil
}

// Migrate runs SQL migration statements in order.
func Migrate(migrations ...string) error {
	for i, stmt := range migrations {
		if _, err := DB().Exec(stmt); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}
	return nil
}
