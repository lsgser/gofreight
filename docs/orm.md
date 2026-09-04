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

## Migrations

Use the CLI for schema versioning:

```bash
gofreight generate migration create_posts
gofreight db:migrate
gofreight db:rollback
gofreight db:status
```

For ad-hoc schema changes in development, use the [Admin Dashboard](admin.md).
