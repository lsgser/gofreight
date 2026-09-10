package main

/*
|--------------------------------------------------------------------------
| Console
|--------------------------------------------------------------------------
|
| Implements Console as part of the gofreight package in the Gofreight
| framework. Key symbols: RunConsole.
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
| Symbols defined here include: RunConsole (RunConsole starts an
| interactive SQL console for the connected database.).
| 
*/

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lsgser/gofreight/database"
)

// RunConsole starts an interactive SQL console for the connected database.
func RunConsole() {
	fmt.Println("Gofreight console — type SQL and press Enter (exit to quit)")
	scanner := bufio.NewScanner(os.Stdin)
	ctx := context.Background()
	for {
		fmt.Print("gofreight> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}
		rows, err := database.ExecQuery(ctx, line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		for _, row := range rows {
			fmt.Println(row)
		}
		if len(rows) == 0 {
			fmt.Println("OK")
		}
	}
}
