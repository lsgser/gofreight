# ORM

Gofreight includes an ActiveRecord/Eloquent-style ORM. See the [main README](../README.md#orm-activerecord--eloquent) for the full API reference.

## Highlights

- Chainable query builder (`Where`, `Order`, `Limit`, `Joins`)
- Associations (`HasMany`, `BelongsTo`, eager loading with `With`)
- Scopes and pagination
- Validations and lifecycle callbacks
- Soft deletes, transactions, dirty tracking
- Multi-database support (PostgreSQL, SQLite, MySQL, MariaDB)

## Example

```go
posts, err := Posts.Query(ctx).
    WhereEq("published", true).
    With("Comments").
    OrderDesc("created_at").
    Paginate(1, 20)

post := &Post{Title: "Hello", Body: "World"}
if err := Posts.Create(ctx, post); err != nil {
    // handle validation errors
}
```

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

## Migrations

Use the CLI for schema versioning:

```bash
gofreight generate migration create_posts
gofreight db:migrate
gofreight db:rollback
gofreight db:status
```

For ad-hoc schema changes in development, use the [Admin Dashboard](admin.md).
