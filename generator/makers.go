package generator

/*
|--------------------------------------------------------------------------
| Makers
|--------------------------------------------------------------------------
|
| Implements Makers as part of the generator package in the Gofreight
| framework. Key symbols: Mail, Job, Middleware, Policy, Request.
| 
| The generator package powers gofreight new and all make:* scaffolds.
| 
| It writes idiomatic directory layouts, GFT views, migrations, tests, and
| auth stubs from templates.
| 
| CLI handlers in cmd/gofreight call into this package; templates live
| primarily in templates.go.
| 
| Symbols defined here include: Mail (Mail generates a mailable in
| app/mail/ and a view template.); Job (Job generates an app job in
| app/jobs/.); Middleware (Middleware generates app middleware in
| app/middleware/.); Policy (Policy generates an authorization policy in
| app/policies/.); Request (Request generates a form request in
| app/requests/.).
| 
*/

import (
	"os"
	"path/filepath"
	"strings"
)

func moduleName(appPath string) string {
	module := filepath.Base(appPath)
	if module == "." {
		return "app"
	}
	return module
}

// Mail generates a mailable in app/mail/ and a view template.
func Mail(appPath, name string) error {
	base := snakeCase(name)
	if !strings.HasSuffix(base, "_mail") {
		base += "_mail"
	}
	structName := structName(strings.TrimSuffix(base, "_mail")) + "Mail"

	mailDir := filepath.Join(appPath, "app", "mail")
	viewDir := filepath.Join(appPath, "app", "views", "mail")
	os.MkdirAll(mailDir, 0755)
	os.MkdirAll(viewDir, 0755)

	data := struct{ StructName, Module, ViewName string }{
		StructName: structName,
		Module:     moduleName(appPath),
		ViewName:   "mail/" + base + ".gft",
	}
	if err := writeTemplate(filepath.Join(mailDir, base+".go"), mailTmpl, data); err != nil {
		return err
	}
	layoutDir := filepath.Join(appPath, "app", "views", "layouts", "mail")
	os.MkdirAll(layoutDir, 0755)
	layoutPath := filepath.Join(layoutDir, "default.gft")
	if _, err := os.Stat(layoutPath); os.IsNotExist(err) {
		if err := writeTemplate(layoutPath, mailLayoutDefaultTmpl, nil); err != nil {
			return err
		}
	}
	return writeTemplate(filepath.Join(viewDir, base+".gft"), mailViewTmpl, data)
}

// Job generates an app job in app/jobs/.
func Job(appPath, name string) error {
	base := snakeCase(name)
	if !strings.HasSuffix(base, "_job") {
		base += "_job"
	}
	structName := structName(strings.TrimSuffix(base, "_job")) + "Job"
	dir := filepath.Join(appPath, "app", "jobs")
	os.MkdirAll(dir, 0755)
	return writeTemplate(filepath.Join(dir, base+".go"), jobTmpl, struct {
		StructName, Module, JobName string
	}{structName, moduleName(appPath), base})
}

// Middleware generates app middleware in app/middleware/.
func Middleware(appPath, name string) error {
	base := snakeCase(name)
	if !strings.HasSuffix(base, "_middleware") {
		base += "_middleware"
	}
	structName := structName(strings.TrimSuffix(base, "_middleware")) + "Middleware"
	dir := filepath.Join(appPath, "app", "middleware")
	os.MkdirAll(dir, 0755)
	return writeTemplate(filepath.Join(dir, base+".go"), middlewareTmpl, struct{ StructName string }{structName})
}

// Policy generates an authorization policy in app/policies/.
func Policy(appPath, name string) error {
	base := snakeCase(name)
	if !strings.HasSuffix(base, "_policy") {
		base += "_policy"
	}
	structName := structName(strings.TrimSuffix(base, "_policy")) + "Policy"
	dir := filepath.Join(appPath, "app", "policies")
	os.MkdirAll(dir, 0755)
	return writeTemplate(filepath.Join(dir, base+".go"), policyTmpl, struct{ StructName string }{structName})
}

// Request generates a form request in app/requests/.
func Request(appPath, name string) error {
	base := snakeCase(name)
	if !strings.HasSuffix(base, "_request") {
		base += "_request"
	}
	structName := structName(strings.TrimSuffix(base, "_request")) + "Request"
	dir := filepath.Join(appPath, "app", "requests")
	os.MkdirAll(dir, 0755)
	return writeTemplate(filepath.Join(dir, base+".go"), requestTmpl, struct{ StructName string }{structName})
}

const mailTmpl = `package mail

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Mailable email class. Template: app/views/mail/{{.ViewName}}
| Send via mailer configured in .env (MAIL_DRIVER).
|
*/

import (
	gofmail "github.com/lsgser/gofreight/mail"
)

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Mailable email class.
|
*/
type {{.StructName}} struct {
	To      []string
	Subject string
	Data    map[string]any
}

func New{{.StructName}}(to ...string) *{{.StructName}} {
	return &{{.StructName}}{
		To:      to,
		Subject: "{{.StructName}}",
		Data:    make(map[string]any),
	}
}

func (m *{{.StructName}}) Send(mailer gofmail.Mailer) error {
	msg := gofmail.NewMailable(gofmail.DefaultViewsRoot, "{{.ViewName}}", m.Subject, m.To...)
	for k, v := range m.Data {
		msg.With(k, v)
	}
	return msg.Send(mailer)
}
`

const mailViewTmpl = `{#
|--------------------------------------------------------------------------
| Mailable View
|--------------------------------------------------------------------------
|
| GFT content for a generated mailable. Uses layouts.mail.default.
|
#}
#layout "layouts.mail.default"

#slot "content"
<h1>{{.StructName}}</h1>
<p>Hello,</p>
#endslot
`

const jobTmpl = `package jobs

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Queueable job. Registered by name for Redis workers. Process with:
|   gofreight queue:work
|
*/

import (
	"context"

	"github.com/lsgser/gofreight/jobs"
)

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Handles background work.
|
*/
type {{.StructName}} struct{}

func New{{.StructName}}() *{{.StructName}} { return &{{.StructName}}{} }

func (j *{{.StructName}}) Handle(ctx context.Context) error {
	return nil
}

func init() {
	jobs.RegisterJob("{{.JobName}}", New{{.StructName}}().Handle)
}
`

const middlewareTmpl = `package middleware

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| HTTP middleware — runs before the controller. Register in bootstrap/app.go
| or on specific route groups.
|
*/

import "net/http"

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Application HTTP middleware.
|
*/
func {{.StructName}}(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
`

const policyTmpl = `package policies

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Authorization policy — return true/false for each action on a resource.
|
*/

import "net/http"

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Authorizes actions for a resource.
|
*/
type {{.StructName}} struct{}

func New{{.StructName}}() *{{.StructName}} { return &{{.StructName}}{} }

func (p *{{.StructName}}) View(r *http.Request) bool   { return true }
func (p *{{.StructName}}) Create(r *http.Request) bool { return true }
func (p *{{.StructName}}) Update(r *http.Request) bool { return true }
func (p *{{.StructName}}) Delete(r *http.Request) bool { return true }
`

const requestTmpl = `package requests

/*
|--------------------------------------------------------------------------
| {{.StructName}}
|--------------------------------------------------------------------------
|
| Validator for incoming request data — use with controller.ValidateUsing.
|
*/

import "github.com/lsgser/gofreight/vine"

/*
|--------------------------------------------------------------------------
| {{.StructName}}Validator
|--------------------------------------------------------------------------
|
| Validates {{.StructName}} input.
|
*/
var {{.StructName}}Validator = vine.Object(map[string]vine.Rule{
	"email": vine.String().Required().Email(),
})
`
