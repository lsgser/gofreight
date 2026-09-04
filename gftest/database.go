package gftest

import (
	"context"
	"fmt"
	"testing"

	"github.com/lsgser/gofreight/database"
)

// DatabaseAssertions provides database-level test assertions.
type DatabaseAssertions struct {
	T *testing.T
}

// ToHaveCount asserts a table has the expected row count.
func (d *DatabaseAssertions) ToHaveCount(table string, expected int64) *DatabaseAssertions {
	d.T.Helper()
	count, err := d.count(table)
	if err != nil {
		d.T.Fatalf("count %s: %v", table, err)
	}
	Expect(count).Bind(d.T).ToEqual(expected)
	return d
}

// ToHaveRecord asserts a record exists matching conditions.
func (d *DatabaseAssertions) ToHaveRecord(table string, column string, value any) *DatabaseAssertions {
	d.T.Helper()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = %s", table, column, database.Placeholder(1))
	var count int64
	if err := database.DB().QueryRowContext(context.Background(), query, value).Scan(&count); err != nil {
		d.T.Fatalf("query: %v", err)
	}
	Expect(count > 0).Bind(d.T).ToBeTrue()
	return d
}

// ToBeEmpty asserts a table has no rows.
func (d *DatabaseAssertions) ToBeEmpty(table string) *DatabaseAssertions {
	return d.ToHaveCount(table, 0)
}

func (d *DatabaseAssertions) count(table string) (int64, error) {
	var count int64
	err := database.DB().QueryRowContext(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	return count, err
}

// RefreshDatabase resets and re-runs migrations (like Laravel RefreshDatabase trait).
func RefreshDatabase(t *testing.T, databaseURL string, migrations ...string) {
	t.Helper()
	database.Reset()
	if _, err := database.Connect(databaseURL); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if len(migrations) > 0 {
		if err := database.Migrate(migrations...); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}
}

// RunMigrations runs SQL files from a directory.
func RunMigrations(dir string) error {
	m := database.NewMigrator(dir)
	return m.Up()
}

// TruncateTables removes all rows from the given tables.
func TruncateTables(t *testing.T, tables ...string) {
	t.Helper()
	for _, table := range tables {
		if _, err := database.DB().Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
}

// UseTransactionalTests wraps each test in a transaction that rolls back.
// Note: SQLite in-memory supports this pattern for test isolation.
func UseTransactionalTests(t *testing.T, fn func(t *testing.T)) {
	t.Helper()
	tx, err := database.DB().Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()
	fn(t)
}

// AssertDatabaseCount is a standalone assertion helper.
func AssertDatabaseCount(t *testing.T, table string, expected int64) {
	t.Helper()
	(&DatabaseAssertions{T: t}).ToHaveCount(table, expected)
}
