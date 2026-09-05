package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lsgser/gofreight/config"
	"github.com/lsgser/gofreight/database"
	"github.com/lsgser/gofreight/generator"
	"github.com/lsgser/gofreight/version"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "new":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gofreight new <app_name>")
			os.Exit(1)
		}
		if err := generator.NewApp(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "generate", "g":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gofreight generate <type> <name> [field:type ...]")
			os.Exit(1)
		}
		handleGenerate(os.Args[2:])

	case "db:migrate":
		handleDBMigrate()

	case "db:rollback":
		handleDBRollback()

	case "db:status":
		handleDBStatus()

	case "db:create":
		handleDBCreate()

	case "test":
		handleTest(os.Args[2:])

	case "db:seed":
		handleDBSeed()

	case "routes":
		handleRoutes()

	case "console":
		handleConsole()

	case "watch":
		handleWatch(os.Args[2:])

	case "make":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gofreight make <type> <name> [field:type ...]")
			os.Exit(1)
		}
		handleGenerate(os.Args[2:])

	case "version", "-v":
		fmt.Println("gofreight v" + version.Version)

	default:
		printUsage()
		os.Exit(1)
	}
}

func handleGenerate(args []string) {
	genType := args[0]
	name := args[1]
	appPath := "."

	fields := parseFields(args[2:])

	switch genType {
	case "model":
		if err := generator.Model(appPath, name, fields); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated model: %s\n", name)

	case "controller":
		if err := generator.Controller(appPath, name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated controller: %s\n", name)

	case "migration":
		dir := filepath.Join(appPath, "db", "migrate")
		path, err := database.CreateMigration(dir, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created migration: %s\n", path)

	case "scaffold", "resource":
		if err := generator.Resource(appPath, name, fields); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated scaffold: %s\n", name)

	case "auth":
		if err := generator.Auth(appPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Generated auth (User model, seeds, route snippets)")

	default:
		fmt.Printf("Unknown generator: %s\n", genType)
		fmt.Println("Available: model, controller, migration, scaffold, auth")
		os.Exit(1)
	}
}

func handleDBMigrate() {
	loadEnv()
	cfg := config.Load()

	if err := connectDB(cfg.DatabaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	migrator := database.NewMigrator("db/migrate")
	fmt.Println("Running migrations...")
	if err := migrator.Up(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}

func handleDBRollback() {
	loadEnv()
	cfg := config.Load()

	if err := connectDB(cfg.DatabaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	migrator := database.NewMigrator("db/migrate")
	fmt.Println("Rolling back last migration...")
	if err := migrator.Down(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}

func handleDBStatus() {
	loadEnv()
	cfg := config.Load()

	if err := connectDB(cfg.DatabaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	migrator := database.NewMigrator("db/migrate")
	statuses, err := migrator.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(statuses) == 0 {
		fmt.Println("No migrations found.")
		return
	}

	fmt.Printf("%-40s %s\n", "Migration", "Status")
	fmt.Println(strings.Repeat("-", 55))
	for _, s := range statuses {
		status := "down"
		if s.Applied {
			status = "up"
		}
		fmt.Printf("%-40s %s\n", s.File, status)
	}
}

func handleDBCreate() {
	loadEnv()
	cfg := config.Load()

	driver, err := database.DetectDriver(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch driver {
	case database.SQLite:
		path := database.SQLitePath(cfg.DatabaseURL)
		dir := filepath.Dir(path)
		if dir != "." {
			os.MkdirAll(dir, 0755)
		}
		f, err := os.Create(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		f.Close()
		fmt.Printf("Created SQLite database: %s\n", path)
	default:
		fmt.Printf("Database driver: %s\n", driver)
		fmt.Println("Ensure the database exists on your server, then run: gofreight db:migrate")
	}
}

func handleTest(args []string) {
	cmdArgs := append([]string{"test", "./..."}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

func connectDB(url string) error {
	database.Reset()
	_, err := database.Connect(url)
	return err
}

func loadEnv() {
	_ = godotenv.Load()
}

func parseFields(args []string) map[string]string {
	fields := make(map[string]string)
	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) == 2 {
			fields[parts[0]] = parts[1]
		}
	}
	return fields
}

func printUsage() {
	fmt.Println(`Gofreight — A Ruby on Rails-inspired web framework for Go

Usage:
  gofreight new <app_name>                         Create a new application
  gofreight generate model <name> [fields]         Generate a model
  gofreight generate controller <name>             Generate a controller
  gofreight generate migration <name>              Generate a migration
  gofreight generate scaffold <name> [fields]      Full CRUD resource (Laravel-style)
  gofreight generate resource <name> [fields]      Alias for scaffold
  gofreight make scaffold Post title:string        Laravel-style alias
  gofreight generate auth                          Generate User model + auth setup
  gofreight db:create                              Create the database
  gofreight db:migrate                             Run pending migrations
  gofreight db:rollback                            Rollback last migration
  gofreight db:status                              Show migration status
  gofreight db:seed                                Run db/seeds/*.sql
  gofreight routes                                 Route helper info
  gofreight console                                Interactive SQL console
  gofreight watch [dir]                            Reload on file changes (dev)
  gofreight test [packages]                        Run the test suite
  gofreight version                                Show version

Examples:
  gofreight new blog
  gofreight make scaffold Post title:string body:text status:enum:draft,published
  gofreight generate model Post title:string body:text
  gofreight generate migration add_published_to_posts
  gofreight db:migrate

Field types: string, text, email, url, integer, bigint, float, boolean, datetime, date, time, uuid, json, enum, references
  See: https://github.com/lsgser/gofreight/blob/main/docs/generators.md`)
}
