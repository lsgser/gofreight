<!--
| --------------------------------------------------------------------------
| Database
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
| Database
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Database

Gofreight uses **Laravel-style blueprint migrations** (Go files in `db/migrate/`), seeders, and an optional raw SQL path for legacy apps. Models and repositories are in **[Models](models.md)**; the full column type catalog is in **[Column Types](column-types.md)**. Query building is in **[ORM](orm.md)**.

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

Migration files are **Go blueprint files** in `db/migrate/` (like Laravel's `Schema::create` migrations):

```bash
gofreight make:migration create_users_table
# → db/migrate/20260102150405_create_users_table.go
```

```go
package migrate

func init() {
    database.RegisterMigration("20260102150405_create_users_table", up..., down...)
}

func up20260102150405CreateUsersTable(ctx context.Context) error {
    return database.SchemaCreate(ctx, "users", func(b *database.Blueprint) {
        b.Id()
        b.String("email").NotNull().Unique()
        b.Timestamps()
    })
}

func down20260102150405CreateUsersTable(ctx context.Context) error {
    return database.SchemaDropIfExists(ctx, "users")
}
```

Scaffolds (`make:model`, `make:scaffold`) generate the same blueprint migration files automatically.

New apps include `tools/migrate/main.go`, which imports `db/migrate` and runs registered migrations when you call `gofreight migrate`.

### CLI commands

```bash
gofreight db:create              # create database (Postgres/MySQL)
gofreight migrate                # run pending migrations
gofreight migrate:rollback       # rollback last batch
gofreight migrate:status         # show migration status
gofreight migrate:fresh          # drop all tables and re-migrate
gofreight migrate:fresh --seed   # fresh + run seeders
gofreight migrate:reset          # rollback all migrations
gofreight migrate:refresh        # rollback all, then migrate up
```

Generate a migration:

```bash
gofreight make:migration create_posts_table
gofreight make:migration add_slug_to_posts_table
```

Migrations are tracked in a `schema_migrations` table.

### Legacy SQL migrations

Older apps without `tools/migrate/` can still use plain SQL in `db/migrate/*.sql` with optional `*_down.sql` rollback files.

## Blueprint DSL

The blueprint DSL powers generated migration files. You can also use it directly:

```go
import "github.com/lsgser/gofreight/database"

// Runtime (inside a migration Up function)
database.SchemaCreate(ctx, "users", func(b *database.Blueprint) {
    b.Id()
    b.String("email").NotNull().Unique()
    b.String("name").NotNull()
    b.Index("name")
    b.Timestamps()
})

// Or generate SQL strings programmatically
up, down := database.CreateTableBlueprint("users", func(b *database.Blueprint) {
    b.String("email").NotNull().Unique()
})
database.WriteMigrationPair("db/migrate", "004_create_users", up, down)
```

Column helpers: fluent `String`, `Text`, `Integer`, `Boolean`, `DateTime`, `BigInteger`, `ForeignId`, `Json`, `Uuid`, `Float`, `Decimal`, `Date`, `Time`, `Timestamp` (chain `.NotNull()`, `.Unique()`, `.Default()`), plus Laravel helpers `Id()`, `Timestamps()`, `TimestampsTz()`, `SoftDeletes()`, `SoftDeletesTz()`, `RememberToken()`, `DropSoftDeletes()`, `DropTimestamps()`, `Index`, and `Unique`/`UniqueIndex`. Use `SchemaRename`, `HasTable`, and `HasColumn` like Laravel's Schema facade.

Named migrations follow [Laravel conventions](https://laravel.com/docs/migrations):

```bash
gofreight make:migration create_users_table
gofreight make:migration add_email_to_users_table
gofreight make:migration add_soft_deletes_to_posts_table
gofreight make:migration add_timestamps_to_posts_table
gofreight make:migration add_remember_token_to_users_table
```

Alter existing tables:

```go
up, down := database.AlterTableBlueprint("posts", func(b *database.Blueprint) {
    b.String("slug").NotNull().Unique()
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

SQL seeds in `db/seeds/` run first, then Go seeders via `cmd/seed/`. See **[Seeding](seeding.md)** for the full Laravel-style guide (`DatabaseSeeder`, `Seeder.Call()`, factories).

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
