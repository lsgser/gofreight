package tests

import (
	"net/http"
	"testing"

	"blog/config"
	"blog/tests/factories"
	"github.com/gofreight/gofreight/gftest"
)

const postsMigration = `CREATE TABLE posts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	body TEXT NOT NULL,
	created_at TEXT DEFAULT (datetime('now')),
	updated_at TEXT DEFAULT (datetime('now'))
);
CREATE TABLE comments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	post_id INTEGER NOT NULL,
	body TEXT NOT NULL,
	author TEXT,
	created_at TEXT DEFAULT (datetime('now')),
	updated_at TEXT DEFAULT (datetime('now'))
);`

func newTestApp(t *testing.T) *gftest.App {
	t.Helper()
	app := gftest.NewApp(t,
		gftest.WithDatabase("sqlite://:memory:"),
		gftest.WithMigrations(postsMigration),
		gftest.WithViewsRoot("../app/views"),
	)
	app.Draw(config.Routes)
	return app
}

func TestPostsSuite(t *testing.T) {
	gftest.Describe(t, "Posts", func(d *gftest.DescribeContext) {
		var app *gftest.App

		d.BeforeEach(func(t *testing.T) {
			app = gftest.NewApp(t,
				gftest.WithDatabase("sqlite://:memory:"),
				gftest.WithMigrations(postsMigration),
				gftest.WithViewsRoot("../app/views"),
			)
			app.Draw(config.Routes)
		})

		d.It("lists posts as HTML", func(t *testing.T) {
			factories.PostFactory.Create(t)

			app.Get("/posts").
				AssertOk().
				AssertSee("All Posts")
		})

		d.It("creates a valid post via JSON API", func(t *testing.T) {
			app.PostJSON("/posts", map[string]string{
				"title": "Hello World",
				"body":  "This is a valid post body.",
			}).AssertCreated()

			app.DB().ToHaveCount("posts", 1)
		})

		d.ItWith("rejects invalid posts", gftest.Datasets(map[string]map[string]any{
			"short title": {"title": "ab", "body": "Valid body content here."},
			"short body":  {"title": "Valid Title", "body": "short"},
		}), func(t *testing.T, ds gftest.Dataset) {
			app.PostJSON("/posts", ds.Data).AssertUnprocessable()
			app.DB().ToBeEmpty("posts")
		})

		d.It("returns 404 for missing posts", func(t *testing.T) {
			app.Get("/posts/999", gftest.JSON()).AssertNotFound()
		})
	})
}

func TestHomeSuite(t *testing.T) {
	gftest.Describe(t, "Home", func(d *gftest.DescribeContext) {
		d.It("renders the welcome page", func(t *testing.T) {
			app := newTestApp(t)
			app.Get("/").AssertOk().AssertSee("Welcome to Gofreight")
		})
	})
}

func TestFactories(t *testing.T) {
	gftest.Test(t, "post factory creates valid records", func(t *testing.T) {
		app := newTestApp(t)
		_ = app

		post := factories.PostFactory.Create(t)
		gftest.Expect(post.ID).Bind(t).ToBeGreaterThan(0)
		gftest.Expect(post.Title).Bind(t).ToEqual("Test Post")
	})

	gftest.Test(t, "post factory create many", func(t *testing.T) {
		app := newTestApp(t)
		_ = app

		posts := factories.PostFactory.CreateMany(t, 3)
		gftest.Expect(len(posts)).Bind(t).ToEqual(3)
		gftest.AssertDatabaseCount(t, "posts", 3)
	})
}

func TestExpectations(t *testing.T) {
	gftest.Expect(200).ToEqual(200)
	gftest.Expect("hello world").ToContain("world")
	gftest.Expect([]int{1, 2, 3}).ToHaveCount(3)
	gftest.Expect(true).ToBeTrue()
	gftest.Expect(http.StatusOK).ToBeLessThan(300)
}

func TestFakes(t *testing.T) {
	fakes := gftest.UseFakes()

	gftest.Expect(fakes.Mail.Count()).ToEqual(0)
	gftest.Expect(fakes.Queue.Pending()).ToEqual(0)
	gftest.Expect(fakes.Cache.Has("key")).ToBeFalse()

	fakes.Cache.Put("key", "value")
	gftest.Expect(fakes.Cache.Has("key")).ToBeTrue()
}
