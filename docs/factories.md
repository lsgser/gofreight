# Factories

Gofreight model factories create test (and seed) data with sensible defaults, similar to Laravel [Eloquent Factories](https://laravel.com/docs/eloquent-factories).

---

## Generating factories

```bash
gofreight make:factory Post
gofreight make:scaffold Article title:string body:text  # includes factory
```

Output: `tests/factories/post_factory.go` with faker defaults from `gftest/faker`.

---

## Defining defaults

```go
var PostFactory = gftest.NewFactory(models.Posts).Define(map[string]any{
    "title": faker.Lazy(func() any { return faker.Sentence() }),
    "body":  faker.Lazy(func() any { return faker.Paragraph() }),
})
```

Use `faker.Lazy(fn)` so faker values are evaluated on each `Make()` / `Create()` call.

---

## Instantiating models

**Make** builds an unsaved instance:

```go
post := PostFactory.Make()
post := PostFactory.Make(map[string]any{"title": "Fixed"})
```

**Create** persists via the repository (runs validations):

```go
post := PostFactory.Create(t)
posts := PostFactory.CreateMany(t, 5)
posts := PostFactory.Count(5).Create(t)  // Laravel count(n)->create()
```

---

## Factory states

Apply attribute overrides with **State** (Laravel `state()`):

```go
UnverifiedUser := UserFactory.State(map[string]any{
    "email_verified_at": "",
})

user := UnverifiedUser.Make()
```

Dynamic states with **StateFn**:

```go
admin := UserFactory.StateFn(func(attrs map[string]any) map[string]any {
    attrs["role"] = "admin"
    return attrs
})
```

Generated factories include a named state helper stub:

```go
user := factories.UnverifiedUser().Create(t)
```

---

## Sequences

Increment a field on each build (Laravel sequences):

```go
PostFactory.Sequence("title", func(n int) any {
    return fmt.Sprintf("Post %d", n+1)
})

a := PostFactory.Make() // Post 1
b := PostFactory.Make() // Post 2
```

---

## Callbacks

Run logic after building or persisting:

```go
PostFactory.
    AfterMaking(func(p *models.Post) {
        p.Slug = slugify(p.Title)
    }).
    AfterCreating(func(p *models.Post) {
        // side effects after save
    })
```

---

## Raw create

Skip validations for edge-case tests:

```go
PostFactory.RawCreate(t, map[string]any{"title": ""})
```

---

## Faker helpers

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

Use `faker.DefinitionsForModel("Post")` in generated factories for field-aware defaults.

---

## Related

- [Testing](testing.md) — `gftest`, HTTP tests, database assertions
- [Models](models.md) — struct conventions and repositories
- [Seeding](seeding.md) — seeders and `DatabaseSeeder`
- [Generators](generators.md) — `make:factory` output layout
