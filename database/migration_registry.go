package database

/*
|--------------------------------------------------------------------------
| Migration Registry
|--------------------------------------------------------------------------
|
| Implements Migration Registry as part of the database package in the
| Gofreight framework. Key symbols: RegisteredMigration,
| RegisterMigration, RegisteredMigrations, ResetRegisteredMigrations.
| 
| The database package manages connections, fluent schema blueprints, Go
| and SQL migrations, seeding, and introspection.
| 
| Migrations run via gofreight migrate; blueprints generate portable DDL
| across SQLite, PostgreSQL, and MySQL.
| 
| Lower-level query helpers complement the model ORM in the model package.
| 
| Symbols defined here include: RegisteredMigration (exported type);
| RegisterMigration (RegisterMigration registers a Go migration. Call from
| init() in db/migrate/*.go files.); RegisteredMigrations
| (RegisteredMigrations returns a sorted copy of all registered
| migrations.); ResetRegisteredMigrations (ResetRegisteredMigrations
| clears the registry (for tests).).
| 
*/

import (
	"context"
	"sort"
	"sync"
)

// RegisteredMigration is a programmatic migration (Laravel-style blueprint).
type RegisteredMigration struct {
	Version string
	Up      func(context.Context) error
	Down    func(context.Context) error
}

var (
	migrationMu        sync.RWMutex
	registeredMigrations []RegisteredMigration
)

// RegisterMigration registers a Go migration. Call from init() in db/migrate/*.go files.
func RegisterMigration(version string, up, down func(context.Context) error) {
	migrationMu.Lock()
	defer migrationMu.Unlock()
	registeredMigrations = append(registeredMigrations, RegisteredMigration{
		Version: version,
		Up:      up,
		Down:    down,
	})
}

// RegisteredMigrations returns a sorted copy of all registered migrations.
func RegisteredMigrations() []RegisteredMigration {
	migrationMu.RLock()
	defer migrationMu.RUnlock()
	out := append([]RegisteredMigration(nil), registeredMigrations...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Version < out[j].Version
	})
	return out
}

// ResetRegisteredMigrations clears the registry (for tests).
func ResetRegisteredMigrations() {
	migrationMu.Lock()
	defer migrationMu.Unlock()
	registeredMigrations = nil
}
