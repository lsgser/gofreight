package main

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
