<!--
| --------------------------------------------------------------------------
| README
| --------------------------------------------------------------------------
|
| Database migration defining schema changes applied in order by gofreight
| migrate.
|
| Pair up/down migrations when altering columns; prefer blueprint Go
| migrations for new projects.
|

-->

<!--
| --------------------------------------------------------------------------
| Database migrations
| --------------------------------------------------------------------------
|
| Database migration or schema SQL for versioned database changes.
|

-->

# Database migrations

Blueprint migrations live in this directory as Go files (Laravel-style). Run them with:

```bash
gofreight migrate
gofreight migrate:status
```

Create a new migration:

```bash
gofreight make:migration create_posts_table
```

Example migration file:

```go
database.SchemaCreate(ctx, "posts", func(b *database.Blueprint) {
    b.Id()
    b.String("title").NotNull()
    b.Text("body").NotNull()
    b.Timestamps()
})
```

See the [Database docs](https://lsgser.github.io/gofreight-web/docs/database) for the full blueprint reference.
