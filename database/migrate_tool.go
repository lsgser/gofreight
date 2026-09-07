package database

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const migrateToolPath = "tools/migrate/main.go"

// UsesGoMigrations reports whether the app uses the Go blueprint migration runner.
func UsesGoMigrations(appDir string) bool {
	_, err := os.Stat(filepath.Join(appDir, migrateToolPath))
	return err == nil
}

// RunGoMigrateTool runs tools/migrate with the given subcommand (up, down, etc.).
func RunGoMigrateTool(appDir, command string) error {
	tool := filepath.Join(appDir, migrateToolPath)
	if _, err := os.Stat(tool); err != nil {
		return fmt.Errorf("Go migration runner not found at %s — run gofreight new or add tools/migrate/main.go", migrateToolPath)
	}

	cmd := exec.Command("go", "run", "./tools/migrate", command)
	cmd.Dir = appDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
