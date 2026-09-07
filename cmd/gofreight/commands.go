package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type command struct {
	Name        string
	Aliases     []string
	Category    string
	Description string
	Usage       string
	Run         func(args []string)
	Hidden      bool
}

var allCommands []command

func register(c command) {
	allCommands = append(allCommands, c)
}

func initCommands() {
	register(command{
		Name: "about", Category: "app", Description: "Display basic information about your application",
		Run: handleAbout,
	})
	register(command{
		Name: "list", Aliases: []string{"commands"}, Category: "app", Description: "List all available commands",
		Run: func(args []string) { printCommandList("") },
	})
	register(command{
		Name: "help", Aliases: []string{"-h", "--help"}, Category: "app", Description: "Display help for a command",
		Usage: "gofreight help <command>", Run: handleHelp,
	})
	register(command{
		Name: "new", Category: "app", Description: "Create a new Gofreight application",
		Usage: "gofreight new <app_name>", Run: handleNew,
	})
	register(command{
		Name: "serve", Category: "app", Description: "Serve the application on the development server",
		Usage: "gofreight serve", Run: handleServe,
	})
	register(command{
		Name: "test", Category: "app", Description: "Run the application tests",
		Usage: "gofreight test [packages]", Run: handleTest,
	})
	register(command{
		Name: "dev", Category: "app", Description: "Run the dev server with file watching",
		Usage: "gofreight dev [dir]", Run: handleWatch,
	})
	register(command{
		Name: "watch", Category: "app", Description: "Reload on file changes (alias for dev)",
		Usage: "gofreight watch [dir]", Run: handleWatch, Hidden: true,
	})
	register(command{
		Name: "tinker", Category: "app", Description: "Interact with your application (SQL console)",
		Run: handleConsole,
	})
	register(command{
		Name: "console", Category: "app", Description: "Interactive SQL console (alias for tinker)",
		Run: handleConsole, Hidden: true,
	})
	register(command{
		Name: "env", Category: "app", Description: "Display the current framework environment",
		Run: handleEnv,
	})
	register(command{
		Name: "inspire", Category: "app", Description: "Display an inspiring quote",
		Run: handleInspire,
	})
	register(command{
		Name: "down", Category: "app", Description: "Put the application into maintenance mode",
		Run: handleDown,
	})
	register(command{
		Name: "up", Category: "app", Description: "Bring the application out of maintenance mode",
		Run: handleUp,
	})
	register(command{
		Name: "version", Aliases: []string{"-v", "--version"}, Category: "app", Description: "Show the Gofreight version",
		Run: func(args []string) { handleVersion() },
	})

	register(command{
		Name: "db", Category: "db", Description: "Start a new database CLI session",
		Run: handleConsole,
	})
	register(command{
		Name: "db:create", Category: "db", Description: "Create the database",
		Run: func(args []string) { handleDBCreate() },
	})
	register(command{
		Name: "db:seed", Category: "db", Description: "Seed the database with records",
		Usage: "gofreight db:seed [--class=SeederName]", Run: handleDBSeed,
	})
	register(command{
		Name: "db:wipe", Category: "db", Description: "Drop all tables, views, and types",
		Run: handleDBWipe,
	})
	register(command{
		Name: "db:show", Category: "db", Description: "Display information about the database connection",
		Run: handleDBShow,
	})

	register(command{
		Name: "migrate", Category: "migrate", Description: "Run the database migrations",
		Run: func(args []string) { handleDBMigrate() },
	})
	register(command{
		Name: "migrate:status", Aliases: []string{"db:status"}, Category: "migrate",
		Description: "Show the status of each migration", Run: func(args []string) { handleDBStatus() },
	})
	register(command{
		Name: "migrate:rollback", Aliases: []string{"db:rollback"}, Category: "migrate",
		Description: "Rollback the last database migration", Run: func(args []string) { handleDBRollback() },
	})
	register(command{
		Name: "migrate:reset", Category: "migrate", Description: "Rollback all database migrations",
		Run: handleMigrateReset,
	})
	register(command{
		Name: "migrate:refresh", Category: "migrate", Description: "Reset and re-run all migrations",
		Run: handleMigrateRefresh,
	})
	register(command{
		Name: "migrate:fresh", Category: "migrate", Description: "Drop all tables and re-run migrations",
		Usage: "gofreight migrate:fresh [--seed]", Run: handleMigrateFresh,
	})
	register(command{
		Name: "db:migrate", Category: "migrate", Description: "Run pending migrations (alias for migrate)",
		Run: func(args []string) { handleDBMigrate() }, Hidden: true,
	})

	registerMakeCommands()
	registerQueueCommands()
	registerCacheCommands()
	registerConfigCommands()
	registerRouteCommands()
	registerScheduleCommands()
	registerRouteCacheCommand()
	registerAuthCommands()
	registerOptimizeCommands()
	registerViewCommands()
	registerKeyCommands()

	// Legacy top-level aliases
	register(command{
		Name: "generate", Aliases: []string{"g"}, Category: "make", Description: "Generate application code",
		Usage: "gofreight generate <type> <name> [fields]", Run: handleGenerateCLI, Hidden: true,
	})
	register(command{
		Name: "make", Category: "make", Description: "Generate application code (alias for make:*)",
		Usage: "gofreight make <type> <name> [fields]", Run: handleGenerateCLI, Hidden: true,
	})
	register(command{
		Name: "routes", Category: "route", Description: "List registered routes (alias for route:list)",
		Run: func(args []string) { handleRouteList(args) }, Hidden: true,
	})
}

func registerMakeCommands() {
	makers := []struct {
		name, desc string
	}{
		{"model", "Create a new model class"},
		{"controller", "Create a new controller class"},
		{"migration", "Create a new migration file"},
		{"scaffold", "Create a full CRUD resource"},
		{"api", "Create a JSON API controller + resource"},
		{"auth", "Install authentication scaffolding"},
		{"service", "Create a new service class"},
		{"mail", "Create a new email class"},
		{"job", "Create a new job class"},
		{"middleware", "Create a new HTTP middleware class"},
		{"policy", "Create a new policy class"},
		{"request", "Create a new form request class"},
		{"seeder", "Create a new seeder class"},
		{"factory", "Create a new model factory"},
		{"test", "Create a new test class"},
	}
	for _, m := range makers {
		name := m.name
		desc := m.desc
		register(command{
			Name: "make:" + name, Category: "make", Description: desc,
			Usage: fmt.Sprintf("gofreight make:%s <name> [field:type ...]", name),
			Run: func(args []string) { handleMake(name, args) },
		})
	}
	register(command{
		Name: "make:resource", Category: "make", Description: "Create a full CRUD resource (alias for make:scaffold)",
		Usage: "gofreight make:resource <name> [field:type ...]",
		Run: func(args []string) { handleMake("scaffold", args) },
	})
	register(command{
		Name: "make:mailable", Category: "make", Description: "Create a new email class (alias for make:mail)",
		Usage: "gofreight make:mailable <name>",
		Run: func(args []string) { handleMake("mail", args) },
	})
}

func dispatch(name string, args []string) bool {
	for _, c := range allCommands {
		if c.Name == name {
			runCommand(c.Name, args, c.Run)
			return true
		}
		for _, a := range c.Aliases {
			if a == name {
				runCommand(c.Name, args, c.Run)
				return true
			}
		}
	}
	// make:foo from "make foo" legacy
	if name == "make" && len(args) >= 2 {
		makeName := "make:" + args[0]
		for _, c := range allCommands {
			if c.Name == makeName {
				runCommand(makeName, args[1:], c.Run)
				return true
			}
		}
	}
	return false
}

func runCommand(name string, args []string, run func([]string)) {
	if needsProductionGuard(name) && !guardProduction(name, args) {
		fmt.Println("Aborted.")
		os.Exit(0)
	}
	run(args)
}

func printCommandList(filter string) {
	printCommandListTo(os.Stdout, filter, isTerminal(os.Stdout))
}

func printCommandListTo(w io.Writer, filter string, useColor bool) {
	byCategory := map[string][]command{}
	var categories []string
	for _, c := range allCommands {
		if c.Hidden {
			continue
		}
		if filter != "" && !strings.Contains(c.Name, filter) && c.Category != filter {
			continue
		}
		if _, ok := byCategory[c.Category]; !ok {
			categories = append(categories, c.Category)
		}
		byCategory[c.Category] = append(byCategory[c.Category], c)
	}
	sort.Strings(categories)

	style := newConsoleStyle(useColor)

	fmt.Fprintf(w, "%s\n\n", style.bold("  Available commands"))
	for _, cat := range categories {
		fmt.Fprintf(w, "  %s\n", style.cyan(cat))
		cmds := byCategory[cat]
		sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })
		for _, c := range cmds {
			fmt.Fprintf(w, "    %-26s %s\n", style.bold(c.Name), style.dim(c.Description))
		}
		fmt.Fprintln(w)
	}
}

func handleHelp(args []string) {
	if len(args) == 0 {
		printCommandList("")
		return
	}
	name := args[0]
	for _, c := range allCommands {
		if c.Name == name || contains(c.Aliases, name) {
			fmt.Printf("Description:\n  %s\n", c.Description)
			if c.Usage != "" {
				fmt.Printf("\nUsage:\n  %s\n", c.Usage)
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", name)
	os.Exit(1)
}

func handleGenerateCLI(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: gofreight generate <type> <name> [field:type ...]")
		os.Exit(1)
	}
	handleMake(args[0], args[1:])
}

func handleMake(genType string, args []string) {
	if genType == "auth" {
		handleGenerate("auth", "", nil)
		return
	}
	if len(args) < 1 {
		fmt.Printf("Usage: gofreight make:%s <name> [field:type ...]\n", genType)
		os.Exit(1)
	}
	handleGenerate(genType, args[0], parseFields(args[1:]))
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func printUsage() {
	printCLIBanner(os.Stdout)
	printCommandList("")
	useColor := isTerminal(os.Stdout)
	style := newConsoleStyle(useColor)
	fmt.Fprintf(os.Stdout, "%s\n", style.dim("  Run 'gofreight help <command>' for more information on a command."))
}
