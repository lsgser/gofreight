package main

/*
|--------------------------------------------------------------------------
| Support
|--------------------------------------------------------------------------
|
| Implements Support as part of the gofreight package in the Gofreight
| framework.
| 
| This directory contains the gofreight CLI binary: command registration,
| terminal UI, and handlers for make:*, migrate, serve, test, and
| mail:preview.
| 
| Each subcommand lives in its own source file; commands.go registers the
| catalog shown by gofreight list.
| 
| Install locally with go install ./cmd/gofreight from the framework
| repository root.
| 
*/

import (
	"fmt"
	"os"

	"github.com/lsgser/gofreight/config"
	"github.com/lsgser/gofreight/database"
	"github.com/lsgser/gofreight/dev"
	"github.com/joho/godotenv"
)

func handleConsole(args []string) {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fail(err)
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

func connectDB(url string) error {
	database.Reset()
	_, err := database.Connect(url)
	return err
}

func loadEnv() {
	_ = godotenv.Load()
}
