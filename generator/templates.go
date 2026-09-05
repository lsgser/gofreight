package generator

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

	"{{.Module}}/bootstrap"
	"{{.Module}}/routes"
)

func main() {
	app := bootstrap.Application()

	if err := app.ConnectDatabase(); err != nil {
		log.Printf("warning: database not connected: %v", err)
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

// Application creates and configures the Gofreight application instance.
func Application() *application.Application {
	app := application.New()

	if config.ResolveSessionDriver() == "redis" {
		_ = app.UseRedisSessions(config.ResolveRedisURL())
	}

	if config.ResolveQueueConnection() == "redis" {
		_ = app.UseRedisQueue(config.ResolveRedisURL())
	}

	_ = app.LoadLocales("config/locales")
	app.UseLocale()
	app.UseCSRF()

	// Service container — register application services here.
	app.Singleton("example", func() any {
		return services.NewExampleService()
	})

	return app
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

// Register loads web and API route groups onto the router.
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

// Web registers browser-facing HTTP routes.
func Web(r *router.Router) {
	r.Get("/", controller.Handler(func(base controller.Base) error {
		return base.RenderView("home/index", map[string]any{"Name": "{{.Name}}"})
	}))

	// Generated resources register here, e.g.:
	// controllers.RegisterPostRoutes(r)
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

// API registers JSON API routes (prefix applied by Register).
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
gofreight db:create
gofreight migrate
gofreight serve          # http://localhost:5000
` + "```" + `

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

Full structure: https://github.com/lsgser/gofreight/blob/main/docs/project-structure.md
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

// ExampleService demonstrates the service layer pattern.
type ExampleService struct{}

// NewExampleService creates a new ExampleService instance.
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
| Mailable classes encapsulate email content and recipients. Templates live
| in app/views/mail/. Configure delivery via MAIL_DRIVER in .env.
|
| Generate a mailable:
|   gofreight make:mail WelcomeMail
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

	// Generated resources register here, e.g.:
	// controllers.RegisterPostRoutes(r)
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

// RunMigrations runs inline SQL migrations (optional — prefer db/migrate/).
func RunMigrations() error {
	return database.Migrate(
		// Add migration SQL here
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
{{range .Fields}}	{{.DBTag}} {{.SQLType}} NOT NULL,
{{end}}	created_at TEXT DEFAULT (datetime('now')),
	updated_at TEXT DEFAULT (datetime('now'))
);
`

const gftLayoutTmpl = `{# --------------------------------------------------------------------------
   Layout: Application Shell
   -------------------------------------------------------------------------- #}
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>#place "title" "{{.Name}}"</title>
  <link rel="stylesheet" href="/assets/app.css">
</head>
<body>
  <header class="site-header">
    <div class="container header-inner">
      <a class="brand" href="/">{{.Name}}</a>
      <span class="framework-badge">Gofreight</span>
    </div>
  </header>
  <main class="container">
    #place "content"
  </main>
  <footer class="site-footer">
    <div class="container">
      <small>Built with <strong>Gofreight</strong> — batteries-included Go web framework</small>
    </div>
  </footer>
</body>
</html>
`

const gftFlashPartialTmpl = `{# Partial: flash messages — include with #partial "partials.flash" #}
#when .Flash
<div class="flash">{= .Flash }</div>
#endwhen
`

const gftHomeTmpl = `{# View: home/index — rendered by the / route in routes/web.go #}
#layout "layouts.application"

#slot "title"
{= .Name } — Welcome
#endslot

#slot "content"
<section class="hero">
  <p class="eyebrow">Powered by Gofreight</p>
  <h1>Welcome to {= .Name }</h1>
  <p class="lead">Your Go web application is ready. Start building routes, models, and views — then ship it as a single binary.</p>
  <div class="hero-actions">
    <a class="btn btn-primary" href="https://github.com/lsgser/gofreight/tree/main/docs">Read the docs</a>
    <a class="btn btn-secondary" href="/admin">Open admin</a>
  </div>
</section>

<section class="steps">
  <h2>Get started</h2>
  <div class="step-grid">
    <article class="step-card">
      <span class="step-num">1</span>
      <h3>Generate a resource</h3>
      <p><code>gofreight make:scaffold Post title:string body:text</code></p>
    </article>
    <article class="step-card">
      <span class="step-num">2</span>
      <h3>Run migrations</h3>
      <p><code>gofreight migrate</code></p>
    </article>
    <article class="step-card">
      <span class="step-num">3</span>
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

const cssTmpl = `/*
 * Public assets — served from public/ at /assets/*
 */
:root {
  --gf-cyan: #00add8;
  --gf-navy: #0f172a;
  --gf-orange: #f59e0b;
  --gf-muted: #64748b;
  --gf-border: #e2e8f0;
  --gf-bg: #f8fafc;
}

* { box-sizing: border-box; }

body {
  font-family: system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
  margin: 0;
  color: var(--gf-navy);
  background: linear-gradient(180deg, #f0f9ff 0%, #fff 320px);
  line-height: 1.6;
}

.container { max-width: 960px; margin: 0 auto; padding: 0 1.5rem; }

.site-header {
  border-bottom: 1px solid var(--gf-border);
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(8px);
}

.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 4rem;
}

.brand {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--gf-navy);
  text-decoration: none;
}

.framework-badge {
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--gf-cyan);
  border: 1px solid rgba(0, 173, 216, 0.35);
  background: rgba(0, 173, 216, 0.08);
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
}

main { padding: 2.5rem 0 4rem; }

.hero {
  padding: 2rem 0 2.5rem;
}

.eyebrow {
  display: inline-block;
  margin: 0 0 1rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--gf-cyan);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.hero h1 {
  font-size: clamp(2rem, 5vw, 3rem);
  line-height: 1.15;
  margin: 0 0 1rem;
}

.lead {
  font-size: 1.125rem;
  color: var(--gf-muted);
  max-width: 42rem;
  margin: 0 0 1.75rem;
}

.hero-actions { display: flex; flex-wrap: wrap; gap: 0.75rem; }

.btn {
  display: inline-block;
  padding: 0.7rem 1.1rem;
  border-radius: 0.5rem;
  font-weight: 600;
  text-decoration: none;
  border: 1px solid transparent;
}

.btn-primary {
  background: var(--gf-cyan);
  color: #fff;
}

.btn-primary:hover { filter: brightness(0.95); }

.btn-secondary {
  background: #fff;
  color: var(--gf-navy);
  border-color: var(--gf-border);
}

.steps h2 { margin: 0 0 1.25rem; font-size: 1.35rem; }

.step-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.step-card {
  background: #fff;
  border: 1px solid var(--gf-border);
  border-radius: 0.75rem;
  padding: 1.25rem;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
}

.step-num {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 999px;
  background: rgba(0, 173, 216, 0.12);
  color: var(--gf-cyan);
  font-weight: 700;
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
}

.step-card h3 { margin: 0 0 0.5rem; font-size: 1rem; }
.step-card p { margin: 0; color: var(--gf-muted); font-size: 0.95rem; }
.step-card code {
  font-size: 0.82rem;
  word-break: break-word;
}

code {
  background: var(--gf-bg);
  padding: 0.15rem 0.35rem;
  border-radius: 0.25rem;
}

.site-footer {
  border-top: 1px solid var(--gf-border);
  padding: 1.25rem 0 2rem;
  color: var(--gf-muted);
}

.flash {
  background: #ecfdf5;
  border: 1px solid #a7f3d0;
  color: #065f46;
  padding: 0.75rem 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
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

import "github.com/lsgser/gofreight/gftest/faker"

// Add model factories here — see gofreight make:factory.
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
	// Parse request body and create record
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
