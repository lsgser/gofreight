# Database

Gofreight uses SQL file migrations, optional blueprint DSL, and seeders. The ORM layer (models, queries) is documented separately in **[ORM](orm.md)**.

## Connecting

Call `app.ConnectDatabase()` in `main.go`:

```go
if err := app.ConnectDatabase(); err != nil {
    log.Printf("database: %v", err)
}
```

Configuration via `.env` — see **[Configuration](configuration.md)**.

Supported drivers: **SQLite** (default), **PostgreSQL**, **MySQL**, **MariaDB**.

## Migrations

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

Optional rollback file:

```sql
-- db/migrate/001_create_posts_down.sql
DROP TABLE IF EXISTS posts;
```

### CLI commands

```bash
gofreight db:create              # create database (Postgres/MySQL)
gofreight migrate                # run pending migrations
gofreight migrate:rollback       # rollback last batch
gofreight migrate:status         # show migration status
gofreight migrate:fresh          # drop all tables and re-migrate
gofreight migrate:fresh --seed   # fresh + run seeders
gofreight migrate:reset          # rollback all, then migrate up
```

Generate a migration:

```bash
gofreight generate migration create_posts
gofreight make:migration create_posts
```

Migrations are tracked in a `schema_migrations` table.

## Blueprint DSL

Programmatic migrations with auto-generated rollback:

```go
import "github.com/lsgser/gofreight/database"

up, down := database.CreateTableBlueprint("comments", func(b *database.Blueprint) {
    b.IntegerColumn("post_id", colNotNull())
    b.StringColumn("body")
    b.Index("post_id")
})
database.WriteMigrationPair("db/migrate", "004_create_comments", up, down)
```

Column helpers: `StringColumn`, `IntegerColumn`, `BooleanColumn`, `DateTimeColumn`, `DropColumn`, `Index`.

Alter existing tables:

```go
up, down := database.AlterTableBlueprint("posts", func(b *database.Blueprint) {
    b.StringColumn("slug")
})
database.WriteMigrationPair("db/migrate", "005_add_slug_to_posts", up, down)
```

`CreateTableBlueprint` automatically adds `id`, `created_at`, and `updated_at`.

## Seeding

### SQL seeds

Place files in `db/seeds/`:

```sql
-- db/seeds/001_posts.sql
INSERT INTO posts (title, body) VALUES ('Hello', 'World');
```

Run:

```bash
gofreight db:seed
```

### Go seeders

Generate a seeder class:

```bash
gofreight make:seeder PostSeeder
```

Seeders live in `db/seeders/` and run via `cmd/seed/`.

## Inline migrations (development)

For quick schema changes during development:

```go
// config/database.go
database.Migrate(`
    CREATE TABLE IF NOT EXISTS posts (...);
`)
```

Prefer `db/migrate/` files for versioned schema in team projects.

## Admin dashboard

In development, manage schema visually at `http://localhost:5000/admin` — browse tables, edit rows, export SQL. See **[Admin Dashboard](admin.md)**.

## Testing

Use in-memory SQLite in tests:

```go
app := gftest.NewApp(t,
    gftest.WithDatabase("sqlite://:memory:"),
    gftest.WithMigrations(migrationSQL),
)
```

Or load from your migration directory:

```go
gftest.WithMigrateDir("db/migrate")
```

## Cross-database notes

| Feature | SQLite | PostgreSQL | MySQL |
|---------|--------|------------|-------|
| Default timestamps | `TEXT` + `datetime('now')` | `TIMESTAMP DEFAULT NOW()` | `DATETIME DEFAULT NOW()` |
| Auto increment | `AUTOINCREMENT` | `SERIAL` | `AUTO_INCREMENT` |
| Upsert | Not supported | `ON CONFLICT` | Insert only |

The ORM uses dialect helpers so the same model code works across drivers.

## Wipe (development)

Drop all user tables:

```go
database.Wipe()
```

Use with caution — development and test only.

## Related

- [ORM](orm.md) — models, queries, associations
- [Configuration](configuration.md) — database env vars
- [Testing](testing.md) — database setup in tests
- [Admin Dashboard](admin.md) — visual schema browser
