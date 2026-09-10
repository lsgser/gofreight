<!--
| --------------------------------------------------------------------------
| Orm
| --------------------------------------------------------------------------
|
| Official Gofreight documentation consumed by the docs site and linked
| from the CLI welcome page.
|
| Keep examples aligned with the current CLI and version; sync copies to
| gofreight-web when publishing.
|
| Covers concepts, tutorials, and reference material for this topic.
|

-->

<!--
| --------------------------------------------------------------------------
| ORM
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# ORM

Gofreight includes a type-safe ORM for Go structs — chainable queries, associations, validations, lifecycle callbacks, soft deletes, and transactions.

> **New to models?** Start with **[Models](models.md)** for defining structs and repositories. For migration column types, see **[Column Types](column-types.md)**.

For a complete working example, see the [blog app](https://github.com/lsgser/gofreight/tree/main/examples/blog).

## Quick start

### 1. Define a model

Models are plain Go structs with `db` tags. Embed `model.Record` for `id`, timestamps, and optional soft-delete support:

```go
// app/models/post.go
package models

import (
    "context"

    "github.com/lsgser/gofreight/model"
)

type Post struct {
    model.Record
    Title string `db:"title" json:"title"`
    Body  string `db:"body" json:"body"`
}

var Posts = model.NewRepository[Post]("posts")

func (p *Post) Save(ctx context.Context) error {
    return Posts.Save(ctx, p)
}
```

### 2. Create a migration

Migration files are plain SQL in `db/migrate/`:

```sql
-- db/migrate/001_create_posts.sql
CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);
```

Run migrations:

```bash
gofreight migrate
```

### 3. Query and persist

```go
// Chainable query
posts, err := models.Posts.Query(ctx).
    WhereEq("published", true).
    OrderDesc("created_at").
    Limit(20).
    Get()

// Create with validation + callbacks
post := &models.Post{Title: "Hello", Body: "World"}
if err := models.Posts.Create(ctx, post); err != nil {
    // handle validation errors
}
```

## Models & conventions

| Convention | Detail |
|------------|--------|
| `db` tags | Required on every persisted field — maps struct fields to column names |
| `model.Record` | Provides `ID`, `CreatedAt`, `UpdatedAt`, and `DeletedAt *string` |
| Repository | One `var Posts = model.NewRepository[Post]("posts")` per table |
| Table name | Passed explicitly to `NewRepository` — not inferred from the struct name |
| JSON tags | Optional — used by API resources, not the ORM itself |

Fields without a `db` tag are ignored on insert and update. The ORM uses reflection; pointer fields (`*string`) map to nullable columns.

## Query builder

Start every query from a repository:

```go
q := Posts.Query(ctx)
```

### Filtering

```go
Posts.Query(ctx).WhereEq("published", true).Get()
Posts.Query(ctx).Where("views", model.OpGt, 100).Get()
Posts.Query(ctx).WhereIn("id", []any{1, 2, 3}).Get()
Posts.Query(ctx).WhereNotIn("status", []any{"archived"}).Get()
Posts.Query(ctx).WhereNull("deleted_at").Get()
Posts.Query(ctx).WhereNotNull("published_at").Get()
Posts.Query(ctx).WhereBetween("views", 10, 100).Get()
Posts.Query(ctx).WhereRaw("title LIKE ?", "%hello%").Get()
Posts.Query(ctx).WhereEq("a", 1).OrWhere("b", model.OpEq, 2).Get()
```

### Ordering, limiting, joins

```go
Posts.Query(ctx).OrderDesc("created_at").Limit(10).Get()
Posts.Query(ctx).Latest("created_at").Get()          // ORDER BY created_at DESC
Posts.Query(ctx).Join("users", "users.id = posts.user_id").Get()
Posts.Query(ctx).LeftJoin("comments", "comments.post_id = posts.id").Get()
Posts.Query(ctx).GroupBy("user_id").Having("count", model.OpGt, 5).Get()
Posts.Query(ctx).Select("id", "title").Distinct().Get()
```

### Retrieving records

```go
posts, _  := Posts.Query(ctx).Get()
post, _   := Posts.Query(ctx).First()
post, _   := Posts.Query(ctx).Find(1)
post, _   := Posts.Query(ctx).FindBy("slug", "hello")
exists, _ := Posts.Query(ctx).WhereEq("slug", "hello").Exists()
sql, args := Posts.Query(ctx).WhereEq("draft", true).ToSQL()
```

### Aggregates

```go
count, _  := Posts.Query(ctx).Count()
sum, _    := Posts.Query(ctx).Sum("views")
avg, _    := Posts.Query(ctx).Avg("rating")
min, _    := Posts.Query(ctx).Min("price")
max, _    := Posts.Query(ctx).Max("price")
titles, _ := Posts.Query(ctx).PluckStrings("title")
ids, _    := Posts.Query(ctx).PluckInt64s("user_id")
ids, _    := Posts.Query(ctx).IDs()
```

### Batch operations

Bulk updates and deletes **do not** run model callbacks or validations:

```go
Posts.Query(ctx).WhereEq("draft", true).UpdateAll(map[string]any{"published": true})
Posts.Query(ctx).WhereEq("draft", true).DeleteAll()
Posts.Query(ctx).Increment("views")
Posts.Query(ctx).Decrement("stock", 5)
Posts.Query(ctx).Touch() // sets updated_at on matching rows
Posts.Query(ctx).FindEach(100, func(p Post) error { ... })
```

### Find-or-create helpers

```go
post, created, _ := Posts.Query(ctx).FirstOrCreate(map[string]any{"title": "Hello"})
post, isNew, _   := Posts.Query(ctx).FirstOrInit(map[string]any{"title": "Draft"})
post, created, _ := Posts.Query(ctx).FindOrCreate(
    map[string]any{"slug": "hello"},           // search
    map[string]any{"title": "Hello", "body": ""}, // create attrs
)
```

### Locking

```go
Posts.Query(ctx).WhereEq("id", 1).ForUpdate().First() // SELECT ... FOR UPDATE
```

## Repository CRUD

Shorthand methods on the repository itself:

```go
Posts.All(ctx)
Posts.Find(ctx, id)
Posts.FindBy(ctx, "slug", "hello")
Posts.First(ctx)
Posts.Where(ctx, "published", true)
Posts.Count(ctx)
Posts.Paginate(ctx, 1, 20)
Posts.Create(ctx, record)
Posts.Update(ctx, record)
Posts.Save(ctx, record)       // insert or update based on ID
Posts.Delete(ctx, id)
Posts.Destroy(ctx, record)
Posts.Reload(ctx, record)
Posts.CreateRecord(ctx, map[string]any{"title": "Hello"})
Posts.UpdateRecord(ctx, id, map[string]any{"title": "Updated"})
Posts.FindOrCreateBy(ctx, map[string]any{"slug": "hello"})
Posts.Touch(ctx, id)
Posts.Exists(ctx)
Posts.Pluck(ctx, "title")
```

`Save` runs validations and callbacks, then inserts (ID == 0) or updates.

## Scopes

Named, reusable query fragments registered on the repository:

```go
var Posts = model.NewRepository[Post]("posts").
    Scope("published", func(q *model.Query[Post]) *model.Query[Post] {
        return q.WhereEq("published", true)
    }).
    Scope("recent", func(q *model.Query[Post]) *model.Query[Post] {
        return q.Latest()
    })

Posts.Query(ctx).Scope("published").Scope("recent").Get()
```

## Associations

Register associations in an `init()` function. Supported types: `belongs_to`, `has_one`, `has_many`, and `many_to_many`.

```go
func init() {
    Posts.Association(model.HasManyAssociation("Comments", "comments", "post_id"))
    Comments.Association(model.BelongsToAssociation("Post", "posts", "post_id"))
}
```

### Eager loading

Use `With()` to preload associations and avoid N+1 queries:

```go
posts, _ := Posts.Query(ctx).With("Comments").Get()
post, _  := Posts.Query(ctx).With("Comments", "Author").Find(1)
```

### Reading association data

Eager-loaded data is stored on the record and retrieved with `GetAssociation`:

```go
comments, _ := model.GetAssociation(post, "Comments") // []map[string]any
author, _    := model.GetAssociation(comment, "Post")  // map[string]any
```

Association data is returned as maps, not typed structs. Map keys match column names.

### Many-to-many

Requires a pivot table with local and foreign keys:

```go
Posts.Association(model.ManyToManyAssociation(
    "Tags", "tags", "post_tag", "post_id", "tag_id",
))
```

## Validations

Implement `Validators()` on your model. Validations run automatically on `Save()` and `Create()`:

```go
func (p *Post) Validators() []model.Validator {
    return []model.Validator{
        model.Presence("Title"),
        model.Length("Title", 3, 255),
        model.Format("Slug", `^[a-z0-9-]+$`),
        model.Numericality("Rating", 0, 5),
        model.Email("AuthorEmail"),
        model.Uniqueness("Title", "title", Posts, p.ID),
        model.Inclusion("Status", []string{"draft", "published"}),
        model.Accepted("Terms"),
    }
}
```

Validation errors are returned from `Save()` / `Create()` as a structured error you can surface in forms or API responses.

## Callbacks

Implement `Callbacks()` on your model to hook into the persistence lifecycle:

```go
func (p *Post) Callbacks() *model.Callbacks {
    if p.cbs == nil {
        p.cbs = model.NewCallbacks()
        p.cbs.BeforeSave(func(ctx context.Context, record any) error {
            post := record.(*Post)
            post.Slug = slugify(post.Title)
            return nil
        })
    }
    return p.cbs
}
```

Available hooks: `BeforeValidate`, `AfterValidate`, `BeforeSave`, `AfterSave`, `BeforeCreate`, `AfterCreate`, `BeforeUpdate`, `AfterUpdate`, `BeforeDelete`, `AfterDelete`.

## Transactions

Wrap multiple operations in a transaction. Pass the transaction context to nested calls:

```go
err := model.Transaction(ctx, func(txCtx context.Context) error {
    if err := Posts.Create(txCtx, &post); err != nil {
        return err
    }
    return Comments.Create(txCtx, &comment)
})
```

Operations that receive the original `ctx` instead of `txCtx` run outside the transaction.

## Pagination

Full pagination returns metadata (total count, page count) like Laravel's `paginate()`:

```go
page, err := Posts.Query(ctx).WhereEq("published", true).Paginate(1, 20)
page.SetLinks("/api/v1/posts") // first, prev, next, last URLs
// page.Data, page.CurrentPage, page.PerPage, page.Total, page.LastPage, page.Links
```

Simple pagination skips the count query (Laravel `simplePaginate()`):

```go
page, err := Posts.Query(ctx).SimplePaginate(2, 20)
// page.Data, page.HasMorePages
```

For JSON APIs, use `api.PaginatedResponse(baseURL, page, perPage, total, data)`.

## Eloquent collections

Query results can be wrapped in a Laravel-style collection:

```go
col, err := Posts.Query(ctx).WhereEq("published", true).GetCollection()
col.Find(1)
col.Filter(func(p Post) bool { return p.Views > 10 })
col.Pluck("title")
col.ModelKeys()
```

See [Laravel Eloquent Collections](https://laravel.com/docs/eloquent-collections) for the conceptual model — Gofreight's `model.Collection[T]` provides `Find`, `Filter`, `Map`, `Pluck`, `ModelKeys`, `Only`, and `Except`.

## Dirty tracking

Track which fields changed between reads and saves:

```go
d := &model.Dirty{}
d.Snapshot(post)
post.Title = "Changed"
d.Changed()        // true
d.ChangedFields()  // ["title"]
d.Changes()        // map with [original, current] pairs
```

Useful for partial updates or audit logging.

## Soft deletes

Soft delete keeps rows in the database and sets a `deleted_at` timestamp instead of running `DELETE`. Trashed records are excluded from normal queries.

### 1. Add `deleted_at` to your table

```sql
-- db/migrate/NNN_add_deleted_at_to_posts.sql
ALTER TABLE posts ADD COLUMN deleted_at TEXT;
```

For new tables, include a nullable column:

```sql
CREATE TABLE posts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  created_at TEXT DEFAULT (datetime('now')),
  updated_at TEXT DEFAULT (datetime('now')),
  deleted_at TEXT
);
```

`model.Record` includes `DeletedAt *string` with `db:"deleted_at"`.

### 2. Enable on the repository

```go
var Posts = model.NewRepository[Post]("posts").EnableSoftDelete()

// Custom column name (optional):
var Archives = model.NewRepository[Post]("posts").EnableSoftDelete("archived_at")
```

### 3. Usage

```go
// Normal queries skip soft-deleted rows
posts, _ := Posts.Query(ctx).Get()
post, _ := Posts.Find(ctx, id)

// Soft delete (sets deleted_at)
Posts.Destroy(ctx, post)
Posts.Delete(ctx, id)
Posts.Query(ctx).WhereEq("draft", true).DeleteAll()

// Include trashed records
Posts.Query(ctx).WithTrashed().Find(id)
Posts.Query(ctx).OnlyTrashed().Get()

// Restore
Posts.Restore(ctx, post)
Posts.RestoreByID(ctx, id)
Posts.Query(ctx).WithTrashed().WhereEq("id", id).RestoreAll()

// Permanent delete (bypasses soft delete)
Posts.ForceDestroy(ctx, post)
Posts.ForceDelete(ctx, id)
Posts.Query(ctx).WithTrashed().WhereEq("id", id).ForceDelete()

// Check on the model
if post.IsSoftDeleted() { ... }
```

| Method | Behavior |
|--------|----------|
| `EnableSoftDelete()` | Enable soft delete for the repository |
| `Destroy` / `Delete` | Sets `deleted_at` when enabled |
| `Query().Get()` / `Find` | Excludes trashed rows |
| `WithTrashed()` | Include trashed rows |
| `OnlyTrashed()` | Only trashed rows |
| `Restore` / `RestoreByID` | Clears `deleted_at` |
| `ForceDestroy` / `ForceDelete` | Hard `DELETE` |
| `DeleteAll()` | Soft-deletes matching rows when enabled |

## Upsert (PostgreSQL)

Insert or update on conflict — PostgreSQL only:

```go
Posts.Upsert(ctx, &post, []string{"slug"})
```

## Polymorphic associations

For commentable-style relationships, use the polymorphic helpers:

```go
parent, err := model.PolymorphicBelongsTo(ctx, "Post", commentableID)
children, err := model.PolymorphicHasMany(ctx, assoc, "Post", postID)
```

These return map data. Register the association config via `PolymorphicAssoc`.

## Database configuration

Supports **SQLite** (default), **PostgreSQL**, **MySQL**, and **MariaDB**.

Configure with discrete `DB_*` variables (recommended) or `DATABASE_URL`:

```env
# SQLite (default)
DB_CONNECTION=sqlite

# PostgreSQL
DB_CONNECTION=pgsql
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=myapp
DB_USERNAME=postgres
DB_PASSWORD=
DB_SSLMODE=disable
```

Connect in `main.go`:

```go
app.ConnectDatabase() // wraps database.Connect via config.ResolveDatabaseURL()
```

## Migrations

Use the CLI for versioned schema changes:

```bash
gofreight generate migration create_posts
gofreight migrate
gofreight migrate:rollback
gofreight migrate:status
gofreight migrate:fresh --seed
gofreight db:seed
```

Migration files live in `db/migrate/` as plain SQL. Each migration can have a paired `*_down.sql` rollback file.

For ad-hoc schema changes in development, use the [Admin Dashboard](admin.md).

### Blueprint DSL

Programmatic migrations with auto-generated rollback SQL:

```go
up, down := database.CreateTableBlueprint("users", func(b *database.Blueprint) {
    b.String("email").NotNull().Unique()
    b.String("name").NotNull()
    b.Index("name")
})
database.WriteMigrationPair("db/migrate", "004_create_users", up, down)
```

Column helpers: fluent `String`, `Text`, `Integer`, `Boolean`, `DateTime` (chain `.NotNull()`, `.Unique()`, `.Default()`), or functional `StringColumn`, `IntegerColumn`, etc. with `database.ColNotNull()`, `database.ColUnique()`. Use `Index` and `UniqueIndex` for indexes.

For altering existing tables:

```go
up, down := database.AlterTableBlueprint("posts", func(b *database.Blueprint) {
    b.String("slug").NotNull().Unique()
})
database.WriteMigrationPair("db/migrate", "005_add_slug_to_posts", up, down)
```

`CreateTableBlueprint` automatically adds `id`, `created_at`, and `updated_at`.

## Testing

Use `gftest` with an in-memory SQLite database:

```go
app := gftest.NewApp(t,
    gftest.WithDatabase("sqlite://:memory:"),
    gftest.WithMigrations(`
        CREATE TABLE posts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            body TEXT NOT NULL,
            created_at TEXT,
            updated_at TEXT
        );
    `),
)
app.Draw(routes.Register)
```

Or point at your migration directory:

```go
gftest.WithMigrateDir("db/migrate")
```

Database assertions:

```go
app.DB().ToHaveCount("posts", 3)
app.DB().ToHaveRecord("posts", map[string]any{"title": "Hello"})
```

Generate factories with faker defaults:

```bash
gofreight make:factory Post
```

See **[Testing](testing.md)** for HTTP tests, factories, and fakes.

## Cross-database notes

| Feature | SQLite | PostgreSQL | MySQL/MariaDB |
|---------|--------|------------|---------------|
| Default timestamps | `TEXT` with `datetime('now')` | `TIMESTAMP DEFAULT NOW()` | `DATETIME DEFAULT NOW()` |
| Upsert | Not supported | `ON CONFLICT` | Insert only |
| `ForUpdate` | Supported | Supported | Supported |
| Column rename in rollback | Limited | Full | Full |

The ORM uses dialect helpers (`database.Placeholder`, `database.ReturningClause`, `database.NowFunc`) so the same model code works across drivers.

## Generators

Scaffold models, migrations, and factories from the CLI:

```bash
gofreight make:model Post title:string body:text
gofreight generate migration create_posts
gofreight make:factory Post
gofreight make:scaffold Article title:string body:text  # model + migration + factory
```

See **[Generators](generators.md)** for field types and output layout.

## Related

- [Example blog app](https://github.com/lsgser/gofreight/tree/main/examples/blog) — full CRUD with associations and tests
- [Testing](testing.md) — `gftest`, factories, database assertions
- [Admin Dashboard](admin.md) — schema browser and SQL export (development)
- [Getting Started](getting-started.md) — database setup and first migration
