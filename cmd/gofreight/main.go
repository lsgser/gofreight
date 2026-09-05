package main

import (
	"os"
	"strings"
)

func main() {
	initCommands()

	if len(os.Args) < 2 {
		printCLIBanner(os.Stdout)
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
