package gftest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/lsgser/gofreight/application"
	"github.com/lsgser/gofreight/database"
	"github.com/lsgser/gofreight/router"
	"github.com/lsgser/gofreight/view"
)

// App is a test application with HTTP helpers and database access.
type App struct {
	T      *testing.T
	App    *application.Application
	Router *router.Router
}

// AppOption configures test app creation.
type AppOption func(*appConfig)

type appConfig struct {
	databaseURL string
	migrations  []string
	migrateDir  string
	viewsRoot   string
	chdir       string
	env         string
}

// WithDatabase sets the database URL.
func WithDatabase(url string) AppOption {
	return func(c *appConfig) { c.databaseURL = url }
}

// WithMigrations provides inline SQL migrations.
func WithMigrations(sql ...string) AppOption {
	return func(c *appConfig) { c.migrations = sql }
}

// WithMigrateDir sets the migrations directory to run on setup.
func WithMigrateDir(dir string) AppOption {
	return func(c *appConfig) { c.migrateDir = dir }
}

// WithViewsRoot sets the views directory path.
func WithViewsRoot(path string) AppOption {
	return func(c *appConfig) { c.viewsRoot = path }
}

// WithChdir changes working directory during app setup (for tests in subpackages).
func WithChdir(dir string) AppOption {
	return func(c *appConfig) { c.chdir = dir }
}

// NewApp creates a fully configured test application.
func NewApp(t *testing.T, opts ...AppOption) *App {
	t.Helper()

	cfg := &appConfig{
		databaseURL: "sqlite://:memory:",
		env:         "test",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	t.Setenv("GOFREIGHT_ENV", cfg.env)
	t.Setenv("DATABASE_URL", cfg.databaseURL)

	if cfg.chdir != "" {
		if err := os.Chdir(cfg.chdir); err != nil {
			t.Fatalf("chdir: %v", err)
		}
	}

	database.Reset()

	app := application.New()
	if cfg.databaseURL != "" {
		if err := app.ConnectDatabase(); err != nil {
			t.Fatalf("connect database: %v", err)
		}
	}

	if len(cfg.migrations) > 0 {
		if err := database.Migrate(cfg.migrations...); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}

	if cfg.migrateDir != "" {
		if err := RunMigrations(cfg.migrateDir); err != nil {
			t.Fatalf("migrate dir: %v", err)
		}
	}

	if cfg.viewsRoot != "" {
		app.Views = view.New(cfg.viewsRoot)
	}

	if err := app.LoadViews(); err != nil {
		t.Logf("views not loaded: %v", err)
	}

	return &App{T: t, App: app, Router: app.Router}
}

// Draw registers routes.
func (a *App) Draw(routes func(*router.Router)) {
	a.App.Draw(routes)
}

// Request performs an HTTP request and returns a fluent Response.
func (a *App) Request(method, path string, body io.Reader, headers ...http.Header) *Response {
	a.T.Helper()
	req := httptest.NewRequest(method, path, body)
	if len(headers) > 0 {
		for k, vals := range headers[0] {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
	}
	rec := httptest.NewRecorder()
	a.Router.ServeHTTP(rec, req)
	return &Response{T: a.T, Recorder: rec}
}

// Get performs GET.
func (a *App) Get(path string, headers ...http.Header) *Response {
	return a.Request(http.MethodGet, path, nil, headers...)
}

// Post performs POST with a body reader.
func (a *App) Post(path string, body io.Reader, headers ...http.Header) *Response {
	return a.Request(http.MethodPost, path, body, headers...)
}

// PostJSON performs POST with JSON body.
func (a *App) PostJSON(path string, payload any) *Response {
	a.T.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		a.T.Fatalf("marshal JSON: %v", err)
	}
	h := http.Header{"Content-Type": {"application/json"}}
	return a.Request(http.MethodPost, path, bytes.NewReader(data), h)
}

// PutJSON performs PUT with JSON body.
func (a *App) PutJSON(path string, payload any) *Response {
	a.T.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		a.T.Fatalf("marshal JSON: %v", err)
	}
	h := http.Header{"Content-Type": {"application/json"}}
	return a.Request(http.MethodPut, path, bytes.NewReader(data), h)
}

// PatchJSON performs PATCH with JSON body.
func (a *App) PatchJSON(path string, payload any) *Response {
	a.T.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		a.T.Fatalf("marshal JSON: %v", err)
	}
	h := http.Header{"Content-Type": {"application/json"}}
	return a.Request(http.MethodPatch, path, bytes.NewReader(data), h)
}

// Delete performs DELETE.
func (a *App) Delete(path string) *Response {
	return a.Request(http.MethodDelete, path, nil)
}

// JSON sets Accept: application/json header helper.
func JSON() http.Header {
	return http.Header{"Accept": {"application/json"}}
}

// Context returns a background context.
func (a *App) Context() context.Context {
	return context.Background()
}

// DB returns the database connection for direct queries in tests.
func (a *App) DB() *DatabaseAssertions {
	return &DatabaseAssertions{T: a.T}
}

// Response wraps httptest.ResponseRecorder with fluent assertions.
type Response struct {
	T        *testing.T
	Recorder *httptest.ResponseRecorder
}

// AssertStatus asserts HTTP status code.
func (r *Response) AssertStatus(code int) *Response {
	r.T.Helper()
	Expect(r.Recorder.Code).Bind(r.T).ToEqual(code)
	return r
}

// AssertOk asserts 200 OK.
func (r *Response) AssertOk() *Response {
	return r.AssertStatus(http.StatusOK)
}

// AssertCreated asserts 201 Created.
func (r *Response) AssertCreated() *Response {
	return r.AssertStatus(http.StatusCreated)
}

// AssertNotFound asserts 404.
func (r *Response) AssertNotFound() *Response {
	return r.AssertStatus(http.StatusNotFound)
}

// AssertUnprocessable asserts 422.
func (r *Response) AssertUnprocessable() *Response {
	return r.AssertStatus(http.StatusUnprocessableEntity)
}

// AssertRedirect asserts a redirect status.
func (r *Response) AssertRedirect() *Response {
	code := r.Recorder.Code
	pass := code >= 300 && code < 400
	Expect(pass).Bind(r.T).ToBeTrue()
	return r
}

// AssertSee asserts body contains text (like Laravel assertSee).
func (r *Response) AssertSee(text string) *Response {
	r.T.Helper()
	Expect(r.Recorder.Body.String()).Bind(r.T).ToContain(text)
	return r
}

// AssertDontSee asserts body does not contain text.
func (r *Response) AssertDontSee(text string) *Response {
	r.T.Helper()
	Expect(r.Recorder.Body.String()).Bind(r.T).Not().ToContain(text)
	return r
}

// AssertJSON decodes JSON body into dest.
func (r *Response) AssertJSON(dest any) *Response {
	r.T.Helper()
	if err := json.NewDecoder(r.Recorder.Body).Decode(dest); err != nil {
		r.T.Fatalf("decode JSON: %v\nbody: %s", err, r.Recorder.Body.String())
	}
	return r
}

// AssertHeader asserts a response header value.
func (r *Response) AssertHeader(key, value string) *Response {
	r.T.Helper()
	got := r.Recorder.Header().Get(key)
	Expect(got).Bind(r.T).ToEqual(value)
	return r
}

// AssertJsonPath asserts a top-level JSON key value (simple path).
func (r *Response) AssertJsonPath(key string, expected any) *Response {
	r.T.Helper()
	var data map[string]any
	r.AssertJSON(&data)
	Expect(data[key]).Bind(r.T).ToEqual(expected)
	return r
}

// Body returns the response body string.
func (r *Response) Body() string {
	return r.Recorder.Body.String()
}

// Status returns the status code.
func (r *Response) Status() int {
	return r.Recorder.Code
}
