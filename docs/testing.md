# Testing

Gofreight ships **`gftest`**, a full-featured testing library for HTTP endpoints, database assertions, factories, and BDD-style test organization.

## Running tests

```bash
gofreight test              # from app root — runs go test ./...
go test ./...               # standard Go
go test -v ./tests/...      # verbose, specific package
go test -run TestPosts      # single test
```

Set `GOFREIGHT_ENV=test` automatically via `gftest.NewApp`.

---

## Quick start

```go
package tests

import (
    "testing"
    "myapp/routes"
    "github.com/lsgser/gofreight/gftest"
)

func TestHomePage(t *testing.T) {
    app := gftest.NewApp(t)
    app.Draw(routes.Register)
    app.Get("/").AssertOk().AssertSee("Welcome")
}
```

---

## Creating test apps

### `gftest.NewApp(t, opts...)`

| Option | Purpose |
|--------|---------|
| `WithDatabase(url)` | Database URL (default `sqlite://:memory:`) |
| `WithMigrations(sql...)` | Inline SQL migrations |
| `WithMigrateDir(dir)` | Run migrations from `db/migrate/` |
| `WithViewsRoot(path)` | Custom views directory |
| `WithChdir(dir)` | Change working directory during setup |

```go
app := gftest.NewApp(t,
    gftest.WithDatabase("sqlite://:memory:"),
    gftest.WithMigrateDir("db/migrate"),
    gftest.WithViewsRoot("app/views"),
)
app.Draw(routes.Register)
```

There is **no** `WithRoutes` helper — register routes with `app.Draw(routes.Register)`.

---

## HTTP testing

### Request methods

```go
app.Get("/posts")
app.Post("/posts", strings.NewReader("title=Hello"))
app.PostJSON("/api/posts", map[string]string{"title": "Hello"})
app.PutJSON("/api/posts/1", payload)
app.PatchJSON("/api/posts/1", payload)
app.Delete("/api/posts/1")
app.Request("HEAD", "/posts", nil)
```

### JSON requests

```go
app.PostJSON("/api/posts", map[string]any{
    "title": "Hello",
    "body":  "World",
}).AssertCreated().AssertJsonPath("title", "Hello")
```

### Custom headers

```go
app.Get("/api/me", gftest.JSON()) // Accept: application/json

h := http.Header{"Authorization": {"Bearer token"}}
app.Get("/api/me", h)
```

---

## Response assertions

Chain assertions fluently:

```go
resp := app.Get("/posts/1")

resp.AssertOk()                    // 200
resp.AssertCreated()               // 201
resp.AssertNotFound()              // 404
resp.AssertUnprocessable()         // 422
resp.AssertRedirect()              // 3xx
resp.AssertStatus(204)

resp.AssertSee("Hello World")
resp.AssertDontSee("Error")
resp.AssertHeader("Content-Type", "text/html; charset=utf-8")

resp.AssertJSON(&dest)
resp.AssertJsonPath("id", float64(1)) // JSON numbers decode as float64

resp.Body()    // raw string
resp.Status()  // int
```

---

## Describe / It (BDD style)

```go
func TestPosts(t *testing.T) {
    gftest.Describe(t, "Posts API", func(d *gftest.DescribeContext) {
        var app *gftest.App

        d.BeforeEach(func(t *testing.T) {
            app = gftest.NewApp(t, gftest.WithMigrateDir("db/migrate"))
            app.Draw(routes.Register)
        })

        d.It("lists posts", func(t *testing.T) {
            app.Get("/posts").AssertOk()
        })

        d.It("creates a post", func(t *testing.T) {
            app.PostJSON("/posts", map[string]string{"title": "New"}).AssertCreated()
        })
    })
}
```

### Hook reference

| Hook | Runs |
|------|------|
| `BeforeAll(fn)` | Once before all tests in the group |
| `AfterAll(fn)` | Once after all tests |
| `BeforeEach(fn)` | Before each `It` |
| `AfterEach(fn)` | After each `It` |

### Parameterized tests

```go
datasets := gftest.Datasets(map[string]map[string]any{
    "admin":  {"role": "admin", "expected": 200},
    "guest":  {"role": "guest", "expected": 403},
})

d.ItWith("access control", datasets, func(t *testing.T, ds gftest.Dataset) {
    role := ds.Data["role"].(string)
    code := ds.Data["expected"].(int)
    // ...
})
```

### Single test shorthand

```go
gftest.Test(t, "health check", func(t *testing.T) {
    app := gftest.NewApp(t)
    app.Get("/health").AssertOk()
})
```

---

## Expect assertions

Standalone assertions (also used internally):

```go
gftest.Expect(42).ToEqual(42)
gftest.Expect("hello").ToContain("ell")
gftest.Expect([]int{1,2,3}).ToContain(2)
gftest.Expect(err).ToBeNil()
gftest.Expect(ok).ToBeTrue()
gftest.Expect(value).Not().ToEqual(0)
```

Bind to a test for better failure output:

```go
gftest.Expect(result).Bind(t).ToEqual(expected)
```

---

## Database assertions

```go
app.DB().ToHaveCount("posts", 3)
app.DB().ToHaveRecord("posts", map[string]any{"title": "Hello"})
app.DB().ToHaveRecord("users", map[string]any{"email": "admin@example.com"})
```

Requires an active database connection (default in `NewApp`).

---

## Factories

Generate test data with **`gftest/faker`**:

```bash
gofreight make:factory Post
gofreight make:scaffold Article title:string body:text  # includes factory
```

```go
import "github.com/lsgser/gofreight/gftest/faker"

var PostFactory = gftest.NewFactory(models.Posts).Define(map[string]any{
    "title": faker.Lazy(func() any { return faker.Sentence() }),
    "body":  faker.Lazy(func() any { return faker.Paragraph() }),
})

post := PostFactory.Create(t, map[string]any{"title": "Fixed title"})
```

### Faker helpers

| Function | Example output |
|----------|----------------|
| `faker.Name()` | Random name |
| `faker.Email()` | Random email |
| `faker.Sentence()` | Short sentence |
| `faker.Paragraph()` | Long text |
| `faker.URL()` | Random URL |
| `faker.UUID()` | UUID string |
| `faker.Date()` | Date string |
| `faker.DateTime()` | Datetime string |
| `faker.Int(min, max)` | Random int |
| `faker.Bool()` | true/false |

---

## Fakes (mail, cache, queue)

```go
func TestWelcomeEmail(t *testing.T) {
    gftest.UseFakes()
    app := gftest.NewApp(t)
    app.Draw(routes.Register)

    app.Post("/register", body).AssertRedirect()

    gftest.AssertMailSent(t, 1)
}
```

Swap framework services with in-memory fakes for isolated tests.

---

## Testing with authentication

```go
func TestProtectedRoute(t *testing.T) {
    app := gftest.NewApp(t, gftest.WithMigrateDir("db/migrate"))
    app.Draw(routes.Register)

    h := http.Header{"Authorization": {"Bearer " + testToken}}
    app.Get("/api/me", h).AssertOk()
}
```

For session auth, set session cookies via `httptest` or login through the test HTTP client.

---

## Testing views

```go
app := gftest.NewApp(t, gftest.WithViewsRoot("app/views"))
app.Draw(routes.Register)
app.Get("/login").AssertOk().AssertSee("Sign in")
```

---

## Example: full CRUD test

```go
func TestPostsCRUD(t *testing.T) {
    gftest.Describe(t, "Posts", func(d *gftest.DescribeContext) {
        var app *gftest.App

        d.BeforeEach(func(t *testing.T) {
            app = gftest.NewApp(t,
                gftest.WithMigrations(`CREATE TABLE posts (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    title TEXT NOT NULL,
                    body TEXT,
                    created_at TEXT DEFAULT (datetime('now'))
                )`),
            )
            app.Draw(routes.Register)
        })

        d.It("creates and shows a post", func(t *testing.T) {
            app.PostJSON("/api/posts", map[string]string{
                "title": "Test",
                "body":  "Content",
            }).AssertCreated()

            app.DB().ToHaveCount("posts", 1)
            app.Get("/api/posts/1").AssertOk().AssertJsonPath("title", "Test")
        })
    })
}
```

---

## Tips

- Use `sqlite://:memory:` for fast isolated tests
- Call `app.Draw(routes.Register)` after every `NewApp`
- Use factories instead of hard-coded SQL inserts
- Prefer `PostJSON` for API tests, `Post` with form body for HTML forms
- Run `gofreight test` in CI the same way as locally

## Related

- [Generators](generators.md) — `make:factory`, `make:test`
- [Database](database.md) — migrations in tests
- [Routing](routing.md) — testing named routes
- [Authentication](authentication.md) — testing protected endpoints
