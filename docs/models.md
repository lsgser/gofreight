# Models

Models are plain Go structs mapped to database tables. Each model gets a **repository** for queries, validation, and persistence. For the query builder, associations, and advanced ORM features, see **[ORM](orm.md)**. For column and migration types, see **[Column Types](column-types.md)**.

---

## Quick start

### 1. Generate or define a model

```bash
gofreight make:model Post title:string body:text published:boolean
gofreight migrate
```

Or create manually in `app/models/post.go`:

```go
package models

import (
	"context"

	"github.com/lsgser/gofreight/model"
)

type Post struct {
	model.Record
	Title     string `db:"title" json:"title"`
	Body      string `db:"body" json:"body"`
	Published bool   `db:"published" json:"published"`
}

var Posts = model.NewRepository[Post]("posts")

func (p *Post) Save(ctx context.Context) error {
	return Posts.Save(ctx, p)
}
```

### 2. Match the migration

Every `db` tag needs a column. Generator migrations are created automatically; manual SQL example:

```sql
-- db/migrate/002_create_posts.sql
CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    published INTEGER NOT NULL DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);
```

### 3. Create and query

```go
post := &models.Post{Title: "Hello", Body: "World", Published: true}
if err := models.Posts.Create(ctx, post); err != nil {
    // validation or DB error
}

items, err := models.Posts.Query(ctx).
    WhereEq("published", true).
    OrderDesc("created_at").
    Limit(20).
    Get()
```

---

## Conventions

| Topic | Detail |
|-------|--------|
| **`db` tags** | Required on every persisted field — maps struct fields to column names |
| **`model.Record`** | Embeds `ID`, `CreatedAt`, `UpdatedAt`, and optional `DeletedAt` |
| **Repository** | One `var Posts = model.NewRepository[Post]("posts")` per table |
| **Table name** | Passed explicitly to `NewRepository` — not inferred from the struct name |
| **JSON tags** | Optional — used by API resources, not the ORM itself |
| **Nullable fields** | Use pointers (`*string`, `*int64`) for nullable columns |
| **Ignored fields** | Fields without a `db` tag are skipped on insert/update |

---

## `model.Record`

Embed `model.Record` on every model for standard primary key and timestamps:

```go
type Post struct {
    model.Record
    Title string `db:"title"`
}
```

| Field | Column | Type in Go |
|-------|--------|------------|
| `ID` | `id` | `int64` |
| `CreatedAt` | `created_at` | `string` (ISO-8601 text in SQLite) |
| `UpdatedAt` | `updated_at` | `string` |
| `DeletedAt` | `deleted_at` | `*string` (nullable, for soft deletes) |

---

## Repositories

A repository is the entry point for all database operations on a model:

```go
var Posts = model.NewRepository[Post]("posts")
```

### Create, save, update, delete

```go
// Create (runs validations + callbacks)
post := &models.Post{Title: "Draft"}
err := models.Posts.Create(ctx, post)

// Save — insert if ID is zero, otherwise update
err := models.Posts.Save(ctx, post)

// Find by primary key
post, err := models.Posts.Find(ctx, 1)

// Find by column
post, err := models.Posts.FindBy(ctx, "slug", "hello-world")

// Update specific fields
err := models.Posts.Update(ctx, post, map[string]any{"title": "New title"})

// Delete (hard delete unless soft delete is enabled)
err := models.Posts.Delete(ctx, post)
```

### Scopes

Reusable query fragments registered on the repository:

```go
var Posts = model.NewRepository[Post]("posts").Scope("published", func(q *model.Query[Post]) *model.Query[Post] {
    return q.WhereEq("published", true)
})

Posts.Query(ctx).Scope("published").Get()
```

---

## Soft deletes

Add a nullable `deleted_at` column and enable soft delete on the repository:

```sql
ALTER TABLE posts ADD COLUMN deleted_at TEXT;
```

```go
var Posts = model.NewRepository[Post]("posts").EnableSoftDelete()

// Soft delete sets deleted_at instead of removing the row
err := Posts.Delete(ctx, post)

// Include trashed records
Posts.Query(ctx).WithTrashed().Get()

// Only trashed
Posts.Query(ctx).OnlyTrashed().Get()

// Restore
err := Posts.Restore(ctx, post)
```

See **[ORM — Soft deletes](orm.md#soft-deletes)** for full API.

---

## Validations

Attach Vine rules to models for create/update validation:

```go
func (p *Post) Rules() map[string]vine.Rule {
    return map[string]any{
        "title": vine.String().Required().Max(255),
        "body":  vine.String().Required(),
    }
}
```

Validation runs automatically on `Create` and `Save` when rules are defined.

---

## Associations

Register relationships on the repository:

```go
var Posts = model.NewRepository[Post]("posts").Association(model.Association{
    Name:  "User",
    Type:  model.BelongsTo,
    Table: "users",
    ForeignKey: "user_id",
    LocalKey:   "id",
})
```

See **[ORM — Associations](orm.md#associations)** for has-many, belongs-to, and eager loading.

---

## Foreign keys

Generator syntax for foreign key columns:

```bash
gofreight make:model Comment body:text post_id:references:posts
```

Creates an `INTEGER` column and `int64` Go field. Wire the association in code as shown above.

---

## CLI generators

| Command | Output |
|---------|--------|
| `gofreight make:model Post fields...` | Model + migration |
| `gofreight make:scaffold Post fields...` | Model + migration + controller + views + routes |
| `gofreight make:factory Post` | Test factory in `tests/factories/` |

Field types (`title:string`, `price:float`, `published:boolean`, etc.) are documented in **[Column Types](column-types.md)** and **[Generators](generators.md)**.

---

## Project layout

```
app/models/
├── post.go          # Post struct + Posts repository
├── user.go
└── doc.go           # Package docs (generated apps)
```

Models are imported by controllers, services, jobs, and tests — keep business logic in **services**, not model files, for larger apps.

---

## Related

- **[ORM](orm.md)** — query builder, transactions, callbacks, associations
- **[Column Types](column-types.md)** — migration column reference
- **[Database](database.md)** — migrations, seeding, blueprint DSL
- **[Generators](generators.md)** — `make:model`, `make:scaffold` field syntax
