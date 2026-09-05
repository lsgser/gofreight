package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResourceData holds scaffold template data.
type ResourceData struct {
	Name, Module, Plural, Singular, Table, Title string
	Fields                                         []ResourceField
}

type ResourceField struct {
	ParsedField
}

func buildResourceData(appPath, name string, fields map[string]string) ResourceData {
	plural := pluralize(strings.ToLower(name))
	module := filepath.Base(appPath)
	if module == "." {
		module = "app"
	}
	var flist []ResourceField
	for fname, ftype := range fields {
		pf := ParseField(fname, ftype)
		pf.SQLType = pf.MigrationColumnDef()
		flist = append(flist, ResourceField{ParsedField: pf})
	}
	return ResourceData{
		Name:     title(name),
		Module:   module,
		Plural:   plural,
		Singular: strings.ToLower(name),
		Table:    plural,
		Title:    title(plural),
		Fields:   flist,
	}
}

func formFieldHTML(f ResourceField) string {
	switch f.FormType {
	case "textarea":
		return fmt.Sprintf(`  <div class="form-group">
    <label for="%s">%s</label>
    <textarea name="%s" id="%s">{= .Item.%s }</textarea>
  </div>
`, f.DBTag, f.Name, f.DBTag, f.DBTag, f.Name)
	case "checkbox":
		return fmt.Sprintf(`  <div class="form-group">
    <label><input type="checkbox" name="%s" value="1"> %s</label>
  </div>
`, f.DBTag, f.Name)
	case "select":
		var opts strings.Builder
		for _, v := range f.EnumValues {
			opts.WriteString(fmt.Sprintf("    <option value=\"%s\">%s</option>\n", v, title(v)))
		}
		return fmt.Sprintf(`  <div class="form-group">
    <label for="%s">%s</label>
    <select name="%s" id="%s">
%s    </select>
  </div>
`, f.DBTag, f.Name, f.DBTag, f.DBTag, opts.String())
	default:
		inputType := f.HTMLInputType()
		return fmt.Sprintf(`  <div class="form-group">
    <label for="%s">%s</label>
    <input type="%s" name="%s" id="%s" value="{= .Item.%s }">
  </div>
`, f.DBTag, f.Name, inputType, f.DBTag, f.DBTag, f.Name)
	}
}
func Resource(appPath, name string, fields map[string]string) error {
	data := buildResourceData(appPath, name, fields)

	if err := writeResourceModel(appPath, data); err != nil {
		return err
	}
	if err := writeResourceMigration(appPath, data); err != nil {
		return err
	}
	if err := writeResourceController(appPath, data); err != nil {
		return err
	}
	if err := writeResourceViews(appPath, data); err != nil {
		return err
	}
	if err := writeResourceFactory(appPath, data); err != nil {
		return err
	}
	if err := writeResourceTest(appPath, data); err != nil {
		return err
	}
	return appendRouteRegistration(appPath, data)
}

func writeResourceModel(appPath string, data ResourceData) error {
	dir := filepath.Join(appPath, "app", "models")
	os.MkdirAll(dir, 0755)
	return writeTemplate(filepath.Join(dir, data.Singular+".go"), resourceModelTmpl, data)
}

func writeResourceMigration(appPath string, data ResourceData) error {
	dir := filepath.Join(appPath, "db", "migrate")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, fmt.Sprintf("001_create_%s.sql", data.Table))
	return writeTemplate(path, migrationSQLTmpl, data)
}

func writeResourceController(appPath string, data ResourceData) error {
	dir := filepath.Join(appPath, "app", "controllers")
	os.MkdirAll(dir, 0755)
	return writeTemplate(filepath.Join(dir, data.Singular+"_controller.go"), resourceControllerTmpl, data)
}

func writeResourceViews(appPath string, data ResourceData) error {
	dir := filepath.Join(appPath, "app", "views", data.Plural)
	os.MkdirAll(dir, 0755)

	fieldLines := ""
	formFields := ""
	showFields := ""
	for _, f := range data.Fields {
		fieldLines += fmt.Sprintf("    <p><strong>%s:</strong> {= .%s }</p>\n", f.Name, f.Name)
		showFields += fmt.Sprintf("<p><strong>%s:</strong> {= .Item.%s }</p>\n", f.Name, f.Name)
		formFields += formFieldHTML(f)
	}

	views := map[string]string{
		"index.gft": fmt.Sprintf(`#layout "layouts.application"

#slot "content"
<div class="page-header">
  <h1>%s</h1>
  <a href="/%s/new" class="btn">+ New %s</a>
</div>

#partial "partials.flash"

#eachor .Items as item
  <article class="card">
    <h2><a href="/%s/{= .ID }">#{= .ID }</a></h2>
%s    <a href="/%s/{= .ID }/edit">Edit</a>
  </article>
#otherwise
  <p>No records yet. <a href="/%s/new">Create one</a>.</p>
#endeach
#endslot
`, data.Title, data.Plural, data.Name, data.Plural, fieldLines, data.Plural, data.Plural),

		"show.gft": fmt.Sprintf(`#layout "layouts.application"

#slot "content"
<h1>%s #{= .Item.ID }</h1>
#partial "partials.flash"
%s
<a href="/%s">Back</a>
<a href="/%s/{= .Item.ID }/edit">Edit</a>
#endslot
`, data.Name, showFields, data.Plural, data.Plural),

		"new.gft": fmt.Sprintf(`#layout "layouts.application"

#slot "content"
<h1>New %s</h1>
#partial "partials.flash"
<form method="POST" action="/%s">
  #token
%s  <button type="submit">Create</button>
</form>
#endslot
`, data.Name, data.Plural, formFields),

		"edit.gft": fmt.Sprintf(`#layout "layouts.application"

#slot "content"
<h1>Edit %s</h1>
#partial "partials.flash"
<form method="POST" action="/%s/{= .Item.ID }">
  #token
%s  <button type="submit">Save</button>
</form>
#endslot
`, data.Name, data.Plural, formFields),
	}

	for file, content := range views {
		if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0644); err != nil {
			return err
		}
	}

	pdir := filepath.Join(appPath, "app", "views", "partials")
	os.MkdirAll(pdir, 0755)
	return os.WriteFile(filepath.Join(pdir, "flash.gft"), []byte(`#when .Flash
<div class="flash">{= .Flash }</div>
#endwhen
`), 0644)
}

func writeResourceFactory(appPath string, data ResourceData) error {
	dir := filepath.Join(appPath, "tests", "factories")
	os.MkdirAll(dir, 0755)
	return writeTemplate(filepath.Join(dir, data.Singular+"_factory.go"), resourceFactoryTmpl, data)
}

func writeResourceTest(appPath string, data ResourceData) error {
	dir := filepath.Join(appPath, "tests")
	os.MkdirAll(dir, 0755)
	return writeTemplate(filepath.Join(dir, data.Singular+"_test.go"), resourceTestTmpl, data)
}

func appendRouteRegistration(appPath string, data ResourceData) error {
	path := filepath.Join(appPath, "config", "routes.go")
	content, err := os.ReadFile(path)
	if err != nil {
		snippet := fmt.Sprintf("\n\tcontrollers.Register%sRoutes(r)\n", data.Name)
		return os.WriteFile(filepath.Join(appPath, "config", "routes_snippet.txt"), []byte(snippet), 0644)
	}
	importLine := fmt.Sprintf("\t\"%s/app/controllers\"\n", data.Module)
	registerLine := fmt.Sprintf("\tcontrollers.Register%sRoutes(r)\n", data.Name)
	s := string(content)
	if !strings.Contains(s, registerLine) {
		if !strings.Contains(s, data.Module+"/app/controllers") {
			s = strings.Replace(s, "import (\n", "import (\n"+importLine, 1)
		}
		s = strings.Replace(s, "// Add your routes here:", "// Add your routes here:\n"+registerLine, 1)
	}
	return os.WriteFile(path, []byte(s), 0644)
}

// Scaffold is an alias for Resource (Laravel make:scaffold style).
func Scaffold(appPath, name string, fields map[string]string) error {
	return Resource(appPath, name, fields)
}

const resourceModelTmpl = `package models

import (
	"context"

	"github.com/lsgser/gofreight/model"
)

// {{.Name}} represents a {{.Singular}} record.
type {{.Name}} struct {
	model.Record
{{range .Fields}}	{{.Name}} {{.GoType}} ` + "`db:\"{{.DBTag}}\" json:\"{{.JSONTag}}\"`" + `
{{end}}}

// {{.Name}}s is the repository for {{.Name}}.
var {{.Name}}s = model.NewRepository[{{.Name}}]("{{.Table}}")

// Validators returns validation rules for {{.Name}}.
func (m *{{.Name}}) Validators() []model.Validator {
	return []model.Validator{
{{range .Fields}}		model.Presence("{{.Name}}"),
{{end}}	}
}

// Save validates and persists the record.
func (m *{{.Name}}) Save(ctx context.Context) error {
	return {{.Name}}s.Save(ctx, m)
}
`

const resourceControllerTmpl = `package controllers

import (
	"context"
	"net/http"
	"strconv"

	"{{.Module}}/app/models"
	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
	"github.com/lsgser/gofreight/validation"
)

// {{.Name}}Controller handles {{.Plural}} resources.
type {{.Name}}Controller struct{}

// Register{{.Name}}Routes registers RESTful routes for {{.Plural}}.
func Register{{.Name}}Routes(r *router.Router) {
	c := {{.Name}}Controller{}
	r.Resources("{{.Plural}}", router.ResourceHandlers{
		Index:   controller.Handler(c.Index),
		New:     controller.Handler(c.New),
		Create:  controller.Handler(c.Create),
		Show:    controller.Handler(c.Show),
		Edit:    controller.Handler(c.Edit),
		Update:  controller.Handler(c.Update),
		Destroy: controller.Handler(c.Destroy),
	})
}

func (c {{.Name}}Controller) Index(base controller.Base) error {
	items, err := models.{{.Name}}s.All(context.Background())
	if err != nil {
		return err
	}
	return base.RenderView("{{.Plural}}/index", map[string]any{
		"Title":  "{{.Title}}",
		"Items":  items,
	})
}

func (c {{.Name}}Controller) Show(base controller.Base) error {
	item, err := c.find(base)
	if err != nil {
		return nil
	}
	return base.RenderView("{{.Plural}}/show", map[string]any{
		"Title":  "{{.Name}}",
		"Item":   item,
	})
}

func (c {{.Name}}Controller) New(base controller.Base) error {
	return base.RenderView("{{.Plural}}/new", map[string]any{
		"Title": "New {{.Name}}",
		"Item":  &models.{{.Name}}{},
	})
}

func (c {{.Name}}Controller) Create(base controller.Base) error {
	item := &models.{{.Name}}{}
	if err := c.bind(base, item); err != nil {
		return nil
	}
	if err := item.Save(context.Background()); err != nil {
		base.Unprocessable(map[string][]string{"base": {err.Error()}})
		return nil
	}
	base.Redirect("/{{.Plural}}/"+strconv.FormatInt(item.ID, 10), http.StatusSeeOther)
	return nil
}

func (c {{.Name}}Controller) Edit(base controller.Base) error {
	item, err := c.find(base)
	if err != nil {
		return nil
	}
	return base.RenderView("{{.Plural}}/edit", map[string]any{
		"Title": "Edit {{.Name}}",
		"Item":  item,
	})
}

func (c {{.Name}}Controller) Update(base controller.Base) error {
	item, err := c.find(base)
	if err != nil {
		return nil
	}
	if err := c.bind(base, item); err != nil {
		return nil
	}
	if err := item.Save(context.Background()); err != nil {
		base.Unprocessable(map[string][]string{"base": {err.Error()}})
		return nil
	}
	base.Redirect("/{{.Plural}}/"+base.Param("id"), http.StatusSeeOther)
	return nil
}

func (c {{.Name}}Controller) Destroy(base controller.Base) error {
	id, _ := strconv.ParseInt(base.Param("id"), 10, 64)
	if err := models.{{.Name}}s.Delete(context.Background(), id); err != nil {
		return err
	}
	base.Redirect("/{{.Plural}}", http.StatusSeeOther)
	return nil
}

func (c {{.Name}}Controller) find(base controller.Base) (*models.{{.Name}}, error) {
	id, err := strconv.ParseInt(base.Param("id"), 10, 64)
	if err != nil {
		base.NotFound("invalid id")
		return nil, err
	}
	item, err := models.{{.Name}}s.Find(context.Background(), id)
	if err != nil {
		base.NotFound("{{.Name}} not found")
		return nil, err
	}
	return item, nil
}

func (c {{.Name}}Controller) bind(base controller.Base, item *models.{{.Name}}) error {
	v, err := validation.New(base.Request)
	if err != nil {
		return err
	}
{{range .Fields}}	item.{{.Name}} = v.Get("{{.DBTag}}")
{{end}}	if v.Fails() {
		base.Unprocessable(v.Errors())
		return err
	}
	return nil
}
`

const resourceFactoryTmpl = `package factories

import (
	"testing"

	"{{.Module}}/app/models"
	"github.com/lsgser/gofreight/gftest"
)

var {{.Name}}Factory = gftest.NewFactory(models.{{.Name}}s).Define(map[string]any{
{{range .Fields}}	"{{.DBTag}}": "sample {{.DBTag}}",
{{end}}})

func Create{{.Name}}(t *testing.T, attrs ...map[string]any) *models.{{.Name}} {
	t.Helper()
	return {{.Name}}Factory.Create(t, attrs...)
}
`

const resourceTestTmpl = `package tests

import (
	"testing"

	"{{.Module}}/config"
	"github.com/lsgser/gofreight/gftest"
)

func Test{{.Name}}Resource(t *testing.T) {
	gftest.Describe(t, "{{.Title}}", func(d *gftest.DescribeContext) {
		var app *gftest.App

		d.BeforeEach(func(t *testing.T) {
			app = gftest.NewApp(t, gftest.WithDatabase("sqlite://:memory:"))
			app.Draw(config.Routes)
		})

		d.It("lists {{.Plural}}", func(t *testing.T) {
			app.Get("/{{.Plural}}").AssertOk()
		})

		d.It("shows new form", func(t *testing.T) {
			app.Get("/{{.Plural}}/new").AssertOk()
		})
	})
}
`
