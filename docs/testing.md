# Testing

Gofreight ships with `gftest`, a Pest/Jest-inspired testing library.

## Running tests

```bash
gofreight test              # from app root
go test ./...               # standard Go
cd examples/blog && go test ./...
```

## Pest-style syntax

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
    gftest.WithRoutes(config.Routes),
)
app.Get("/posts").AssertOk().AssertSee("Title")
app.Post("/posts", body).AssertRedirect("/posts/1")
```

## Factories and fakes

```go
post := factories.Post(t, map[string]any{"title": "Test"})
gftest.UseFakeMailer()
gftest.UseFakeCache()
```

## Database assertions

```go
app.DB().ToHaveCount("posts", 3)
app.DB().ToHaveRecord("posts", map[string]any{"title": "Hello"})
```

See `gftest/` and `examples/blog/tests/` for complete examples.
