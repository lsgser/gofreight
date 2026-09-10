package generator

/*
|--------------------------------------------------------------------------
| Templates
|--------------------------------------------------------------------------
|
| Large embedded template strings for gofreight new and generator output.
| 
| Edit these constants when changing default app layout, README,
| migrations, or GFT stubs shipped to new projects.
| 
| The generator package powers gofreight new and all make:* scaffolds.
| 
| It writes idiomatic directory layouts, GFT views, migrations, tests, and
| auth stubs from templates.
| 
| CLI handlers in cmd/gofreight call into this package; templates live
| primarily in templates.go.
| 
*/

const appMainTmpl = `package main

/*
|--------------------------------------------------------------------------
| Application Entry Point
|--------------------------------------------------------------------------
|
| This file bootstraps the Gofreight application and starts the HTTP server.
| Application wiring lives in bootstrap/app.go. Routes are registered from
| the routes/ directory via routes.Register.
|
| Run locally:
|   gofreight serve
|   gofreight dev
|
*/

import (
	"log"
	"os"

	"{{.Module}}/bootstrap"
	"{{.Module}}/routes"
)

func main() {
	app := bootstrap.Application()

	if err := app.ConnectDatabase(); err != nil {
		log.Printf("warning: database not connected: %v", err)
	}

	if os.Getenv("GOFREIGHT_SCHEDULE_RUN") == "1" {
		for _, err := range bootstrap.RunSchedule() {
			log.Printf("schedule: %v", err)
		}
		return
	}

	app.Draw(routes.Register)
	app.Run()
}
`

const bootstrapAppTmpl = `package bootstrap

/*
|--------------------------------------------------------------------------
| Application Bootstrap
|--------------------------------------------------------------------------
|
| Configure the Gofreight application here: middleware, sessions, queues,
| locale, and service container bindings. This file is loaded on every
| request via main.go.
|
| Register services:
|   app.Singleton("payment", func() any { return services.NewPaymentService() })
|
| Resolve in controllers:
|   svc := app.Make("payment").(*services.PaymentService)
|
*/

import (
	"{{.Module}}/app/services"
	"github.com/lsgser/gofreight/application"
	"github.com/lsgser/gofreight/config"
)

/*
|--------------------------------------------------------------------------
| Application
|--------------------------------------------------------------------------
|
| Creates and configures the Gofreight application instance.
|
*/
func Application() *application.Application {
	app := application.New()

	if config.ResolveSessionDriver() == "redis" {
		_ = app.UseRedisSessions(config.ResolveRedisURL())
	} else if config.ResolveSessionDriver() == "file" {
		_ = app.UseFileSessions("")
	}

	if config.ResolveQueueConnection() == "redis" {
		_ = app.UseRedisQueue(config.ResolveRedisURL())
	}

	if config.ResolveRedisURL() != "" {
		_ = app.UseRedisBroadcast(config.ResolveRedisURL())
	}

	_ = app.ConfigureStorage()
	_ = app.ConfigureIntegrations()
	app.UseExceptionHandler()
	app.UseVite()

	_ = app.LoadLocales("config/locales")
	app.UseLocale()
	app.UseCSRF()

	/*
	|--------------------------------------------------------------------------
	| Service Container
	|--------------------------------------------------------------------------
	|
	| Register application services here.
	|
	*/
	app.Singleton("example", func() any {
		return services.NewExampleService()
	})

	return app
}
`

const bootstrapScheduleTmpl = `package bootstrap

import (
	"context"

	"github.com/lsgser/gofreight/application"
)

var scheduler = application.NewScheduler()

/*
|--------------------------------------------------------------------------
| Scheduled Tasks
|--------------------------------------------------------------------------
|
| Register recurring tasks here. Run due tasks with:
|   gofreight schedule:run
|
| Example crontab (every minute):
|   * * * * * cd /path/to/app && gofreight schedule:run
|
*/
func init() {
	// scheduler.Every(time.Minute, "heartbeat", func(ctx context.Context) error {
	// 	log.Println("scheduler tick")
	// 	return nil
	// })
}

// RunSchedule executes tasks that are due now (used by schedule:run CLI).
func RunSchedule() []error {
	return scheduler.RunDue(context.Background())
}
`

const routesRegisterTmpl = `package routes

/*
|--------------------------------------------------------------------------
| Route Registration
|--------------------------------------------------------------------------
|
| This file wires together all route groups for the application. Web routes
| (HTML) and API routes (JSON) are defined in separate files and loaded
| from here.
|
*/

import "github.com/lsgser/gofreight/router"

/*
|--------------------------------------------------------------------------
| Register
|--------------------------------------------------------------------------
|
| Loads web and API route groups onto the router.
|
*/
func Register(r *router.Router) {
	Web(r)

	r.Group(func(api *router.Router) {
		API(api)
	}).Prefix("/api/v1").Name("api.").Apply()
}
`

const routesWebTmpl = `package routes

/*
|--------------------------------------------------------------------------
| Web Routes
|--------------------------------------------------------------------------
|
| Register routes for browser requests here. These routes receive session
| state, CSRF protection, and typically return GFT HTML views.
|
| Generate a full CRUD resource:
|   gofreight make:scaffold Post title:string body:text
|
| Then register the generated routes below, e.g.:
|   controllers.RegisterPostRoutes(r)
|
*/

import (
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
)

/*
|--------------------------------------------------------------------------
| Web
|--------------------------------------------------------------------------
|
| Registers browser-facing HTTP routes.
|
*/
func Web(r *router.Router) {
	r.Get("/", controller.Handler(func(base controller.Base) error {
		return base.RenderView("home/index", base.ViewData(map[string]any{
			"Name":             "{{.Name}}",
			"DocsURL":            "{{.DocsURL}}",
			"FrameworkVersion":   "{{.FrameworkVersion}}",
		}))
	}), "home")

	/*
	|--------------------------------------------------------------------------
	| Generated Resources
	|--------------------------------------------------------------------------
	|
	| Register scaffold routes below, e.g.:
	|   controllers.RegisterPostRoutes(r)
	|
	*/

	/* gofreight:generated-resources */
}
`

const routesAPITmpl = `package routes

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
|
| Stateless JSON API routes. Prefix /api/v1 is applied in routes/register.go
| using route groups in routes/register.go.
|
|   api.ApiResource("posts", router.ApiResourceHandlers{ ... })
|
| Nested version groups:
|
|   r.Group(func(api *router.Router) {
|       api.Group(func(v1 *router.Router) { ... }).Prefix("/v1").Apply()
|   }).Prefix("/api").Apply()
|
*/

import (
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
)

/*
|--------------------------------------------------------------------------
| API
|--------------------------------------------------------------------------
|
| Registers JSON API routes (prefix applied by Register).
|
*/
func API(r *router.Router) {
	r.Get("/health", controller.Handler(func(base controller.Base) error {
		base.RenderJSON(map[string]string{"status": "ok"})
		return nil
	}), "health")
}
`

const appGoModTmpl = `module {{.Module}}

go 1.26.0

require github.com/lsgser/gofreight {{.FrameworkVersion}}
`

const appReadmeTmpl = `# {{.Name}}

A Gofreight web application (SQLite by default).

## Run

` + "```bash" + `
go mod tidy
gofreight key:generate
gofreight db:create
gofreight migrate
gofreight serve          # http://localhost:5000
` + "```" + `

## Production

` + "```bash" + `
gofreight build                    # bin/{{.Name}}
gofreight build --os linux --arch amd64
GOFREIGHT_ENV=production ./bin/{{.Name}}
` + "```" + `

See the [deployment guide](https://lsgser.github.io/gofreight-web/docs/deployment).

Admin (development): http://localhost:5000/admin

## Layout

` + "```" + `
bootstrap/app.go   Application wiring
routes/            Web and API routes
app/controllers/   HTTP handlers
app/models/        Database models
app/views/         GFT templates (.gft)
db/migrate/        SQL migrations
public/            Static assets
tests/             HTTP tests
` + "```" + `

Documentation: {{.DocsURL}}
`

const envExampleTmpl = `APP_NAME={{.Name}}
APP_ENV=local
GOFREIGHT_ENV=development
APP_DEBUG=true
APP_URL=http://localhost:5000

APP_LOCALE=en
APP_FALLBACK_LOCALE=en

HOST=0.0.0.0
PORT=5000

APP_KEY=

LOG_CHANNEL=stack
LOG_LEVEL=debug

DB_CONNECTION=sqlite
# DB_HOST=127.0.0.1
# DB_PORT=3306
# DB_DATABASE={{.Name}}
# DB_USERNAME=root
# DB_PASSWORD=

SESSION_DRIVER=file
SESSION_LIFETIME=120

BROADCAST_CONNECTION=log
FILESYSTEM_DISK=local
QUEUE_CONNECTION=sync
CACHE_STORE=file

REDIS_CLIENT=go-redis
REDIS_HOST=127.0.0.1
REDIS_PASSWORD=null
REDIS_PORT=6379

MAIL_MAILER=log
MAIL_HOST=127.0.0.1
MAIL_PORT=2525
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_FROM_ADDRESS=hello@example.com
MAIL_FROM_NAME="${APP_NAME}"

AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_DEFAULT_REGION=us-east-1
AWS_BUCKET=
AWS_USE_PATH_STYLE_ENDPOINT=false

ADMIN_PASSWORD=
`

const appYamlTmpl = `app_name: "{{.Name}}"
host: "0.0.0.0"
port: 5000
app_key: ""
log_level: "debug"

db_connection: sqlite
db_database: db/development.db
`

const localeEnTmpl = `{
  "welcome": "Welcome to {{.Name}}",
  "errors.required": "This field is required"
}
`

const exampleServiceTmpl = `package services

/*
|--------------------------------------------------------------------------
| Example Service
|--------------------------------------------------------------------------
|
| Services hold business logic and keep controllers thin. Register services
| in bootstrap/app.go and resolve them via the container:
|
|   app.Singleton("example", func() any { return services.NewExampleService() })
|
| Generate a new service:
|   gofreight make:service OrderProcessing
|
*/

/*
|--------------------------------------------------------------------------
| ExampleService
|--------------------------------------------------------------------------
|
| Demonstrates the service layer pattern.
|
*/
type ExampleService struct{}

/*
|--------------------------------------------------------------------------
| NewExampleService
|--------------------------------------------------------------------------
|
| Creates a new ExampleService instance.
|
*/
func NewExampleService() *ExampleService { return &ExampleService{} }
`

const controllersDocTmpl = `package controllers

/*
|--------------------------------------------------------------------------
| HTTP Controllers
|--------------------------------------------------------------------------
|
| Controllers handle incoming HTTP requests and return responses — HTML
| views, JSON, or redirects. Keep controllers thin; put business logic in
| app/services/.
|
| Generate a controller:
|   gofreight make:controller Post
|
| Generate a full CRUD resource (controller + model + views + migration):
|   gofreight make:scaffold Post title:string body:text
|
*/
`

const modelsDocTmpl = `package models

/*
|--------------------------------------------------------------------------
| Application Models
|--------------------------------------------------------------------------
|
| Models represent database tables and define validations, associations,
| and query helpers. Each model maps to a table via model.NewRepository.
|
| Generate a model:
|   gofreight make:model Post title:string body:text
|
*/
`

const resourcesDocTmpl = `package resources

/*
|--------------------------------------------------------------------------
| API Resources
|--------------------------------------------------------------------------
|
| API resources transform models into consistent JSON responses for your
| REST API. Use with API controllers in app/controllers/.
|
| Generate an API resource:
|   gofreight make:api Post title:string body:text
|
*/
`

const mailDocTmpl = `package mail

/*
|--------------------------------------------------------------------------
| Mailables
|--------------------------------------------------------------------------
|
| Mailable classes encapsulate email content and recipients. GFT templates live
| in app/views/mail/ with layout app/views/layouts/mail/default.gft.
| Configure delivery via MAIL_DRIVER in .env.
|
| Generate a mailable:
|   gofreight make:mail WelcomeMail
|
| Preview without sending:
|   gofreight mail:preview WelcomeMail
|
*/
`

const jobsDocTmpl = `package jobs

/*
|--------------------------------------------------------------------------
| Queueable Jobs
|--------------------------------------------------------------------------
|
| Jobs handle asynchronous work — sending email, processing uploads, calling
| external APIs. Dispatch from controllers; process with queue:work.
|
| Generate a job:
|   gofreight make:job SendNewsletter
|
| Run the worker (requires REDIS_URL when QUEUE_DRIVER=redis):
|   gofreight queue:work
|
*/
`

const middlewareDocTmpl = `package middleware

/*
|--------------------------------------------------------------------------
| HTTP Middleware
|--------------------------------------------------------------------------
|
| Application middleware runs on every request before your controller.
| Framework middleware (CSRF, sessions, CORS) is configured in bootstrap.
|
| Generate middleware:
|   gofreight make:middleware RequestLogger
|
*/
`

const policiesDocTmpl = `package policies

/*
|--------------------------------------------------------------------------
| Authorization Policies
|--------------------------------------------------------------------------
|
| Policies determine whether a user can perform an action on a resource.
| Call from controllers before create, update, or delete operations.
|
| Generate a policy:
|   gofreight make:policy PostPolicy
|
*/
`

const requestsDocTmpl = `package requests

/*
|--------------------------------------------------------------------------
| Form Requests
|--------------------------------------------------------------------------
|
| Form requests validate and authorize incoming HTTP input before it
| reaches your controller action.
|
| Generate a form request:
|   gofreight make:request StorePostRequest
|
*/
`

const seedersDocTmpl = `package seeders

/*
|--------------------------------------------------------------------------
| Database Seeders
|--------------------------------------------------------------------------
|
| Seeders populate your database with test or default data. Register
| seeders in cmd/seed/main.go and run:
|
|   gofreight db:seed
|   gofreight db:seed --class=DatabaseSeeder
|
| Generate a seeder:
|   gofreight make:seeder DatabaseSeeder
|
*/
`

const seedsReadmeTmpl = `-- --------------------------------------------------------------------------
-- Database Seed Files (SQL)
-- --------------------------------------------------------------------------
--
-- Place .sql files in this directory. They run in alphabetical order when
-- you execute:
--
--   gofreight db:seed
--
-- For Go-based seeders, use db/seeders/ and gofreight make:seeder.
--
`

const routesTmpl = `package config

import (
	"net/http"

	"{{.Module}}/app/controllers"
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
)

func Routes(r *router.Router) {
	r.Get("/", controller.Handler(func(base controller.Base) error {
		return base.RenderView("home/index", map[string]any{"Name": "{{.Name}}"})
	}))

	/*
	|--------------------------------------------------------------------------
	| Generated Resources
	|--------------------------------------------------------------------------
	|
	| Register scaffold routes below, e.g.:
	|   controllers.RegisterPostRoutes(r)
	|
	*/

	/* gofreight:generated-resources */
}
`

const databaseTmpl = `package config

/*
|--------------------------------------------------------------------------
| Database Configuration
|--------------------------------------------------------------------------
|
| Optional programmatic migrations for apps that prefer Go over SQL files.
| Most apps use db/migrate/*.sql and the gofreight migrate command instead.
|
*/

import "github.com/lsgser/gofreight/database"

/*
|--------------------------------------------------------------------------
| RunMigrations
|--------------------------------------------------------------------------
|
| Runs inline SQL migrations (optional — prefer db/migrate/).
|
*/
func RunMigrations() error {
	return database.Migrate(
		/* Add migration SQL here */
	)
}
`

const envTmpl = `APP_NAME={{.Name}}
GOFREIGHT_ENV=development
APP_URL=http://localhost:5000
PORT=5000
DB_CONNECTION=sqlite
# DB_HOST=127.0.0.1
# DB_PORT=3306
# DB_DATABASE={{.Name}}
# DB_USERNAME=root
# DB_PASSWORD=
APP_KEY=
`

const gitignoreTmpl = `.env
db/*.db
tmp/
*.exe
storage/logs/*
storage/framework/cache/*
storage/framework/sessions/*
!storage/**/.gitkeep
`

const modelTmpl = `package models

import (
	"context"

	"github.com/lsgser/gofreight/model"
)

type {{.Name}} struct {
	model.Record
{{range .Fields}}	{{.Name}} {{.GoType}} ` + "`db:\"{{.DBTag}}\" json:\"{{.JSONTag}}\"`" + `
{{end}}}

var {{.Name}}s = model.NewRepository[{{.Name}}]("{{.Table}}")

func (m *{{.Name}}) Save(ctx context.Context) error {
	if m.ID == 0 {
		return {{.Name}}s.Create(ctx, m)
	}
	return {{.Name}}s.Update(ctx, m)
}
`

const migrationSQLTmpl = `-- Migration: create_{{.Table}}
CREATE TABLE IF NOT EXISTS {{.Table}} (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
{{range .Fields}}	{{.DBTag}} {{.SQLType}},
{{end}}	created_at TEXT DEFAULT (datetime('now')),
	updated_at TEXT DEFAULT (datetime('now'))
);
`

const initialMigrationTmpl = `-- Migration: init
-- Bootstrap migration — confirms the migration runner is wired correctly.
-- Add tables with: gofreight make:migration create_posts_table
-- Or scaffold:     gofreight make:scaffold Post title:string body:text

SELECT 1;
`

const initialMigrationDownTmpl = `-- Rollback: init
-- No schema changes to revert.
SELECT 1;
`

const migrateReadmeTmpl = `# Database migrations

Blueprint migrations live in this directory as Go files (Laravel-style). Each file registers
` + "`database.RegisterMigration`" + ` in ` + "`init()`" + ` and uses ` + "`database.SchemaCreate`" + `,
` + "`database.SchemaTable`" + `, or ` + "`database.SchemaDrop`" + ` with the blueprint DSL.

Run migrations:

` + "```bash" + `
gofreight migrate
gofreight migrate:status
` + "```" + `

Create a new migration:

` + "```bash" + `
gofreight make:migration create_users_table
# edit db/migrate/YYYYMMDDHHMMSS_create_users_table.go
` + "```" + `

Example (like Laravel Schema::create):

` + "```go" + `
database.SchemaCreate(ctx, "users", func(b *database.Blueprint) {
    b.Id()
    b.String("email").NotNull().Unique()
    b.Timestamps()
})
` + "```" + `

Legacy SQL files (` + "*.sql`" + `) in this directory are still supported when ` + "`tools/migrate/`" + ` is not present.
`

const mailLayoutDefaultTmpl = `{#
|--------------------------------------------------------------------------
| Mail Layout
|--------------------------------------------------------------------------
|
| Shared HTML wrapper for mailable GFT templates (app/views/layouts/mail/).
|
#}
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>#place "subject" "Email"</title>
  <style>
    body { font-family: system-ui, sans-serif; line-height: 1.5; color: #1a1a1a; margin: 0; padding: 0; background: #f4f4f5; }
    .wrap { max-width: 560px; margin: 24px auto; background: #fff; border-radius: 8px; padding: 32px; box-shadow: 0 1px 3px rgba(0,0,0,.08); }
  </style>
</head>
<body>
  <div class="wrap">#place "content"</div>
</body>
</html>
`

const gftLayoutTmpl = `{#
|--------------------------------------------------------------------------
| Application Layout
|--------------------------------------------------------------------------
|
| Default HTML shell for browser pages: header, flash partial, and content slot.
|
#}
<!DOCTYPE html>
<html lang="en" data-theme="light">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="color-scheme" content="light dark">
  <title>#place "title" "{{.Name}}"</title>
  <link rel="icon" href="/assets/favicon.svg" type="image/svg+xml">
  <link rel="stylesheet" href="/assets/app.css">
  <script>
    (function () {
      var key = "gofreight-theme";
      var saved = localStorage.getItem(key);
      var theme = saved || (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
      document.documentElement.setAttribute("data-theme", theme);
    })();
  </script>
</head>
<body>
  <header class="site-header">
    <div class="container header-inner">
      <a class="brand" href="/">
        <span class="brand-mark" aria-hidden="true">◆</span>
        <span>{{.Name}}</span>
      </a>
      <nav class="header-nav" aria-label="Primary">
        <a href="{{.DocsURL}}docs/getting-started" target="_blank" rel="noreferrer">Docs</a>
        <a href="{{.DocsURL}}docs/tutorial-first-app" target="_blank" rel="noreferrer">Tutorials</a>
        <a href="/admin">Admin</a>
      </nav>
      <div class="header-actions">
        <button type="button" class="theme-toggle" id="theme-toggle" aria-label="Toggle dark mode">
          <span class="theme-icon theme-icon-light" aria-hidden="true">☀</span>
          <span class="theme-icon theme-icon-dark" aria-hidden="true">☾</span>
        </button>
        <span class="framework-badge">Gofreight</span>
      </div>
    </div>
  </header>
  <main class="container">
    #partial "partials/flash"
    #place "content"
  </main>
  <footer class="site-footer">
    <div class="container footer-inner">
      <p>Built with <strong>Gofreight</strong> — batteries-included Go web framework</p>
      <p class="footer-links">
        <a href="{{.DocsURL}}" target="_blank" rel="noreferrer">Documentation</a>
        <a href="{{.DocsURL}}docs/commands" target="_blank" rel="noreferrer">CLI reference</a>
      </p>
    </div>
  </footer>
  <script>
    (function () {
      var btn = document.getElementById("theme-toggle");
      if (!btn) return;
      btn.addEventListener("click", function () {
        var root = document.documentElement;
        var next = root.getAttribute("data-theme") === "dark" ? "light" : "dark";
        root.setAttribute("data-theme", next);
        localStorage.setItem("gofreight-theme", next);
      });
    })();
  </script>
</body>
</html>
`

const gftFlashPartialTmpl = `{#
|--------------------------------------------------------------------------
| Flash Partial
|--------------------------------------------------------------------------
|
| Session flash message markup. Include in layouts with #partial "partials.flash".
|
#}
#when .Flash
<div class="flash">{= .Flash }</div>
#endwhen
`

const gftHomeTmpl = `{#
|--------------------------------------------------------------------------
| Home Index View
|--------------------------------------------------------------------------
|
| Welcome page for new applications. Rendered by GET / in routes/web.go.
|
#}
#layout "layouts.application"

#slot "title"
{= .Name } — Welcome
#endslot

#slot "content"
<section class="hero">
  <div class="hero-badge-row">
    <span class="eyebrow">Gofreight {= .FrameworkVersion }</span>
    <span class="hero-pill">Single binary</span>
    <span class="hero-pill">SQLite ready</span>
  </div>
  <h1>Welcome to {= .Name }</h1>
  <p class="lead">Your application is running. Generate resources, wire routes, and ship production-ready Go — with sessions, migrations, views, and a full CLI built in.</p>
  <div class="hero-actions">
    <a class="btn btn-primary" href="{= .DocsURL }docs/getting-started" target="_blank" rel="noreferrer">Read the docs</a>
    <a class="btn btn-secondary" href="{= .DocsURL }docs/tutorial-first-app" target="_blank" rel="noreferrer">Start a tutorial</a>
    <a class="btn btn-ghost" href="/admin">Open admin</a>
  </div>
</section>

<section class="feature-grid">
  <article class="feature-card">
    <h3>Routing & APIs</h3>
    <p>Groups, named routes, model binding, signed URLs, and JSON resources.</p>
  </article>
  <article class="feature-card">
    <h3>GFT templates</h3>
    <p>Native <code>.gft</code> views with layouts, forms, CSRF, and validation helpers.</p>
  </article>
  <article class="feature-card">
    <h3>CLI generators</h3>
    <p>Scaffold models, controllers, migrations, tests, and auth with <code>gofreight make:*</code>.</p>
  </article>
</section>

<section class="steps">
  <h2>Next steps</h2>
  <div class="step-grid">
    <article class="step-card">
      <span class="step-num">1</span>
      <h3>Set your app key</h3>
      <p><code>gofreight key:generate</code></p>
    </article>
    <article class="step-card">
      <span class="step-num">2</span>
      <h3>Run migrations</h3>
      <p><code>gofreight migrate</code></p>
    </article>
    <article class="step-card">
      <span class="step-num">3</span>
      <h3>Scaffold a resource</h3>
      <p><code>gofreight make:scaffold Post title:string body:text</code></p>
    </article>
    <article class="step-card">
      <span class="step-num">4</span>
      <h3>Customize this page</h3>
      <p>Edit <code>app/views/home/index.gft</code> and <code>public/app.css</code></p>
    </article>
  </div>
</section>
#endslot
`

const layoutTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{.Name}}</title>
  <link rel="stylesheet" href="/assets/app.css">
</head>
<body>
  <header><h1>{{.Name}}</h1></header>
  <main>{{.Content}}</main>
</body>
</html>
`

const faviconTmpl = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" fill="none">
  <rect width="32" height="32" rx="8" fill="#0f172a"/>
  <path d="M7 10h18M7 16h13M7 22h16" stroke="#34d399" stroke-width="2.2" stroke-linecap="round"/>
  <path d="M22 14.5 26 16.5 22 18.5" stroke="#60a5fa" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
</svg>
`

const cssTmpl = `/*
 * Public assets — served from public/ at /assets/*
 */
:root,
[data-theme="light"] {
  color-scheme: light;
  --gf-cyan: #0891b2;
  --gf-cyan-soft: rgba(8, 145, 178, 0.12);
  --gf-accent: #f59e0b;
  --gf-text: #0f172a;
  --gf-muted: #64748b;
  --gf-border: #e2e8f0;
  --gf-bg: #f8fafc;
  --gf-surface: #ffffff;
  --gf-header: rgba(255, 255, 255, 0.88);
  --gf-hero-gradient: radial-gradient(circle at top left, rgba(8, 145, 178, 0.14), transparent 42%), linear-gradient(180deg, #ecfeff 0%, #f8fafc 320px, #ffffff 100%);
  --gf-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);
}

[data-theme="dark"] {
  color-scheme: dark;
  --gf-cyan: #22d3ee;
  --gf-cyan-soft: rgba(34, 211, 238, 0.14);
  --gf-accent: #fbbf24;
  --gf-text: #e2e8f0;
  --gf-muted: #94a3b8;
  --gf-border: #334155;
  --gf-bg: #0b1220;
  --gf-surface: #111827;
  --gf-header: rgba(15, 23, 42, 0.92);
  --gf-hero-gradient: radial-gradient(circle at top left, rgba(34, 211, 238, 0.12), transparent 40%), linear-gradient(180deg, #0f172a 0%, #0b1220 100%);
  --gf-shadow: 0 1px 2px rgba(0, 0, 0, 0.35);
}

* { box-sizing: border-box; }

body {
  font-family: Inter, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
  margin: 0;
  color: var(--gf-text);
  background: var(--gf-hero-gradient);
  line-height: 1.6;
  min-height: 100vh;
}

a { color: var(--gf-cyan); }
a:hover { opacity: 0.92; }

.container { max-width: 1080px; margin: 0 auto; padding: 0 1.5rem; }

.site-header {
  position: sticky;
  top: 0;
  z-index: 10;
  border-bottom: 1px solid var(--gf-border);
  background: var(--gf-header);
  backdrop-filter: blur(10px);
}

.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  min-height: 4rem;
  flex-wrap: wrap;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--gf-text);
  text-decoration: none;
}

.brand-mark {
  color: var(--gf-cyan);
  font-size: 0.95rem;
}

.header-nav {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.header-nav a {
  color: var(--gf-muted);
  text-decoration: none;
  font-weight: 500;
  font-size: 0.95rem;
}

.header-nav a:hover { color: var(--gf-text); }

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.theme-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 999px;
  border: 1px solid var(--gf-border);
  background: var(--gf-surface);
  color: var(--gf-text);
  cursor: pointer;
}

.theme-icon-dark { display: none; }
[data-theme="dark"] .theme-icon-light { display: none; }
[data-theme="dark"] .theme-icon-dark { display: inline; }

.framework-badge {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--gf-cyan);
  border: 1px solid var(--gf-cyan-soft);
  background: var(--gf-cyan-soft);
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
}

main { padding: 2.5rem 0 4rem; }

.hero { padding: 2rem 0 2rem; max-width: 48rem; }

.hero-badge-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.eyebrow {
  display: inline-block;
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--gf-cyan);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.hero-pill {
  display: inline-block;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--gf-muted);
  border: 1px solid var(--gf-border);
  background: var(--gf-surface);
  padding: 0.25rem 0.55rem;
  border-radius: 999px;
}

.hero h1 {
  font-size: clamp(2.2rem, 5vw, 3.4rem);
  line-height: 1.08;
  margin: 0 0 1rem;
  letter-spacing: -0.03em;
}

.lead {
  font-size: 1.125rem;
  color: var(--gf-muted);
  max-width: 42rem;
  margin: 0 0 1.75rem;
}

.hero-actions { display: flex; flex-wrap: wrap; gap: 0.75rem; }

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.72rem 1.15rem;
  border-radius: 0.65rem;
  font-weight: 600;
  text-decoration: none;
  border: 1px solid transparent;
  transition: transform 0.15s ease, filter 0.15s ease;
}

.btn:hover { transform: translateY(-1px); }

.btn-primary {
  background: var(--gf-cyan);
  color: #fff;
  box-shadow: 0 8px 24px rgba(8, 145, 178, 0.22);
}

.btn-secondary {
  background: var(--gf-surface);
  color: var(--gf-text);
  border-color: var(--gf-border);
}

.btn-ghost {
  background: transparent;
  color: var(--gf-muted);
  border-color: var(--gf-border);
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
  margin: 0 0 2.5rem;
}

.feature-card {
  background: var(--gf-surface);
  border: 1px solid var(--gf-border);
  border-radius: 0.85rem;
  padding: 1.25rem;
  box-shadow: var(--gf-shadow);
}

.feature-card h3 { margin: 0 0 0.5rem; font-size: 1rem; }
.feature-card p { margin: 0; color: var(--gf-muted); font-size: 0.95rem; }

.steps h2 { margin: 0 0 1.25rem; font-size: 1.35rem; }

.step-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.step-card {
  background: var(--gf-surface);
  border: 1px solid var(--gf-border);
  border-radius: 0.85rem;
  padding: 1.25rem;
  box-shadow: var(--gf-shadow);
}

.step-num {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 999px;
  background: var(--gf-cyan-soft);
  color: var(--gf-cyan);
  font-weight: 700;
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
}

.step-card h3 { margin: 0 0 0.5rem; font-size: 1rem; }
.step-card p { margin: 0; color: var(--gf-muted); font-size: 0.95rem; }
.step-card code { font-size: 0.82rem; word-break: break-word; }

code {
  background: var(--gf-bg);
  padding: 0.15rem 0.35rem;
  border-radius: 0.25rem;
  border: 1px solid var(--gf-border);
}

.site-footer {
  border-top: 1px solid var(--gf-border);
  padding: 1.5rem 0 2rem;
  color: var(--gf-muted);
}

.footer-inner p { margin: 0 0 0.5rem; }
.footer-links { display: flex; gap: 1rem; flex-wrap: wrap; }
.footer-links a { color: var(--gf-muted); text-decoration: none; }
.footer-links a:hover { color: var(--gf-cyan); }

.flash {
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.35);
  color: #047857;
  padding: 0.75rem 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
}

[data-theme="dark"] .flash {
  color: #6ee7b7;
}

@media (max-width: 720px) {
  .header-nav { width: 100%; order: 3; }
  .header-actions { margin-left: auto; }
}
`

const testExampleTmpl = `package tests

/*
|--------------------------------------------------------------------------
| Example Feature Test
|--------------------------------------------------------------------------
|
| Tests live in tests/ and use gftest for HTTP assertions and database
| setup. Run the suite with:
|
|   gofreight test
|
| Generate a test:
|   gofreight make:test Posts
|
*/

import (
	"testing"

	"{{.Module}}/routes"
	"github.com/lsgser/gofreight/gftest"
)

func TestApp(t *testing.T) {
	gftest.Describe(t, "{{.Name}}", func(d *gftest.DescribeContext) {
		var app *gftest.App

		d.BeforeEach(func(t *testing.T) {
			app = gftest.NewApp(t, gftest.WithDatabase("sqlite://:memory:"))
			app.Draw(routes.Register)
		})

		d.It("returns the homepage", func(t *testing.T) {
			app.Get("/").AssertOk()
		})
	})
}
`

const testFactoriesTmpl = `package factories

/*
|--------------------------------------------------------------------------
| Model Factories
|--------------------------------------------------------------------------
|
| Factories create test model instances with sensible defaults. Use in
| tests to avoid repeating setup boilerplate.
|
| Generate a factory (uses gftest/faker for random defaults):
|   gofreight make:factory Post
|
*/

/*
|--------------------------------------------------------------------------
| Factories
|--------------------------------------------------------------------------
|
| Add model factories here — see gofreight make:factory.
|
*/
`

const controllerTmpl = `package controllers

import (
	"context"
	"fmt"
	"net/http"

	"{{.Module}}/app/models"
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
)

type {{.Name}}Controller struct{}

func Register{{.Name}}Routes(r *router.Router) {
	c := {{.Name}}Controller{}
	r.Resources("{{.Plural}}", router.ResourceHandlers{
		Index:   controller.Handler(c.Index),
		Show:    controller.Handler(c.Show),
		Create:  controller.Handler(c.Create),
		Update:  controller.Handler(c.Update),
		Destroy: controller.Handler(c.Destroy),
	})
}

func (c {{.Name}}Controller) Index(base controller.Base) error {
	records, err := models.{{.Name}}s.All(context.Background())
	if err != nil {
		return err
	}
	base.RenderJSON(records)
	return nil
}

func (c {{.Name}}Controller) Show(base controller.Base) error {
	id := base.Param("id")
	record, err := models.{{.Name}}s.Find(context.Background(), parseID(id))
	if err != nil {
		base.NotFound("{{.Name}} not found")
		return nil
	}
	base.RenderJSON(record)
	return nil
}

func (c {{.Name}}Controller) Create(base controller.Base) error {
	/*
	|--------------------------------------------------------------------------
	| Create
	|--------------------------------------------------------------------------
	|
	| Parse request body and create record.
	|
	*/
	base.Status = http.StatusCreated
	base.RenderJSON(map[string]string{"message": "created"})
	return nil
}

func (c {{.Name}}Controller) Update(base controller.Base) error {
	base.RenderJSON(map[string]string{"message": "updated"})
	return nil
}

func (c {{.Name}}Controller) Destroy(base controller.Base) error {
	base.Status = http.StatusNoContent
	return nil
}

func parseID(s string) int64 {
	var id int64
	fmt.Sscanf(s, "%d", &id)
	return id
}
`
