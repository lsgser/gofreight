package main

/*
|--------------------------------------------------------------------------
| Migration Tool Entry Point
|--------------------------------------------------------------------------
|
| Application main: loads bootstrap, registers routes, and starts the HTTP
| server.
| 
| Use gofreight serve for development; production uses a binary from
| gofreight build.
| 
*/

import (
	"fmt"
	"os"

	_ "demoapp/db/migrate"
	"github.com/joho/godotenv"
	"github.com/lsgser/gofreight/config"
	"github.com/lsgser/gofreight/database"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if _, err := database.Connect(cfg.DatabaseURL); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("usage: migrate [up|down|status|reset|refresh|fresh]")
		os.Exit(1)
	}

	m := database.NewGoMigrator()
	var err error
	switch os.Args[1] {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "status":
		var statuses []database.MigrationStatus
		statuses, err = m.Status()
		if err == nil {
			for _, s := range statuses {
				state := "Pending"
				if s.Applied {
					state = "Ran"
				}
				fmt.Printf("  [%s] %s\n", state, s.Version)
			}
		}
	case "reset":
		err = m.Reset()
	case "refresh":
		err = m.Refresh()
	case "fresh":
		err = m.Fresh()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
