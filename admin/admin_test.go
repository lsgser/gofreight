package admin_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/lsgser/gofreight/admin"
	"github.com/lsgser/gofreight/database"
	"github.com/lsgser/gofreight/gftest"
	"github.com/lsgser/gofreight/router"
)

func setupAdmin(t *testing.T) *gftest.App {
	t.Helper()
	database.Reset()
	app := gftest.NewApp(t,
		gftest.WithDatabase("sqlite://:memory:"),
		gftest.WithMigrations(`
			CREATE TABLE posts (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				title TEXT NOT NULL,
				body TEXT NOT NULL
			);
			INSERT INTO posts (title, body) VALUES ('Hello', 'World');
		`),
	)
	panel := admin.MustNew(admin.DefaultConfig(true))
	panel.Mount(app.Router)
	return app
}

func TestAdminDashboard(t *testing.T) {
	app := setupAdmin(t)
	app.Get("/admin").AssertOk().AssertSee("Database Tables").AssertSee("posts")
}

func TestAdminTableIndex(t *testing.T) {
	app := setupAdmin(t)
	app.Get("/admin/tables/posts").AssertOk().AssertSee("Hello").AssertSee("World")
}

func TestAdminTableShow(t *testing.T) {
	app := setupAdmin(t)
	app.Get("/admin/tables/posts/1").AssertOk().AssertSee("Hello")
}

func TestAdminTableNew(t *testing.T) {
	app := setupAdmin(t)
	app.Get("/admin/tables/posts/new").AssertOk().AssertSee("Create posts")
}

func TestAdminTableCreate(t *testing.T) {
	app := setupAdmin(t)
	form := url.Values{"title": {"New Post"}, "body": {"New body content here."}}
	h := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}
	resp := app.Request(http.MethodPost, "/admin/tables/posts", strings.NewReader(form.Encode()), h)
	if resp.Status() != http.StatusSeeOther && resp.Status() != http.StatusFound {
		t.Fatalf("expected redirect, got %d: %s", resp.Status(), resp.Body())
	}
	app.Get("/admin/tables/posts").AssertSee("New Post")
}

func TestAdminSQLConsole(t *testing.T) {
	app := setupAdmin(t)
	app.Get("/admin/sql").AssertOk().AssertSee("SQL Console")
}

func TestAdminNotMountedInProduction(t *testing.T) {
	r := router.New()
	panel, err := admin.New(admin.Config{Development: false})
	if err != nil {
		t.Fatal(err)
	}
	panel.Mount(r)

	routes := r.Routes()
	for _, route := range routes {
		if strings.Contains(route, "/admin") {
			t.Fatalf("admin should not mount in production, got route: %s", route)
		}
	}
}

func TestAdminTableUpdate(t *testing.T) {
	app := setupAdmin(t)
	form := url.Values{"title": {"Updated Title"}, "body": {"Updated body content."}}
	h := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}
	resp := app.Request(http.MethodPost, "/admin/tables/posts/1", strings.NewReader(form.Encode()), h)
	if resp.Status() != http.StatusSeeOther && resp.Status() != http.StatusFound {
		t.Fatalf("expected redirect, got %d", resp.Status())
	}
	app.Get("/admin/tables/posts/1").AssertSee("Updated Title")
}

func TestAdminTableDelete(t *testing.T) {
	app := setupAdmin(t)
	resp := app.Post("/admin/tables/posts/1/delete", strings.NewReader(""))
	if resp.Status() != http.StatusSeeOther && resp.Status() != http.StatusFound {
		t.Fatalf("expected redirect, got %d", resp.Status())
	}
	app.DB().ToHaveCount("posts", 0)
}

func TestAdminSchemaNew(t *testing.T) {
	app := setupAdmin(t)
	app.Get("/admin/schema/new").AssertOk().AssertSee("Create Table")
}

func TestAdminSchemaCreate(t *testing.T) {
	app := setupAdmin(t)
	form := url.Values{
		"table_name":   {"tags"},
		"column_count": {"1"},
		"col_0_name":   {"label"},
		"col_0_type":   {"TEXT"},
	}
	h := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}
	resp := app.Request(http.MethodPost, "/admin/schema/create", strings.NewReader(form.Encode()), h)
	if resp.Status() != http.StatusSeeOther && resp.Status() != http.StatusFound {
		t.Fatalf("expected redirect, got %d: %s", resp.Status(), resp.Body())
	}
	app.Get("/admin/tables/tags").AssertOk().AssertSee("label")
}

func TestAdminTableStructure(t *testing.T) {
	app := setupAdmin(t)
	app.Get("/admin/tables/posts/structure").AssertOk().AssertSee("Structure").AssertSee("title")
}

func TestAdminImportSQL(t *testing.T) {
	app := setupAdmin(t)
	form := url.Values{"sql": {"CREATE TABLE extras (id INTEGER PRIMARY KEY, name TEXT);"}}
	h := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}
	resp := app.Request(http.MethodPost, "/admin/import", strings.NewReader(form.Encode()), h)
	if resp.Status() != http.StatusSeeOther && resp.Status() != http.StatusFound {
		t.Fatalf("expected redirect, got %d", resp.Status())
	}
	app.Get("/admin").AssertSee("extras")
}

func TestAdminIntegrations(t *testing.T) {
	app := setupAdmin(t)
	app.Get("/admin/integrations").AssertOk().AssertSee("storage").AssertSee("email")
}
