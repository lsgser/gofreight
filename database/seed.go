package database

/*
|--------------------------------------------------------------------------
| Seed
|--------------------------------------------------------------------------
|
| Implements Seed as part of the database package in the Gofreight
| framework. Key symbols: RunSeeds.
| 
| The database package manages connections, fluent schema blueprints, Go
| and SQL migrations, seeding, and introspection.
| 
| Migrations run via gofreight migrate; blueprints generate portable DDL
| across SQLite, PostgreSQL, and MySQL.
| 
| Lower-level query helpers complement the model ORM in the model package.
| 
| Symbols defined here include: RunSeeds (RunSeeds executes all SQL files
| in db/seeds/ directory.).
| 
*/

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RunSeeds executes all SQL files in db/seeds/ directory.
func RunSeeds(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no seeds directory at %s", dir)
		}
		return err
	}
	ctx := context.Background()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		if err := ImportSQL(ctx, string(content)); err != nil {
			return fmt.Errorf("seed %s: %w", e.Name(), err)
		}
		fmt.Printf("  seeded: %s\n", e.Name())
	}
	return nil
}
