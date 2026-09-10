package main

/*
|--------------------------------------------------------------------------
| Extensive
|--------------------------------------------------------------------------
|
| Implements Extensive as part of the fileheaders package in the Gofreight
| framework.
| 
| Maintains Laravel-style file header blocks across the repository.
| 
| Run go run ./tools/fileheaders -force . from the repo root to regenerate
| extensive headers after structural changes.
| 
*/

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

// extensivePackageDocs describes each framework package in depth (multiple sentences).
var extensivePackageDocs = map[string][]string{
	"admin": {
		"The admin package exposes a development-only database browser and schema tools.",
		"It introspects tables and columns so you can inspect SQLite, PostgreSQL, or MySQL data without leaving the browser.",
		"Mount it only in non-production environments; see docs and application wiring for route registration.",
	},
	"api": {
		"The api package implements JSON resource transformers and helpers for versioned HTTP APIs.",
		"Resources map models and structs to consistent JSON shapes, pagination metadata, and optional link collections.",
		"Use api.Group with the router to register /api/v1-style routes with shared middleware.",
	},
	"application": {
		"The application package is the framework kernel: it constructs the Application value that owns the router, ORM, mailer, cache, queue, and view engine.",
		"bootstrap/app.go in your project returns application.New() with your bindings; Run() serves HTTP and optional background workers.",
		"Most cross-cutting services are configured here or via ConfigureIntegrations from environment variables.",
	},
	"assets": {
		"The assets package connects your public/ directory and optional Vite dev server to HTTP handlers.",
		"In development, Vite can proxy hot module replacement; in production, a manifest maps entry points to built files.",
		"GFT templates use #vite and /assets/ paths documented in the templating guide.",
	},
	"auth": {
		"The auth package covers session login, password hashing, API token storage, OAuth callbacks, email verification, and password reset flows.",
		"Controllers compose auth helpers with your User model; tokens and verification stores can be in-memory or database-backed.",
		"Install scaffolding with gofreight make:auth and wire find-user callbacks in app/auth.",
	},
	"build": {
		"The build package implements production compilation invoked by gofreight build.",
		"It sets GOOS/GOARCH, optional ldflags, output paths, and can emit a short deploy checklist after a successful compile.",
		"Use it from CI or locally when you need a single binary artifact rather than go run.",
	},
	"cache": {
		"The cache package defines CacheStore implementations: in-memory, file, Redis, and HTTP cache middleware.",
		"Application wiring selects the driver from CACHE_STORE and related env vars via integrations.",
		"Use cache for rate limiting data, session alternatives, or fragment caching in controllers.",
	},
	"channels": {
		"The channels package provides WebSocket broadcasting and a channel server for real-time features.",
		"Demo applications register chat or notification endpoints; production setups typically sit behind Redis or a dedicated broker.",
		"See channels documentation for server lifecycle and message fan-out patterns.",
	},
	"cmd/gofreight": {
		"This directory contains the gofreight CLI binary: command registration, terminal UI, and handlers for make:*, migrate, serve, test, and mail:preview.",
		"Each subcommand lives in its own source file; commands.go registers the catalog shown by gofreight list.",
		"Install locally with go install ./cmd/gofreight from the framework repository root.",
	},
	"config": {
		"The config package loads .env files, resolves MAIL_DRIVER and database URLs, and reads config/app.yaml.",
		"Database drivers and app keys are validated early so misconfiguration fails fast at boot.",
		"Application code reads config through helpers rather than os.Getenv scattered across the codebase.",
	},
	"container": {
		"The container package implements a lightweight service container for bindings and singletons.",
		"Register factories in bootstrap/app.go; resolve services from controllers or jobs by name or type.",
		"Pattern mirrors Laravel's container at a smaller scale for explicit wiring in Go.",
	},
	"controller": {
		"Controllers wrap http.HandlerFunc with a Base struct that exposes Request, Response, validation, views, and JSON helpers.",
		"This package defines RenderView, Redirect, status helpers, file downloads, and route registration utilities.",
		"Application controllers embed these patterns; see docs/controllers.md for request lifecycle.",
	},
	"database": {
		"The database package manages connections, fluent schema blueprints, Go and SQL migrations, seeding, and introspection.",
		"Migrations run via gofreight migrate; blueprints generate portable DDL across SQLite, PostgreSQL, and MySQL.",
		"Lower-level query helpers complement the model ORM in the model package.",
	},
	"dev": {
		"The dev package implements file watching for gofreight serve and gofreight dev.",
		"It restarts the process or reloads views when .go, .gft, .html, or config files change.",
		"Development-only; production deployments use a compiled binary without the watcher.",
	},
	"generator": {
		"The generator package powers gofreight new and all make:* scaffolds.",
		"It writes idiomatic directory layouts, GFT views, migrations, tests, and auth stubs from templates.",
		"CLI handlers in cmd/gofreight call into this package; templates live primarily in templates.go.",
	},
	"gftest": {
		"gftest is the feature testing harness used from tests/ in your application.",
		"NewApp boots a test HTTP server, runs migrations, exposes HTTP helpers (Get, Post, AssertOk), database assertions, and fakes for mail, cache, and queue.",
		"Run the suite with gofreight test; see docs/testing.md for factories and authentication in tests.",
	},
	"gftest/faker": {
		"Test fakers generate random emails, names, and type-aware values for factories and gftest seed data.",
		"Used by gofreight make:factory and table-driven tests that need realistic but non-production data.",
	},
	"graphql": {
		"The graphql package integrates graphql-go with Gofreight: schema registration, HTTP mount, playground, DataLoader batching, and field-level rate limits.",
		"Generate modules with gofreight make:graphql-module; SDL and resolvers live under app/graphql in your project.",
		"See docs/graphql.md and the tutorial for N+1 avoidance and security middleware.",
	},
	"health": {
		"Health check handlers report process liveness and optional dependency status for load balancers and orchestrators.",
		"Typically mounted at /health returning JSON; extend with database ping checks in your bootstrap.",
	},
	"i18n": {
		"The i18n package loads JSON locale files from config/locales and resolves translation keys in views and controllers.",
		"Middleware can set locale from session or Accept-Language; helpers mirror Laravel-style __() usage in GFT.",
	},
	"integrations": {
		"Integrations register pluggable drivers for mail, storage, cache, queue, and custom third-party APIs.",
		"Active() resolves the configured implementation from environment variables; wire Application in ConfigureIntegrations.",
		"Built-in connectors cover SMTP, SendGrid, S3-compatible storage, and Redis without vendor-specific SDKs in app code.",
	},
	"jobs": {
		"The jobs package defines queue interfaces, in-memory and Redis drivers, retries, and named job registration.",
		"Dispatch struct jobs or func workers from controllers; process with gofreight queue:work.",
		"Queued mail and long-running tasks should use this layer instead of blocking HTTP handlers.",
	},
	"mail": {
		"Mail covers Message and Mailer interfaces, LogMailer for tests, SMTP/SendGrid transports, mailable GFT rendering, and queued delivery.",
		"Mailables render app/views/mail templates through the view engine; preview with gofreight mail:preview.",
		"Configure MAIL_DRIVER in .env; authentication flows accept mail callbacks for reset and verification emails.",
	},
	"middleware": {
		"HTTP middleware implements sessions, CSRF, CORS, locale, structured logging, rate limiting, and maintenance mode.",
		"Register global middleware in bootstrap or attach to route groups for API-specific stacks.",
		"Session drivers include file, cookie, and Redis variants selected by SESSION_DRIVER.",
	},
	"model": {
		"The model package is the ORM layer: repositories, queries, associations, soft deletes, validation, serialization, collections, and pagination.",
		"Models map to tables via struct tags; migrations define schema separately in db/migrate.",
		"See docs/models.md, docs/orm.md, and docs/factories.md for Laravel-aligned patterns.",
	},
	"notification": {
		"Notifications send mail and database inbox messages from a single notification class with Via(), ToMail(), and ToDatabase().",
		"Wire a store callback for in-app alerts; queue mail channel delivery through the jobs system when needed.",
	},
	"plugins": {
		"Plugins register extension hooks so packages can subscribe to framework events without modifying core code.",
		"Use for optional packages or internal modules that need boot-time registration.",
	},
	"request": {
		"Form requests bind and validate HTTP input using vine schemas or validation tags before controller actions run.",
		"Integrates with controller Base for 422 Unprocessable responses and GFT error display helpers.",
	},
	"router": {
		"The router matches verbs and paths, supports groups, prefixes, named routes, constraints, signed URLs, and domain routing.",
		"Routes register in routes/web.go and routes/api.go; see docs/routing.md for middleware and model binding.",
	},
	"schedule": {
		"Schedule defines cron-like tasks invoked by gofreight schedule:run, suitable for system crontab or Kubernetes CronJob.",
		"Register closures or command strings in bootstrap/schedule.go.",
	},
	"storage": {
		"Storage abstracts local disk and cloud disks configured through FILESYSTEM_DISK and integrations.",
		"Uploads and public assets may use different disks; see docs/storage.md.",
	},
	"support/datetime": {
		"Datetime helpers parse and format timestamps consistently across models, APIs, and views.",
		"Complements carbon-style usage documented in configuration and ORM guides.",
	},
	"testrunner": {
		"testrunner powers gofreight test (feature tests in tests/) and gofreight test:unit (Go tests in app/).",
		"It sets GOFREIGHT_ENV=test and forwards flags to go test with sensible default package paths.",
	},
	"testutil": {
		"Shared test utilities for framework packages: temp databases, HTTP recorder helpers, and assertion shortcuts.",
		"Not imported by application code; for framework and generator tests.",
	},
	"tools/fileheaders": {
		"Maintains Laravel-style file header blocks across the repository.",
		"Run go run ./tools/fileheaders -force . from the repo root to regenerate extensive headers after structural changes.",
	},
	"upload": {
		"Upload helpers save multipart form files with sanitized names and size checks to configured storage directories.",
		"Use from controllers handling form posts with enctype multipart; paths typically under storage/uploads.",
	},
	"validation": {
		"Validation provides rule structs and runners used by models, form requests, and the vine DSL.",
		"Errors map to field names for JSON and GFT #error directives.",
	},
	"version": {
		"Central semver constants consumed by the CLI, generator scaffolds, and demo welcome pages.",
		"Bump Version here when releasing; tag Git with v prefix matching Module().",
	},
	"view": {
		"The view engine compiles Gofreight Templates (.gft) to html/template, supports layouts, slots, partials, and RenderString for mail.",
		"Views live under app/views; hot reload in development reloads templates on each request when enabled.",
	},
	"vine": {
		"Vine is a fluent validation DSL: vine.String().Required().Email() builds schemas validated against form maps or JSON.",
		"Used in form requests, API payloads, and anywhere you want Laravel-like validation chains in Go.",
	},
}

// extensiveFileNotes overrides the auto-generated intro for specific paths.
var extensiveFileNotes = map[string][]string{
	"cmd/gofreight/main.go": {
		"Entry point for the gofreight CLI binary.",
		"Parses os.Args, initializes the command registry, and dispatches to the matching handler or prints usage.",
		"Install with go install ./cmd/gofreight; subcommand implementations live alongside this file.",
	},
	"cmd/gofreight/commands.go": {
		"Registers every first-party CLI command with name, category, usage string, and Run func.",
		"initCommands is called from main before dispatch; add new commands here or via registerXxxCommands helpers.",
	},
	"version/version.go": {
		"Defines the current Gofreight semver and Module() tag string.",
		"Update Version when cutting a release; documentation and go install @v tags must stay in sync.",
	},
	"view/gft.go": {
		"Compiles Gofreight Template source into Go html/template syntax.",
		"Handles #layout, #slot, #each, #form, #partial, and output directives; compiled metadata drives layout merging in view.go.",
	},
	"view/view.go": {
		"Loads templates from app/views, executes layouts and sections, and writes HTML to http.ResponseWriter or strings.",
		"RenderString supports mailables and tests; SetReloadOnRender enables development hot reload.",
	},
	"application/application.go": {
		"Constructs Application with default LogMailer, memory queue, and view engine; Run() starts HTTP.",
		"Expose Jobs, Mailer, Router, and DB accessors used throughout controllers and CLI tools.",
	},
	"database/blueprint.go": {
		"Fluent schema builder for migrations: columns, indexes, timestamps, soft deletes, and dialect-specific SQL.",
		"Used by db/migrate/*.go files and gofreight make:migration output.",
	},
	"router/router.go": {
		"Core HTTP router: registers GET/POST/PUT/PATCH/DELETE routes, parameters, middleware chains, and dispatches requests.",
		"Groups and named routes build clean route files in application projects.",
	},
	"graphql/field.go": {
		"Helpers to define GraphQL fields with optional per-resolver rate limiting.",
		"Composes with graphql-go FieldConfig and security middleware documented in the GraphQL guide.",
	},
	"generator/templates.go": {
		"Large embedded template strings for gofreight new and generator output.",
		"Edit these constants when changing default app layout, README, migrations, or GFT stubs shipped to new projects.",
	},
}

func extensiveParagraphs(rel string, f *ast.File) []string {
	var out []string
	if notes, ok := extensiveFileNotes[rel]; ok {
		out = append(out, notes...)
	} else {
		out = append(out, inferFileIntro(rel, f)...)
	}
	dir := filepath.ToSlash(filepath.Dir(rel))
	if pkgParas := lookupPackageDocs(dir); len(pkgParas) > 0 {
		out = append(out, pkgParas...)
	}
	if exp := exportedSummary(f); exp != "" {
		out = append(out, exp)
	}
	if strings.HasSuffix(rel, "_test.go") {
		out = append(out, testRunHint(rel))
	}
	return dedupeParagraphs(out)
}

func lookupPackageDocs(dir string) []string {
	if p, ok := extensivePackageDocs[dir]; ok {
		return p
	}
	for dir != "." && dir != "" {
		if p, ok := extensivePackageDocs[dir]; ok {
			return p
		}
		dir = filepath.ToSlash(filepath.Dir(dir))
	}
	return nil
}

func inferFileIntro(rel string, f *ast.File) []string {
	base := strings.TrimSuffix(filepath.Base(rel), ".go")
	dir := filepath.ToSlash(filepath.Dir(rel))
	pkgName := filepath.Base(dir)

	if strings.HasSuffix(rel, "_test.go") {
		target := strings.TrimSuffix(base, "_test")
		return []string{
			fmt.Sprintf("Test suite for %s in the %s package.", humanIdent(target), pkgName),
			"Uses table-driven tests, httptest, or gftest where applicable. Failures should indicate regressions in public API or HTTP behavior.",
		}
	}
	if base == "doc" {
		return []string{
			fmt.Sprintf("Overview and conventions for the %s layer in a Gofreight application.", humanPackage(filepath.Dir(rel))),
			"Documents generator commands and where application code should live; not imported at runtime for logic.",
		}
	}
	if base == "main" {
		return []string{
			"Application main: loads bootstrap, registers routes, and starts the HTTP server.",
			"Use gofreight serve for development; production uses a binary from gofreight build.",
		}
	}

	role := humanIdent(base)
	exports := exportedNames(f)
	intro := fmt.Sprintf("Implements %s as part of the %s package in the Gofreight framework.", role, pkgName)
	if len(exports) > 0 {
		max := 6
		if len(exports) < max {
			max = len(exports)
		}
		intro += fmt.Sprintf(" Key symbols: %s.", strings.Join(exports[:max], ", "))
	}
	return []string{intro}
}

func exportedSummary(f *ast.File) string {
	var items []string
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.TYPE && d.Tok != token.CONST && d.Tok != token.VAR {
				continue
			}
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if !s.Name.IsExported() {
						continue
					}
					line := firstCommentLine(s.Doc)
					if line == "" {
						line = "exported type"
					}
					items = append(items, fmt.Sprintf("%s (%s)", s.Name.Name, line))
				case *ast.ValueSpec:
					for _, id := range s.Names {
						if !id.IsExported() {
							continue
						}
						line := firstCommentLine(s.Doc)
						if line == "" {
							line = "exported value"
						}
						items = append(items, fmt.Sprintf("%s (%s)", id.Name, line))
					}
				}
			}
		case *ast.FuncDecl:
			if d.Name == nil || !d.Name.IsExported() {
				continue
			}
			line := firstCommentLine(d.Doc)
			if line == "" {
				continue
			}
			items = append(items, fmt.Sprintf("%s (%s)", d.Name.Name, line))
		}
		if len(items) >= 8 {
			break
		}
	}
	if len(items) == 0 {
		return ""
	}
	return "Symbols defined here include: " + strings.Join(items, "; ") + "."
}

func exportedNames(f *ast.File) []string {
	var names []string
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.IsExported() {
					names = append(names, ts.Name.Name)
				}
			}
		case *ast.FuncDecl:
			if d.Name != nil && d.Name.IsExported() {
				names = append(names, d.Name.Name)
			}
		}
	}
	return names
}

func firstCommentLine(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	for _, c := range doc.List {
		text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		if text != "" {
			return text
		}
	}
	return ""
}

func testRunHint(rel string) string {
	dir := filepath.ToSlash(filepath.Dir(rel))
	if strings.HasPrefix(dir, "demoapp") || strings.HasPrefix(dir, "examples/") {
		return "Run from the app directory with go test ./... or gofreight test for tests/ packages."
	}
	if dir == "tests" || strings.Contains(dir, "/tests") {
		return "Run with gofreight test from the application root (sets GOFREIGHT_ENV=test)."
	}
	return "Run with go test ./" + dir + "/... or go test for this package from the framework root."
}

func extensiveTextMeta(rel string) (title string, paragraphs []string) {
	base := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
	title = humanIdent(base)
	ext := filepath.Ext(rel)

	switch {
	case strings.HasPrefix(rel, "docs/"):
		paragraphs = []string{
			"Official Gofreight documentation consumed by the docs site and linked from the CLI welcome page.",
			"Keep examples aligned with the current CLI and version; sync copies to gofreight-web when publishing.",
		}
		if strings.HasSuffix(rel, ".md") {
			paragraphs = append(paragraphs, "Covers concepts, tutorials, and reference material for this topic.")
		}
	case strings.Contains(rel, "views/"):
		paragraphs = []string{
			"Gofreight Template (GFT) view rendered by the view engine with optional layout and slots.",
			"Edit GFT syntax here; use gofreight serve in development for hot reload on .gft changes.",
		}
	case strings.Contains(rel, "db/migrate") || ext == ".sql":
		paragraphs = []string{
			"Database migration defining schema changes applied in order by gofreight migrate.",
			"Pair up/down migrations when altering columns; prefer blueprint Go migrations for new projects.",
		}
	case ext == ".css":
		paragraphs = []string{
			"Stylesheet served from public/ at /assets/ and linked from GFT layouts.",
			"Uses CSS variables for theming; adjust colors and layout for your brand.",
		}
	case ext == ".yaml" || ext == ".yml":
		paragraphs = []string{
			"YAML configuration loaded at application boot alongside .env.",
			"Non-secret defaults belong here; secrets stay in environment variables.",
		}
	default:
		paragraphs = []string{"Part of the Gofreight project source tree."}
	}
	return title, paragraphs
}

func dedupeParagraphs(in []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, p := range in {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}
