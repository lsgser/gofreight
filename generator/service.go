package generator

/*
|--------------------------------------------------------------------------
| Service
|--------------------------------------------------------------------------
|
| Implements Service as part of the generator package in the Gofreight
| framework. Key symbols: Service.
| 
| The generator package powers gofreight new and all make:* scaffolds.
| 
| It writes idiomatic directory layouts, GFT views, migrations, tests, and
| auth stubs from templates.
| 
| CLI handlers in cmd/gofreight call into this package; templates live
| primarily in templates.go.
| 
| Symbols defined here include: Service (Service generates a service class
| in app/services/.).
| 
*/

import (
	"os"
	"path/filepath"
	"strings"
)

// Service generates a service class in app/services/.
// Services hold business logic — controllers stay thin.
func Service(appPath, name string) error {
	dir := filepath.Join(appPath, "app", "services")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	base := serviceFileName(name)
	structName := title(strings.TrimSuffix(base, "_service"))
	if !strings.HasSuffix(strings.ToLower(structName), "service") {
		structName += "Service"
	}

	module := filepath.Base(appPath)
	if module == "." {
		module = "app"
	}

	data := struct {
		StructName, Module string
	}{
		StructName: structName,
		Module:     module,
	}

	path := filepath.Join(dir, base+".go")
	return writeTemplate(path, serviceTmpl, data)
}

func serviceFileName(name string) string {
	s := strings.ToLower(name)
	s = strings.TrimSuffix(s, "service")
	s = strings.TrimSpace(s)
	if s == "" {
		s = "service"
	}
	if !strings.HasSuffix(s, "_service") {
		s += "_service"
	}
	return s
}

const serviceTmpl = `package services

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Encapsulates business logic for this domain. Register in bootstrap/app.go
| and resolve via the service container, or construct directly in tests.
|
|   app.Singleton("{{.StructName}}", func() any { return services.New{{.StructName}}() })
|
*/

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Handles domain business logic.
|
*/
type {{.StructName}} struct {
}

/*
|--------------------------------------------------------------------------
| New{{.StructName}}
|--------------------------------------------------------------------------
|
| Creates a new {{.StructName}} instance.
|
*/
func New{{.StructName}}() *{{.StructName}} {
	return &{{.StructName}}{}
}
`
