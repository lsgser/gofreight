package database

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
