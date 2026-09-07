package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Seeder generates a Go seeder class in db/seeders/.
func Seeder(appPath, name string) error {
	base := snakeCase(name)
	if !strings.HasSuffix(base, "_seeder") {
		base += "_seeder"
	}
	structName := structName(strings.TrimSuffix(base, "_seeder")) + "Seeder"

	dir := filepath.Join(appPath, "db", "seeders")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	mod := moduleName(appPath)
	data := struct {
		StructName, Module, FileBase string
		IsDatabaseSeeder             bool
	}{structName, mod, base, structName == "DatabaseSeeder"}

	tmpl := seederTmpl
	if data.IsDatabaseSeeder {
		tmpl = databaseSeederTmpl
	}

	if err := writeTemplate(filepath.Join(dir, base+".go"), tmpl, data); err != nil {
		return err
	}
	if err := ensureSeedCmd(appPath, mod); err != nil {
		return err
	}
	return registerSeederInCmd(appPath, structName, mod)
}

// Test generates a feature test file in tests/.
func Test(appPath, name string) error {
	base := snakeCase(name)
	if !strings.HasSuffix(base, "_test") {
		base += "_test"
	}
	structName := structName(strings.TrimSuffix(base, "_test"))
	dir := filepath.Join(appPath, "tests")
	os.MkdirAll(dir, 0755)

	return writeTemplate(filepath.Join(dir, base+".go"), testFeatureTmpl, struct {
		Name, Module, StructName string
	}{base, moduleName(appPath), structName})
}

// Factory generates a model factory in tests/factories/.
func Factory(appPath, name string) error {
	structName := structName(name)
	singular := strings.ToLower(structName)
	dir := filepath.Join(appPath, "tests", "factories")
	os.MkdirAll(dir, 0755)

	return writeTemplate(filepath.Join(dir, singular+"_factory.go"), standaloneFactoryTmpl, struct {
		Name, Module, Singular string
	}{structName, moduleName(appPath), singular})
}

func ensureSeedCmd(appPath, mod string) error {
	cmdDir := filepath.Join(appPath, "cmd", "seed")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(cmdDir, "main.go")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return writeTemplate(path, seedCmdTmpl, struct{ Module string }{mod})
}

func registerSeederInCmd(appPath, structName, mod string) error {
	path := filepath.Join(appPath, "cmd", "seed", "main.go")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	line := fmt.Sprintf("\t\t\"%s\": seeders.New%s().Run,", structName, structName)
	s := string(content)
	if strings.Contains(s, structName) {
		return nil
	}
	importLine := fmt.Sprintf("\t\"%s/db/seeders\"\n", mod)
	if !strings.Contains(s, mod+"/db/seeders") {
		s = strings.Replace(s, "\"github.com/lsgser/gofreight/database\"\n", "\"github.com/lsgser/gofreight/database\"\n"+importLine, 1)
	}
	s = strings.Replace(s, "registry := map[string]func(context.Context) error{\n\t}", "registry := map[string]func(context.Context) error{\n"+line+"\n\t}", 1)
	s = strings.Replace(s, "registry := map[string]func(context.Context) error{\n\t\n\t}", "registry := map[string]func(context.Context) error{\n"+line+"\n\t}", 1)
	// append to non-empty registry
	if strings.Contains(s, "registry := map[string]func(context.Context) error{") && !strings.Contains(s, line) {
		s = strings.Replace(s, "\n\t}\n\n\tif *class", "\n"+line+"\n\t}\n\n\tif *class", 1)
	}
	return os.WriteFile(path, []byte(s), 0644)
}

const seederTmpl = `package seeders

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Seeds the database with records. Register in cmd/seed/main.go, then run:
|
|   gofreight db:seed --class={{.StructName}}
|
*/

import (
	"context"

	"github.com/lsgser/gofreight/database"
)

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Seeds database records.
|
*/
type {{.StructName}} struct{}

/*
|--------------------------------------------------------------------------
| New{{.StructName}}
|--------------------------------------------------------------------------
|
| Creates a new seeder instance.
|
*/
func New{{.StructName}}() *{{.StructName}} { return &{{.StructName}}{} }

/*
|--------------------------------------------------------------------------
| Run
|--------------------------------------------------------------------------
|
| Executes the seeder.
|
*/
func (s *{{.StructName}}) Run(ctx context.Context) error {
	_, err := database.DB().ExecContext(ctx, "-- add seed SQL here")
	return err
}
`

const databaseSeederTmpl = `package seeders

import (
	"context"

	"github.com/lsgser/gofreight/model"
)

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Root database seeder. Call other seeders from Run (Laravel-style):
|
|   return s.Call(NewUserSeeder(), NewPostSeeder())
|
| Run with: gofreight db:seed
|
*/
type {{.StructName}} struct {
	model.Seeder
}

func New{{.StructName}}() *{{.StructName}} {
	return &{{.StructName}}{Seeder: model.NewSeeder()}
}

func (s *{{.StructName}}) Run(ctx context.Context) error {
	s.SetContext(ctx)
	return s.Call(
		// NewUserSeeder(),
	)
}
`

const seedCmdTmpl = `package main

/*
|--------------------------------------------------------------------------
| Database Seeder Command
|--------------------------------------------------------------------------
|
| Runs Go seeders from db/seeders/. Register each seeder in the map below.
|
|   gofreight db:seed
|   gofreight db:seed --class=DatabaseSeeder
|
*/

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/lsgser/gofreight/config"
	"github.com/lsgser/gofreight/database"
	"{{.Module}}/db/seeders"
)

func main() {
	class := flag.String("class", "", "Run a specific seeder class (e.g. DatabaseSeeder)")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()
	if _, err := database.Connect(cfg.DatabaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "database: %v\n", err)
		os.Exit(1)
	}

	/*
	|--------------------------------------------------------------------------
	| Seeder Registry
	|--------------------------------------------------------------------------
	|
	| Register seeders here, e.g.:
	|   "UserSeeder": seeders.NewUserSeeder().Run,
	|
	*/
	registry := map[string]func(context.Context) error{
		"DatabaseSeeder": seeders.NewDatabaseSeeder().Run,
	}

	if *class == "" {
		className := "DatabaseSeeder"
		fmt.Printf("Seeding %s...\n", className)
		if err := registry[className](context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", className, err)
			os.Exit(1)
		}
		return
	}

	run, ok := registry[*class]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown seeder %q\n", *class)
		os.Exit(1)
	}
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
`

const testFeatureTmpl = `package tests

/*
|--------------------------------------------------------------------------
| {{.StructName}} Feature Test
|--------------------------------------------------------------------------
|
| HTTP feature tests for {{.StructName}}. Run with gofreight test.
|
*/

import (
	"testing"

	"{{.Module}}/routes"
	"github.com/lsgser/gofreight/gftest"
)

func Test{{.StructName}}(t *testing.T) {
	gftest.Describe(t, "{{.StructName}}", func(d *gftest.DescribeContext) {
		var app *gftest.App

		d.BeforeEach(func(t *testing.T) {
			app = gftest.NewApp(t, gftest.WithDatabase("sqlite://:memory:"))
			app.Draw(routes.Register)
		})

		d.It("works", func(t *testing.T) {
			app.Get("/").AssertOk()
		})
	})
}
`

const standaloneFactoryTmpl = `package factories

/*
|--------------------------------------------------------------------------
| {{.Name}} Factory
|--------------------------------------------------------------------------
|
| Creates {{.Name}} instances for tests. Usage:
|
|   post := factories.Create{{.Name}}(t, map[string]any{"title": "Hello"})
|   posts := factories.{{.Name}}Factory.Count(3).Create(t)
|
*/

import (
	"testing"

	"{{.Module}}/app/models"
	"github.com/lsgser/gofreight/gftest"
	"github.com/lsgser/gofreight/gftest/faker"
)

var {{.Name}}Factory = gftest.NewFactory(models.{{.Name}}s).
	Define(faker.DefinitionsForModel("{{.Name}}"))

// Unverified{{.Name}} returns a factory state (Laravel factory states).
func Unverified{{.Name}}() *gftest.Factory[models.{{.Name}}] {
	return {{.Name}}Factory.State(map[string]any{
		// override fields for this state, e.g. "verified": false,
	})
}

func Create{{.Name}}(t *testing.T, attrs ...map[string]any) *models.{{.Name}} {
	t.Helper()
	return {{.Name}}Factory.Create(t, attrs...)
}
`
