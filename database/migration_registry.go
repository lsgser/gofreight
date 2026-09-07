package database

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
