package generator

import (
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
		filepath.Join(root, "app", "services"),
		filepath.Join(root, "app", "resources"),
		filepath.Join(root, "app", "mail"),
		filepath.Join(root, "app", "jobs"),
		filepath.Join(root, "app", "middleware"),
		filepath.Join(root, "app", "policies"),
		filepath.Join(root, "app", "requests"),
		filepath.Join(root, "app", "views", "layouts"),
		filepath.Join(root, "app", "views", "partials"),
		filepath.Join(root, "app", "views", "components"),
		filepath.Join(root, "app", "views", "mail"),
		filepath.Join(root, "app", "views", "home"),
		filepath.Join(root, "routes"),
		filepath.Join(root, "bootstrap"),
		filepath.Join(root, "config"),
		filepath.Join(root, "config", "locales"),
		filepath.Join(root, "db", "migrate"),
		filepath.Join(root, "db", "seeds"),
		filepath.Join(root, "db", "seeders"),
		filepath.Join(root, "cmd", "seed"),
		filepath.Join(root, "public"),
		filepath.Join(root, "storage", "app"),
		filepath.Join(root, "storage", "framework", "cache"),
		filepath.Join(root, "storage", "framework", "sessions"),
		filepath.Join(root, "storage", "logs"),
		filepath.Join(root, "storage", "uploads"),
		filepath.Join(root, "tests", "factories"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "tools", "migrate"), 0755); err != nil {
		return err
	}

	// Keep empty app directories in git (storage uses .gitkeep only)
	gitkeepDirs := []string{
		filepath.Join(root, "storage", "app"),
		filepath.Join(root, "storage", "framework", "cache"),
		filepath.Join(root, "storage", "framework", "sessions"),
		filepath.Join(root, "storage", "logs"),
	}
	for _, dir := range gitkeepDirs {
		_ = os.WriteFile(filepath.Join(dir, ".gitkeep"), nil, 0644)
	}

	files := map[string]string{
		filepath.Join(root, "main.go"):               appMainTmpl,
		filepath.Join(root, "bootstrap", "app.go"):  bootstrapAppTmpl,
		filepath.Join(root, "bootstrap", "schedule.go"): bootstrapScheduleTmpl,
		filepath.Join(root, "routes", "register.go"): routesRegisterTmpl,
		filepath.Join(root, "routes", "web.go"):     routesWebTmpl,
		filepath.Join(root, "routes", "api.go"):     routesAPITmpl,
		filepath.Join(root, "go.mod"):                appGoModTmpl,
		filepath.Join(root, "README.md"):             appReadmeTmpl,
		filepath.Join(root, "config", "database.go"): databaseTmpl,
		filepath.Join(root, "config", "app.yaml"):    appYamlTmpl,
		filepath.Join(root, "config", "locales", "en.json"): localeEnTmpl,
		filepath.Join(root, ".env"):                  envTmpl,
		filepath.Join(root, ".env.example"):          envExampleTmpl,
		filepath.Join(root, ".gitignore"):            gitignoreTmpl,
		filepath.Join(root, "app", "views", "layouts", "application.gft"): gftLayoutTmpl,
		filepath.Join(root, "app", "views", "partials", "flash.gft"):       gftFlashPartialTmpl,
		filepath.Join(root, "app", "views", "home", "index.gft"):            gftHomeTmpl,
		filepath.Join(root, "app", "services", "example_service.go"):       exampleServiceTmpl,
		filepath.Join(root, "app", "controllers", "doc.go"):                controllersDocTmpl,
		filepath.Join(root, "app", "models", "doc.go"):                   modelsDocTmpl,
		filepath.Join(root, "app", "resources", "doc.go"):                resourcesDocTmpl,
		filepath.Join(root, "app", "mail", "doc.go"):                     mailDocTmpl,
		filepath.Join(root, "app", "jobs", "doc.go"):                     jobsDocTmpl,
		filepath.Join(root, "app", "middleware", "doc.go"):               middlewareDocTmpl,
		filepath.Join(root, "app", "policies", "doc.go"):                 policiesDocTmpl,
		filepath.Join(root, "app", "requests", "doc.go"):                 requestsDocTmpl,
		filepath.Join(root, "db", "migrate", "0001_init.go"):            initialGoMigrationTmpl,
		filepath.Join(root, "db", "migrate", "README.md"):               migrateReadmeTmpl,
		filepath.Join(root, "tools", "migrate", "main.go"):              migrateToolMainTmpl,
		filepath.Join(root, "db", "seeders", "doc.go"):                    seedersDocTmpl,
		filepath.Join(root, "public", "app.css"):                           cssTmpl,
		filepath.Join(root, "tests", "example_test.go"):                    testExampleTmpl,
		filepath.Join(root, "tests", "factories", "factories.go"):          testFactoriesTmpl,
	}

	data := struct{ Name, Module, FrameworkVersion, DocsURL string }{
		Name:             name,
		Module:           strings.ToLower(name),
		FrameworkVersion: version.Module(),
		DocsURL:          "https://lsgser.github.io/gofreight-web/",
	}

	for path, tmpl := range files {
		if err := writeTemplate(path, tmpl, data); err != nil {
			return err
		}
	}

	return Seeder(root, "DatabaseSeeder")
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
		Name, GoType, SQLType, DBTag, JSONTag string
	}
	var fieldList []Field
	for fname, ftype := range fields {
		pf := ParseField(fname, ftype)
		fieldList = append(fieldList, Field{
			Name:    pf.Name,
			GoType:  pf.GoType,
			SQLType: pf.MigrationColumnDef(),
			DBTag:   pf.DBTag,
			JSONTag: pf.JSONTag,
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

	migrationPath, err := CreateBlueprintMigration(migrateDir, "create_"+table+"_table", fields)
	if err != nil {
		return err
	}
	_ = migrationPath
	return nil
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

