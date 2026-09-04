package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const schemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version VARCHAR(255) PRIMARY KEY,
	applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

// Migrator runs SQL migration files from a directory.
type Migrator struct {
	Dir string
}

// NewMigrator creates a migrator for the given migrations directory.
func NewMigrator(dir string) *Migrator {
	return &Migrator{Dir: dir}
}

// Up runs all pending migrations.
func (m *Migrator) Up() error {
	if err := m.ensureTable(); err != nil {
		return err
	}

	files, err := m.migrationFiles()
	if err != nil {
		return err
	}

	applied, err := m.appliedVersions()
	if err != nil {
		return err
	}

	for _, file := range files {
		version := migrationVersion(file)
		if applied[version] {
			continue
		}

		content, err := os.ReadFile(filepath.Join(m.Dir, file))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		tx, err := DB().Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", file, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version) VALUES ("+Placeholder(1)+")",
			version,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		fmt.Printf("  migrated: %s\n", file)
	}

	return nil
}

// Down rolls back the most recent migration.
func (m *Migrator) Down() error {
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
	downFile := filepath.Join(m.Dir, lastVersion+"_down.sql")

	tx, err := DB().Begin()
	if err != nil {
		return err
	}

	if content, err := os.ReadFile(downFile); err == nil {
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("rollback %s failed: %w", lastVersion, err)
		}
	} else {
		fmt.Printf("  warning: no down file for %s, removing version only\n", lastVersion)
	}

	if _, err := tx.Exec(
		"DELETE FROM schema_migrations WHERE version = "+Placeholder(1),
		lastVersion,
	); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("  rolled back: %s\n", lastVersion)
	return nil
}

// Status prints the status of all migrations.
func (m *Migrator) Status() ([]MigrationStatus, error) {
	if err := m.ensureTable(); err != nil {
		return nil, err
	}

	files, err := m.migrationFiles()
	if err != nil {
		return nil, err
	}

	applied, err := m.appliedVersions()
	if err != nil {
		return nil, err
	}

	var statuses []MigrationStatus
	for _, file := range files {
		version := migrationVersion(file)
		statuses = append(statuses, MigrationStatus{
			Version: version,
			File:    file,
			Applied: applied[version],
		})
	}
	return statuses, nil
}

// MigrationStatus describes a single migration file.
type MigrationStatus struct {
	Version string
	File    string
	Applied bool
}

func (m *Migrator) ensureTable() error {
	_, err := DB().Exec(schemaMigrationsTable)
	return err
}

func (m *Migrator) migrationFiles() ([]string, error) {
	entries, err := os.ReadDir(m.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".sql") && !strings.HasSuffix(name, "_down.sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)
	return files, nil
}

func (m *Migrator) appliedVersions() (map[string]bool, error) {
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

func (m *Migrator) appliedVersionsList() ([]string, error) {
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

func migrationVersion(filename string) string {
	return strings.TrimSuffix(filename, ".sql")
}

// CreateMigration creates up and down migration files with timestamps.
func CreateMigration(dir, name string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	version := time.Now().Format("20060102150405")
	base := fmt.Sprintf("%s_%s", version, name)

	upPath := filepath.Join(dir, base+".sql")
	downPath := filepath.Join(dir, base+"_down.sql")

	upContent := fmt.Sprintf("-- Migration: %s\n\n", name)
	downContent := fmt.Sprintf("-- Rollback: %s\n\n", name)

	if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
		return "", err
	}
	if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
		return "", err
	}

	return upPath, nil
}
