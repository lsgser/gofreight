package database

/*
|--------------------------------------------------------------------------
| Go Migrate
|--------------------------------------------------------------------------
|
| Implements Go Migrate as part of the database package in the Gofreight
| framework. Key symbols: GoMigrator, NewGoMigrator, Up, Down, Reset,
| Refresh.
| 
| The database package manages connections, fluent schema blueprints, Go
| and SQL migrations, seeding, and introspection.
| 
| Migrations run via gofreight migrate; blueprints generate portable DDL
| across SQLite, PostgreSQL, and MySQL.
| 
| Lower-level query helpers complement the model ORM in the model package.
| 
| Symbols defined here include: GoMigrator (exported type); NewGoMigrator
| (NewGoMigrator creates a migrator for registered Go migrations.); Up (Up
| runs all pending registered migrations.); Down (Down rolls back the most
| recent registered migration.); Reset (Reset rolls back all applied
| registered migrations.); Refresh (Refresh resets and re-runs all
| registered migrations.); Fresh (Fresh drops all tables and re-runs
| registered migrations.); Status (Status returns migration status for
| registered migrations.).
| 
*/

import (
	"context"
	"fmt"
)

// GoMigrator runs registered Go blueprint migrations.
type GoMigrator struct{}

// NewGoMigrator creates a migrator for registered Go migrations.
func NewGoMigrator() *GoMigrator {
	return &GoMigrator{}
}

// Up runs all pending registered migrations.
func (m *GoMigrator) Up() error {
	if err := m.ensureTable(); err != nil {
		return err
	}

	migrations := RegisteredMigrations()
	if len(migrations) == 0 {
		fmt.Println("  No Go migrations registered — create one with gofreight make:migration")
		return nil
	}

	applied, err := m.appliedVersions()
	if err != nil {
		return err
	}

	var ran int
	ctx := context.Background()
	for _, mig := range migrations {
		if applied[mig.Version] {
			continue
		}
		if mig.Up == nil {
			return fmt.Errorf("migration %s has no Up function", mig.Version)
		}

		tx, err := DB().Begin()
		if err != nil {
			return err
		}

		if err := mig.Up(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", mig.Version, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version) VALUES ("+Placeholder(1)+")",
			mig.Version,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", mig.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		fmt.Printf("  migrated: %s\n", mig.Version)
		ran++
	}

	if ran == 0 {
		fmt.Println("  Nothing to migrate — database is up to date.")
	}
	return nil
}

// Down rolls back the most recent registered migration.
func (m *GoMigrator) Down() error {
	if err := m.ensureTable(); err != nil {
		return err
	}

	applied, err := m.appliedVersionsList()
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		fmt.Println("  no migrations to rollback")
		return nil
	}

	lastVersion := applied[len(applied)-1]
	mig := findRegisteredMigration(lastVersion)
	if mig == nil {
		fmt.Printf("  warning: no registered migration for %s, removing version only\n", lastVersion)
	} else if mig.Down != nil {
		ctx := context.Background()
		tx, err := DB().Begin()
		if err != nil {
			return err
		}
		if err := mig.Down(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("rollback %s failed: %w", lastVersion, err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	if _, err := DB().Exec(
		"DELETE FROM schema_migrations WHERE version = "+Placeholder(1),
		lastVersion,
	); err != nil {
		return err
	}

	fmt.Printf("  rolled back: %s\n", lastVersion)
	return nil
}

// Reset rolls back all applied registered migrations.
func (m *GoMigrator) Reset() error {
	for {
		applied, err := m.appliedVersionsList()
		if err != nil {
			return err
		}
		if len(applied) == 0 {
			return nil
		}
		if err := m.Down(); err != nil {
			return err
		}
	}
}

// Refresh resets and re-runs all registered migrations.
func (m *GoMigrator) Refresh() error {
	if err := m.Reset(); err != nil {
		return err
	}
	return m.Up()
}

// Fresh drops all tables and re-runs registered migrations.
func (m *GoMigrator) Fresh() error {
	if err := Wipe(); err != nil {
		return err
	}
	return m.Up()
}

// Status returns migration status for registered migrations.
func (m *GoMigrator) Status() ([]MigrationStatus, error) {
	if err := m.ensureTable(); err != nil {
		return nil, err
	}

	applied, err := m.appliedVersions()
	if err != nil {
		return nil, err
	}

	migrations := RegisteredMigrations()
	statuses := make([]MigrationStatus, 0, len(migrations))
	for _, mig := range migrations {
		statuses = append(statuses, MigrationStatus{
			Version: mig.Version,
			File:    mig.Version + ".go",
			Applied: applied[mig.Version],
		})
	}
	return statuses, nil
}

func findRegisteredMigration(version string) *RegisteredMigration {
	for _, mig := range RegisteredMigrations() {
		if mig.Version == version {
			m := mig
			return &m
		}
	}
	return nil
}

func (m *GoMigrator) ensureTable() error {
	_, err := DB().Exec(schemaMigrationsTable)
	return err
}

func (m *GoMigrator) appliedVersions() (map[string]bool, error) {
	rows, err := DB().Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func (m *GoMigrator) appliedVersionsList() ([]string, error) {
	rows, err := DB().Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}
