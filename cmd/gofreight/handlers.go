package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lsgser/gofreight/config"
	"github.com/lsgser/gofreight/database"
	"github.com/lsgser/gofreight/generator"
	"github.com/lsgser/gofreight/jobs"
	"github.com/lsgser/gofreight/version"
)

func handleNew(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gofreight new <app_name>")
		os.Exit(1)
	}
	if err := generator.NewApp(args[0]); err != nil {
		fail(err)
	}
	printNewAppWelcome(args[0])
}

func handleVersion() {
	printCLIBanner(os.Stdout)
	fmt.Printf("  Installed and ready.\n\n")
}

func handleAbout(args []string) {
	loadEnv()
	cfg := config.Load()
	wd, _ := os.Getwd()
	fmt.Println("Gofreight", version.Version)
	fmt.Println("Environment", cfg.Environment)
	fmt.Println("Path", wd)
	fmt.Println("Database", maskURL(cfg.DatabaseURL))
}

func handleEnv(args []string) {
	loadEnv()
	cfg := config.Load()
	fmt.Println(string(cfg.Environment))
}

func handleInspire(args []string) {
	quotes := []string{
		"Simplicity is the ultimate sophistication. — Leonardo da Vinci",
		"Make it work, make it right, make it fast. — Kent Beck",
		"The best way to predict the future is to invent it. — Alan Kay",
		"Programs must be written for people to read. — Harold Abelson",
	}
	n := time.Now().UnixNano() % int64(len(quotes))
	fmt.Println("\n \"" + quotes[n] + "\"\n")
}

func handleServe(args []string) {
	loadEnv()
	cfg := config.Load()
	printServeWelcome(cfg.Port, string(cfg.Environment))
	cmd := exec.Command("go", "run", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

func handleDown(args []string) {
	path := filepath.Join("storage", "framework", "maintenance")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		fail(err)
	}
	msg := "Application is down for maintenance."
	if len(args) > 0 {
		msg = strings.Join(args, " ")
	}
	if err := os.WriteFile(path, []byte(msg), 0644); err != nil {
		fail(err)
	}
	fmt.Println("Application is now in maintenance mode.")
}

func handleUp(args []string) {
	path := filepath.Join("storage", "framework", "maintenance")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		fail(err)
	}
	fmt.Println("Application is now live.")
}

func handleDBMigrate() {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fail(err)
	}
	migrator := database.NewMigrator("db/migrate")
	fmt.Println("Running migrations...")
	if err := migrator.Up(); err != nil {
		fail(err)
	}
	fmt.Println("Done.")
}

func handleDBRollback() {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fail(err)
	}
	migrator := database.NewMigrator("db/migrate")
	fmt.Println("Rolling back last migration...")
	if err := migrator.Down(); err != nil {
		fail(err)
	}
	fmt.Println("Done.")
}

func handleDBStatus() {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fail(err)
	}
	migrator := database.NewMigrator("db/migrate")
	statuses, err := migrator.Status()
	if err != nil {
		fail(err)
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
		fail(err)
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
			fail(err)
		}
		f.Close()
		fmt.Printf("Created SQLite database: %s\n", path)
	default:
		fmt.Printf("Database driver: %s\n", driver)
		fmt.Println("Ensure the database exists on your server, then run: gofreight migrate")
	}
}

func handleDBSeed(args []string) {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fail(err)
	}

	class := flagValue(args, "--class")
	if class != "" {
		runGoSeeder(class)
		return
	}

	fmt.Println("Running SQL seeds...")
	if err := database.RunSeeds("db/seeds"); err != nil {
		if !strings.Contains(err.Error(), "no seeds directory") {
			fail(err)
		}
	}

	if _, err := os.Stat("cmd/seed/main.go"); err == nil {
		fmt.Println("Running Go seeders...")
		runGoSeeder("")
	}
	fmt.Println("Done.")
}

func runGoSeeder(class string) {
	args := []string{"run", "./cmd/seed"}
	if class != "" {
		args = append(args, "--class="+class)
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

func handleDBWipe(args []string) {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fail(err)
	}
	fmt.Println("Dropping all tables...")
	if err := database.Wipe(); err != nil {
		fail(err)
	}
	fmt.Println("Done.")
}

func handleDBShow(args []string) {
	loadEnv()
	cfg := config.Load()
	driver, err := database.DetectDriver(cfg.DatabaseURL)
	if err != nil {
		fail(err)
	}
	fmt.Println("Connection", cfg.Database.Connection)
	if cfg.Database.Host != "" {
		fmt.Println("Host", cfg.Database.Host)
	}
	if cfg.Database.Port != "" {
		fmt.Println("Port", cfg.Database.Port)
	}
	fmt.Println("Database", cfg.Database.Database)
	fmt.Println("Driver", driver)
	fmt.Println("URL", maskURL(cfg.DatabaseURL))
}

func handleMigrateReset(args []string) {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fail(err)
	}
	migrator := database.NewMigrator("db/migrate")
	fmt.Println("Rolling back all migrations...")
	if err := migrator.Reset(); err != nil {
		fail(err)
	}
	fmt.Println("Done.")
}

func handleMigrateRefresh(args []string) {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fail(err)
	}
	migrator := database.NewMigrator("db/migrate")
	fmt.Println("Refreshing migrations...")
	if err := migrator.Refresh(); err != nil {
		fail(err)
	}
	fmt.Println("Done.")
}

func handleMigrateFresh(args []string) {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fail(err)
	}
	migrator := database.NewMigrator("db/migrate")
	fmt.Println("Dropping all tables and migrating...")
	if err := migrator.Fresh(); err != nil {
		fail(err)
	}
	if flagValue(args, "--seed") != "" || containsArg(args, "--seed") {
		handleDBSeed(nil)
		return
	}
	fmt.Println("Done.")
}

func handleGenerate(genType, name string, fields map[string]string) {
	appPath := "."
	switch genType {
	case "model":
		if err := generator.Model(appPath, name, fields); err != nil {
			fail(err)
		}
		fmt.Printf("Generated model: %s\n", name)
	case "controller":
		if err := generator.Controller(appPath, name); err != nil {
			fail(err)
		}
		fmt.Printf("Generated controller: %s\n", name)
	case "migration":
		dir := filepath.Join(appPath, "db", "migrate")
		path, err := database.CreateMigration(dir, name)
		if err != nil {
			fail(err)
		}
		fmt.Printf("Created migration: %s\n", path)
	case "scaffold", "resource":
		if err := generator.Resource(appPath, name, fields); err != nil {
			fail(err)
		}
		fmt.Printf("Generated scaffold: %s\n", name)
	case "api":
		if err := generator.API(appPath, name, fields); err != nil {
			fail(err)
		}
		fmt.Printf("Generated API: %s\n", name)
	case "auth":
		if err := generator.Auth(appPath); err != nil {
			fail(err)
		}
		fmt.Println("Generated auth (User model, seeds, route snippets)")
	case "service":
		if err := generator.Service(appPath, name); err != nil {
			fail(err)
		}
		fmt.Printf("Generated service: %s\n", name)
	case "mail", "mailable":
		if err := generator.Mail(appPath, name); err != nil {
			fail(err)
		}
		fmt.Printf("Generated mail: %s\n", name)
	case "job":
		if err := generator.Job(appPath, name); err != nil {
			fail(err)
		}
		fmt.Printf("Generated job: %s\n", name)
	case "middleware":
		if err := generator.Middleware(appPath, name); err != nil {
			fail(err)
		}
		fmt.Printf("Generated middleware: %s\n", name)
	case "policy":
		if err := generator.Policy(appPath, name); err != nil {
			fail(err)
		}
		fmt.Printf("Generated policy: %s\n", name)
	case "request":
		if err := generator.Request(appPath, name); err != nil {
			fail(err)
		}
		fmt.Printf("Generated request: %s\n", name)
	case "seeder":
		if err := generator.Seeder(appPath, name); err != nil {
			fail(err)
		}
		fmt.Printf("Generated seeder: %s\n", name)
	case "factory":
		if err := generator.Factory(appPath, name); err != nil {
			fail(err)
		}
		fmt.Printf("Generated factory: %s\n", name)
	case "test":
		if err := generator.Test(appPath, name); err != nil {
			fail(err)
		}
		fmt.Printf("Generated test: %s\n", name)
	default:
		fmt.Printf("Unknown generator: %s\n", genType)
		fmt.Println("Run 'gofreight list make' or see docs/generators.md")
		os.Exit(1)
	}
}

func registerQueueCommands() {
	register(command{
		Name: "queue:work", Category: "queue", Description: "Start processing jobs on the queue",
		Usage: "gofreight queue:work [--queue=key]", Run: handleQueueWork,
	})
	register(command{
		Name: "queue:failed", Category: "queue", Description: "List all failed queue jobs",
		Run: handleQueueFailed,
	})
	register(command{
		Name: "queue:retry", Category: "queue", Description: "Retry a failed queue job",
		Usage: "gofreight queue:retry <id>", Run: handleQueueRetry,
	})
	register(command{
		Name: "queue:flush", Category: "queue", Description: "Flush all failed queue jobs",
		Run: handleQueueFlush,
	})
	register(command{
		Name: "queue:clear", Category: "queue", Description: "Delete all jobs from the queue",
		Run: handleQueueClear,
	})
}

func redisQueue(args []string) (*jobs.RedisQueue, error) {
	loadEnv()
	url := config.ResolveRedisURL()
	if url == "" {
		return nil, fmt.Errorf("redis is not configured (set REDIS_HOST or REDIS_URL)")
	}
	key := flagValue(args, "--queue")
	if key == "" {
		key = "gofreight:jobs"
	}
	return jobs.NewRedisQueue(url, key)
}

func handleQueueWork(args []string) {
	q, err := redisQueue(args)
	if err != nil {
		fail(err)
	}
	fmt.Println("Processing jobs. Ctrl+C to stop.")
	ctx := context.Background()
	for {
		if err := q.ProcessOne(ctx); err != nil {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func handleQueueFailed(args []string) {
	q, err := redisQueue(args)
	if err != nil {
		fail(err)
	}
	failed, err := q.Failed(context.Background())
	if err != nil {
		fail(err)
	}
	if len(failed) == 0 {
		fmt.Println("No failed jobs.")
		return
	}
	for _, fj := range failed {
		fmt.Printf("%s  %s  attempts=%d  %s\n",
			fj.Payload.ID, fj.Payload.Name, fj.Payload.Attempts, fj.Exception)
	}
}

func handleQueueRetry(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gofreight queue:retry <id>")
		os.Exit(1)
	}
	q, err := redisQueue(args)
	if err != nil {
		fail(err)
	}
	if !q.RetryFailed(context.Background(), args[0]) {
		fmt.Println("Job not found.")
		os.Exit(1)
	}
	fmt.Println("Job re-queued.")
}

func handleQueueFlush(args []string) {
	q, err := redisQueue(args)
	if err != nil {
		fail(err)
	}
	if err := q.FlushFailed(context.Background()); err != nil {
		fail(err)
	}
	fmt.Println("Failed jobs flushed.")
}

func handleQueueClear(args []string) {
	q, err := redisQueue(args)
	if err != nil {
		fail(err)
	}
	if err := q.Clear(context.Background()); err != nil {
		fail(err)
	}
	fmt.Println("Queue cleared.")
}

func registerCacheCommands() {
	register(command{
		Name: "cache:clear", Category: "cache", Description: "Flush the application cache",
		Run: handleCacheClear,
	})
}

func handleCacheClear(args []string) {
	clearDir(filepath.Join("storage", "framework", "cache"))
	fmt.Println("Application cache cleared.")
}

func registerConfigCommands() {
	register(command{
		Name: "config:show", Category: "config", Description: "Display configuration values",
		Usage: "gofreight config:show [key]", Run: handleConfigShow,
	})
}

func handleConfigShow(args []string) {
	loadEnv()
	cfg := config.Load()
	if len(args) == 0 {
		fmt.Println("app_name:", cfg.AppName)
		fmt.Println("environment:", cfg.Environment)
		fmt.Println("app_url:", cfg.AppURL)
		fmt.Println("host:", cfg.Host)
		fmt.Println("port:", cfg.Port)
		fmt.Println("db_connection:", cfg.Database.Connection)
		fmt.Println("db_database:", cfg.Database.Database)
		if cfg.Database.Host != "" {
			fmt.Println("db_host:", cfg.Database.Host)
		}
		fmt.Println("database_url:", maskURL(cfg.DatabaseURL))
		fmt.Println("log_level:", cfg.LogLevel)
		return
	}
	key := args[0]
	switch key {
	case "environment", "env":
		fmt.Println(cfg.Environment)
	case "host":
		fmt.Println(cfg.Host)
	case "port":
		fmt.Println(cfg.Port)
	case "database_url", "database":
		fmt.Println(maskURL(cfg.DatabaseURL))
	case "db_connection":
		fmt.Println(cfg.Database.Connection)
	case "db_database":
		fmt.Println(cfg.Database.Database)
	case "log_level":
		fmt.Println(cfg.LogLevel)
	case "app_key", "secret_key":
		fmt.Println(maskSecret(cfg.AppKey))
	default:
		fmt.Fprintf(os.Stderr, "Unknown config key: %s\n", key)
		os.Exit(1)
	}
}

func registerRouteCommands() {
	register(command{
		Name: "route:list", Category: "route", Description: "List all registered routes",
		Run: handleRouteList,
	})
	register(command{
		Name: "route:clear", Category: "route", Description: "Remove the route cache file",
		Run: handleRouteClear,
	})
}

func handleRouteList(args []string) {
	routesDir := "routes"
	if _, err := os.Stat(routesDir); os.IsNotExist(err) {
		fmt.Println("No routes/ directory found. Register routes in routes/web.go and routes/api.go.")
		return
	}
	lines := scanRouteFiles(routesDir)
	if len(lines) == 0 {
		fmt.Println("No route definitions found in routes/*.go")
		return
	}
	fmt.Printf("%-7s %s\n", "METHOD", "PATTERN")
	fmt.Println(strings.Repeat("-", 50))
	for _, line := range lines {
		fmt.Println(line)
	}
}

func handleRouteClear(args []string) {
	path := filepath.Join("bootstrap", "cache", "routes.gob")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		fail(err)
	}
	fmt.Println("Route cache cleared.")
}

func registerAuthCommands() {
	register(command{
		Name: "auth:clear-resets", Category: "auth", Description: "Flush expired password reset tokens",
		Run: handleAuthClearResets,
	})
}

func handleAuthClearResets(args []string) {
	loadEnv()
	cfg := config.Load()
	if err := connectDB(cfg.DatabaseURL); err != nil {
		fail(err)
	}
	_, err := database.DB().Exec(`DELETE FROM password_reset_tokens WHERE expires_at < datetime('now')`)
	if err != nil {
		fmt.Println("No password_reset_tokens table or driver-specific cleanup required.")
		return
	}
	fmt.Println("Expired password reset tokens cleared.")
}

func registerOptimizeCommands() {
	register(command{
		Name: "optimize", Category: "optimize", Description: "Cache framework bootstrap files for performance",
		Run: handleOptimize,
	})
	register(command{
		Name: "optimize:clear", Category: "optimize", Description: "Remove cached bootstrap files",
		Run: handleOptimizeClear,
	})
}

func handleOptimize(args []string) {
	dir := filepath.Join("bootstrap", "cache")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, ".optimized")
	if err := os.WriteFile(path, []byte(time.Now().Format(time.RFC3339)), 0644); err != nil {
		fail(err)
	}
	fmt.Println("Files cached successfully.")
}

func handleOptimizeClear(args []string) {
	clearDir(filepath.Join("bootstrap", "cache"))
	fmt.Println("Caches cleared successfully.")
}

func registerViewCommands() {
	register(command{
		Name: "view:clear", Category: "view", Description: "Clear all compiled view files",
		Run: handleViewClear,
	})
}

func handleViewClear(args []string) {
	clearDir(filepath.Join("storage", "framework", "views"))
	fmt.Println("Compiled views cleared.")
}

func registerKeyCommands() {
	register(command{
		Name: "key:generate", Category: "key", Description: "Set the application encryption key",
		Usage: "gofreight key:generate [--show] [--force]",
		Run: handleKeyGenerate,
	})
}

func handleKeyGenerate(args []string) {
	show := false
	force := false
	for _, a := range args {
		switch a {
		case "--show":
			show = true
		case "--force":
			force = true
		}
	}

	key, err := config.GenerateAppKey()
	if err != nil {
		fail(err)
	}
	if show {
		fmt.Println("APP_KEY=" + key)
		return
	}

	if !force {
		if existing, ok := readEnvValue("APP_KEY"); ok && existing != "" && !isPlaceholderKey(existing) {
			fmt.Println("Application key already exists. Use --force to overwrite.")
			os.Exit(1)
		}
		if existing, ok := readEnvValue("SECRET_KEY"); ok && existing != "" && !isPlaceholderKey(existing) {
			fmt.Println("Application key already exists (SECRET_KEY). Use --force to overwrite.")
			os.Exit(1)
		}
	}

	if err := updateEnvKey("APP_KEY", key); err != nil {
		fmt.Println("APP_KEY=" + key)
		fmt.Println("Add APP_KEY to your .env file.")
		return
	}
	_ = removeEnvKey("SECRET_KEY")
	fmt.Println("Application key set successfully.")
}

func isPlaceholderKey(v string) bool {
	switch strings.TrimSpace(v) {
	case "", "change-me-in-production", "change-me", "your-secret-key-here", "dev-secret-key":
		return true
	default:
		return false
	}
}

func readEnvValue(name string) (string, bool) {
	data, err := os.ReadFile(".env")
	if err != nil {
		return "", false
	}
	prefix := name + "="
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimPrefix(line, prefix), true
		}
	}
	return "", false
}

func removeEnvKey(name string) error {
	data, err := os.ReadFile(".env")
	if err != nil {
		return err
	}
	prefix := name + "="
	var kept []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			continue
		}
		kept = append(kept, line)
	}
	return os.WriteFile(".env", []byte(strings.Join(kept, "\n")), 0644)
}

func updateEnvKey(name, value string) error {
	path := ".env"
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, name+"=") {
			lines[i] = name + "=" + value
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, name+"="+value)
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

func clearDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.Name() == ".gitkeep" {
			continue
		}
		os.RemoveAll(filepath.Join(dir, e.Name()))
	}
}

func scanRouteFiles(dir string) []string {
	var out []string
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)
		for _, method := range []string{"Get", "Post", "Put", "Patch", "Delete", "Resources"} {
			search := "r." + method + "("
			idx := 0
			for {
				pos := strings.Index(content[idx:], search)
				if pos < 0 {
					break
				}
				pos += idx
				rest := content[pos+len(search):]
				if method == "Resources" {
					end := strings.Index(rest, ",")
					if end > 0 {
						path := strings.Trim(strings.TrimSpace(rest[:end]), `"`)
						out = append(out, fmt.Sprintf("%-7s /%s", "REST", path))
					}
				} else {
					end := strings.Index(rest, ",")
					if end < 0 {
						end = strings.Index(rest, ")")
					}
					if end > 0 {
						path := strings.Trim(strings.TrimSpace(rest[:end]), `"`)
						out = append(out, fmt.Sprintf("%-7s %s", strings.ToUpper(method), path))
					}
				}
				idx = pos + len(search)
			}
		}
		return nil
	})
	sortStrings(out)
	return out
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func flagValue(args []string, name string) string {
	prefix := name + "="
	for _, a := range args {
		if strings.HasPrefix(a, prefix) {
			return strings.TrimPrefix(a, prefix)
		}
	}
	return ""
}

func containsArg(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}

func maskSecret(value string) string {
	if value == "" || config.IsDefaultAppKey(value) {
		return value
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "..." + value[len(value)-4:]
}

func maskURL(url string) string {
	if strings.Contains(url, "@") {
		parts := strings.SplitN(url, "@", 2)
		return "***@" + parts[1]
	}
	return url
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
