package main

import (
	"fmt"
	"os"
	"os/exec"

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
