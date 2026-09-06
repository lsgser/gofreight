package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lsgser/gofreight/config"
)

func needsProductionGuard(name string) bool {
	switch name {
	case "down", "up", "serve", "dev", "watch", "tinker", "console", "db", "new",
		"db:create", "db:seed", "db:wipe",
		"migrate", "migrate:rollback", "migrate:reset", "migrate:refresh", "migrate:fresh", "db:migrate",
		"cache:clear", "route:clear", "auth:clear-resets",
		"optimize", "optimize:clear", "view:clear", "key:generate",
		"generate", "make":
		return true
	}
	if strings.HasPrefix(name, "make:") {
		return true
	}
	if strings.HasPrefix(name, "queue:") && name != "queue:failed" {
		return true
	}
	return false
}

func guardProduction(commandName string, args []string) bool {
	if hasForceFlag(args) {
		return true
	}

	loadEnv()
	cfg := config.Load()
	if !cfg.IsProduction() {
		return true
	}

	printProductionWarning(os.Stdout, commandName, cfg)
	return confirmProductionContinue(os.Stdin)
}

func hasForceFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			return true
		}
	}
	return false
}

func printProductionWarning(w io.Writer, commandName string, cfg *config.Config) {
	useColor := isTerminal(os.Stdout)
	c := newConsoleStyle(useColor)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\n", c.red("  ╭──────────────────────────────────────────────────────────────╮"))
	fmt.Fprintf(w, "%s\n", productionLine(c, "  ⚠  PRODUCTION ENVIRONMENT"))
	fmt.Fprintf(w, "%s\n", productionLine(c, "  Command: "+commandName))
	if cfg.Database.Database != "" {
		target := cfg.Database.Connection + "/" + cfg.Database.Database
		if cfg.Database.Host != "" {
			target += " @ " + cfg.Database.Host
		}
		fmt.Fprintf(w, "%s\n", productionLine(c, "  Database: "+target))
	}
	fmt.Fprintf(w, "%s\n", c.red("  ╰──────────────────────────────────────────────────────────────╯"))
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\n", c.yellow("  This command can modify your live application."))
	fmt.Fprintf(w, "%s\n", c.dim("  Type \"yes\" to continue, or pass --force to skip this prompt."))
	fmt.Fprint(w, c.bold("\n  Continue? "))
}

func productionLine(c consoleStyle, text string) string {
	inner := 62
	content := text
	if strings.HasPrefix(text, "  ⚠") {
		content = c.red("  ⚠") + strings.TrimPrefix(text, "  ⚠")
	}
	return c.red("  │") + content + pad(inner-visibleLen(content)) + c.red("│")
}

func confirmProductionContinue(in io.Reader) bool {
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(scanner.Text()), "yes")
}
