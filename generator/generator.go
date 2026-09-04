package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/lsgser/gofreight/version"
)

// NewApp scaffolds a new Gofreight application.
func NewApp(name string) error {
	root := filepath.Join(".", name)

	dirs := []string{
		filepath.Join(root, "app", "controllers"),
		filepath.Join(root, "app", "models"),
		filepath.Join(root, "app", "views", "layouts"),
		filepath.Join(root, "app", "views", "partials"),
		filepath.Join(root, "app", "views", "components"),
		filepath.Join(root, "app", "views", "home"),
		filepath.Join(root, "config"),
		filepath.Join(root, "db", "migrate"),
		filepath.Join(root, "db", "seeds"),
		filepath.Join(root, "public"),
		filepath.Join(root, "storage", "uploads"),
		filepath.Join(root, "tests", "factories"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		filepath.Join(root, "main.go"):               appMainTmpl,
		filepath.Join(root, "go.mod"):                appGoModTmpl,
		filepath.Join(root, "README.md"):             appReadmeTmpl,
		filepath.Join(root, "config", "routes.go"):   routesTmpl,
		filepath.Join(root, "config", "database.go"): databaseTmpl,
		filepath.Join(root, ".env"):                  envTmpl,
		filepath.Join(root, ".env.example"):          envExampleTmpl,
		filepath.Join(root, ".gitignore"):            gitignoreTmpl,
		filepath.Join(root, "app", "views", "layouts", "application.gft"): gftLayoutTmpl,
		filepath.Join(root, "app", "views", "partials", "flash.gft"):       gftFlashPartialTmpl,
		filepath.Join(root, "app", "views", "home", "index.gft"):            gftHomeTmpl,
		filepath.Join(root, "public", "app.css"):                           cssTmpl,
		filepath.Join(root, "tests", "example_test.go"):                    testExampleTmpl,
		filepath.Join(root, "tests", "factories", "factories.go"):          testFactoriesTmpl,
	}

	data := struct{ Name, Module, FrameworkVersion string }{
		Name:             name,
		Module:           strings.ToLower(name),
		FrameworkVersion: version.Module(),
	}

	for path, tmpl := range files {
		if err := writeTemplate(path, tmpl, data); err != nil {
			return err
		}
	}

	fmt.Printf("Created new Gofreight app: %s\n", name)
	fmt.Printf("  cd %s && go mod tidy && go run .\n", name)
	return nil
}

// Model generates a model file and migration.
func Model(appPath, name string, fields map[string]string) error {
	table := pluralize(strings.ToLower(name))
	modelDir := filepath.Join(appPath, "app", "models")
	migrateDir := filepath.Join(appPath, "db", "migrate")

	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(migrateDir, 0755); err != nil {
		return err
	}

	type Field struct {
		Name      string
		GoType    string
		SQLType   string
		DBTag     string
		JSONTag   string
	}
	var fieldList []Field
	for fname, ftype := range fields {
		fieldList = append(fieldList, Field{
			Name:    title(fname),
			GoType:  goType(ftype),
			SQLType: sqlType(ftype),
			DBTag:   strings.ToLower(fname),
			JSONTag: strings.ToLower(fname),
		})
	}

	data := struct {
		Name, Table string
		Fields      []Field
	}{
		Name:  title(name),
		Table: table,
		Fields: fieldList,
	}

	modelPath := filepath.Join(modelDir, strings.ToLower(name)+".go")
	if err := writeTemplate(modelPath, modelTmpl, data); err != nil {
		return err
	}

	migrationPath := filepath.Join(migrateDir, fmt.Sprintf("001_create_%s.sql", table))
	return writeTemplate(migrationPath, migrationSQLTmpl, data)
}

// Controller generates a controller with RESTful actions.
func Controller(appPath, name string) error {
	dir := filepath.Join(appPath, "app", "controllers")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	module := filepath.Base(appPath)
	if module == "." {
		module = "app"
	}

	data := struct {
		Name, Module, Plural, Table, ModelName string
	}{
		Name:      title(name),
		Module:    module,
		Plural:    pluralize(strings.ToLower(name)),
		Table:     pluralize(strings.ToLower(name)),
		ModelName: title(name),
	}

	path := filepath.Join(dir, strings.ToLower(name)+"_controller.go")
	return writeTemplate(path, controllerTmpl, data)
}

func writeTemplate(path, tmplStr string, data any) error {
	tmpl, err := template.New(filepath.Base(path)).Parse(tmplStr)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func pluralize(s string) string {
	if strings.HasSuffix(s, "y") {
		return strings.TrimSuffix(s, "y") + "ies"
	}
	if strings.HasSuffix(s, "s") {
		return s + "es"
	}
	return s + "s"
}

func goType(t string) string {
	switch t {
	case "string", "text":
		return "string"
	case "int", "integer":
		return "int"
	case "bool", "boolean":
		return "bool"
	case "float":
		return "float64"
	default:
		return "string"
	}
}

func sqlType(t string) string {
	switch t {
	case "string":
		return "VARCHAR(255)"
	case "text":
		return "TEXT"
	case "int", "integer":
		return "INTEGER"
	case "bool", "boolean":
		return "BOOLEAN DEFAULT FALSE"
	case "float":
		return "DOUBLE PRECISION"
	default:
		return "VARCHAR(255)"
	}
}

const appMainTmpl = `package main

import (
	"log"

	"{{.Module}}/config"
	"github.com/lsgser/gofreight/application"
)

func main() {
	app := application.New()

	if err := app.ConnectDatabase(); err != nil {
		log.Printf("warning: database not connected: %v", err)
	}

	app.Draw(config.Routes)
	app.Run()
}
`

const appGoModTmpl = `module {{.Module}}

go 1.22

require github.com/lsgser/gofreight {{.FrameworkVersion}}
`

const appReadmeTmpl = `# {{.Name}}

Gofreight application.

## Layout

` + "```" + `
app/controllers/   HTTP handlers
app/models/        Database models
app/views/         GFT templates (.gft)
config/routes.go   Routes
db/migrate/        SQL migrations
public/            Static assets
tests/             HTTP tests
` + "```" + `

Full structure: https://github.com/lsgser/gofreight/blob/main/docs/project-structure.md

## Run

` + "```bash" + `
go mod tidy
gofreight db:migrate
GOFREIGHT_ENV=development go run .
` + "```" + `
`

const envExampleTmpl = `GOFREIGHT_ENV=development
PORT=3000
DATABASE_URL=postgres://localhost/{{.Module}}_development?sslmode=disable
SECRET_KEY=change-me-in-production
ADMIN_PASSWORD=
REDIS_URL=redis://localhost:6379
MAIL_DRIVER=log
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

import "github.com/lsgser/gofreight/database"

func RunMigrations() error {
	return database.Migrate(
		// Add migration SQL here
	)
}
`

const envTmpl = `GOFREIGHT_ENV=development
PORT=3000
DATABASE_URL=postgres://localhost/{{.Module}}_development?sslmode=disable
SECRET_KEY=change-me-in-production
ADMIN_PASSWORD=
REDIS_URL=redis://localhost:6379
MAIL_DRIVER=log
# See .env.example for all variables
`

const gitignoreTmpl = `.env
tmp/
*.exe
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
	id SERIAL PRIMARY KEY,
{{range .Fields}}	{{.DBTag}} {{.SQLType}} NOT NULL,
{{end}}	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW()
);
`

const gftLayoutTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>#place "title" "{{.Name}}"</title>
  <link rel="stylesheet" href="/assets/app.css">
</head>
<body>
  <header>
    <h1><a href="/">{{.Name}}</a></h1>
    <nav>
      <a href="/">Home</a>
    </nav>
  </header>
  <main>
    #place "content"
  </main>
  <footer><small>Gofreight</small></footer>
</body>
</html>
`

const gftFlashPartialTmpl = `#when .Flash
<div class="flash">{= .Flash }</div>
#endwhen
`

const gftHomeTmpl = `#layout "layouts.application"

#slot "title"
Welcome to {= .Name }
#endslot

#slot "content"
  <h1>Welcome to {= .Name }!</h1>
  <p>Edit <code>app/views/home/index.gft</code> to customize this page.</p>
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

const cssTmpl = `body { font-family: system-ui, sans-serif; margin: 2rem; }
header { border-bottom: 1px solid #ddd; margin-bottom: 1rem; }
`

const testExampleTmpl = `package tests

import (
	"testing"

	"{{.Module}}/config"
	"github.com/lsgser/gofreight/gftest"
)

func TestApp(t *testing.T) {
	gftest.Describe(t, "{{.Name}}", func(d *gftest.DescribeContext) {
		var app *gftest.App

		d.BeforeEach(func(t *testing.T) {
			app = gftest.NewApp(t, gftest.WithDatabase("sqlite://:memory:"))
			app.Draw(config.Routes)
		})

		d.It("returns the homepage", func(t *testing.T) {
			app.Get("/").AssertOk()
		})
	})
}
`

const testFactoriesTmpl = `package factories

import "github.com/lsgser/gofreight/gftest"

// Add model factories here.
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
