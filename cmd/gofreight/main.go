package main

/*
|--------------------------------------------------------------------------
| Gofreight CLI Entry Point
|--------------------------------------------------------------------------
|
| Entry point for the gofreight CLI binary.
| 
| Parses os.Args, initializes the command registry, and dispatches to the
| matching handler or prints usage.
| 
| Install with go install ./cmd/gofreight; subcommand implementations live
| alongside this file.
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
	"os"
	"strings"
)

func main() {
	initCommands()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	name := os.Args[1]
	args := os.Args[2:]

	if !dispatch(name, args) {
		printUsage()
		os.Exit(1)
	}
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
