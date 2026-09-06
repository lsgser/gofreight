# Testing

Gofreight ships with `gftest`, a testing library for HTTP endpoints and database-backed features.

## Running tests

```bash
gofreight test              # from app root
go test ./...               # standard Go
cd examples/blog && go test ./...
```

## Describe / It syntax

```go
gftest.Describe("Posts", func() {
    gftest.BeforeEach(func(t *testing.T) {
        // setup
    })

    gftest.It("lists published posts", func(t *testing.T) {
        app := gftest.NewApp(t, gftest.WithDatabase("sqlite://:memory:"))
        app.Get("/posts").AssertOk().AssertSee("Hello")
    })
})
```

## Assertions

```go
gftest.Expect(42).ToEqual(42)
gftest.Expect("hello").ToContain("ell")
gftest.Expect(err).ToBeNil()
```

## HTTP testing

```go
app := gftest.NewApp(t,
    gftest.WithMigrations(migrationSQL),
    gftest.WithRoutes(routes.Register),
)
app.Get("/posts").AssertOk().AssertSee("Title")
app.Post("/posts", body).AssertRedirect("/posts/1")
```

## Factories and fakes

Factories use **`gftest/faker`** for random defaults. Values are generated fresh on each `Create` via lazy attributes:

```go
import (
    "github.com/lsgser/gofreight/gftest/faker"
)

post := factories.CreatePost(t) // random title + body
user := factories.CreateUser(t, map[string]any{"email": "fixed@example.com"})
```

Generate a factory with faker wired in:

```bash
gofreight make:factory Post
gofreight make:scaffold Article title:string body:text  # includes faker factory
```

Manual factory example:

```go
var UserFactory = gftest.NewFactory(models.Users).Define(map[string]any{
    "name":  faker.Lazy(func() any { return faker.Name() }),
    "email": faker.Lazy(func() any { return faker.Email() }),
})
```

Faker helpers: `Name`, `Email`, `Sentence`, `Paragraph`, `URL`, `UUID`, `Date`, `DateTime`, `Int`, `Bool`, and more — see `gftest/faker/`.

Service fakes (mail, cache, queue):

```go
gftest.UseFakes()
gftest.AssertMailSent(t, 1)
```

## Database assertions

```go
app.DB().ToHaveCount("posts", 3)
app.DB().ToHaveRecord("posts", map[string]any{"title": "Hello"})
```

See `gftest/` and `examples/blog/tests/` for complete examples.

Related: [Routing](routing.md) (test routes with `gftest.WithRoutes`), [Generators](generators.md) (factories with faker).
