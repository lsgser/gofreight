package main

import (
	"fmt"
	"os"

	"github.com/lsgser/gofreight/config"
	"github.com/lsgser/gofreight/database"
	"github.com/lsgser/gofreight/dev"
	"github.com/lsgser/gofreight/router"
)

func handleDBSeed() {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Running seeds...")
	if err := database.RunSeeds("db/seeds"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}

func handleRoutes() {
	r := router.New()
	fmt.Println("Register routes in config/routes.go, then use app.Run() to list them at startup.")
	fmt.Println("Registered on empty router:", len(r.Routes()))
}

func handleConsole() {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	RunConsole()
}

func handleWatch(args []string) {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}
	if err := dev.Watch(root, "go", "run", "."); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
