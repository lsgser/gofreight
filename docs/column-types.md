<!--
| --------------------------------------------------------------------------
| Column-types
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
| Column Types
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Column Types

Reference for every column type supported by Gofreight migrations, generators, and the blueprint DSL. Gofreight uses **SQL-first migrations** — you can always write raw SQL for types not covered by generators or blueprints.

Three ways to define columns:

| Method | Best for |
|--------|----------|
| **CLI generators** | `gofreight make:model` / `make:scaffold` — fastest for standard apps |
| **Blueprint DSL** | Programmatic Go migrations with auto rollback |
| **Raw SQL** | Full control — any type your database supports |

Supported drivers: **SQLite** (default), **PostgreSQL**, **MySQL**, **MariaDB**.

---

## CLI generator types

Use `name:type` pairs with `make:model`, `make:scaffold`, `make:api`, and `make:graphql-module`:

```bash
gofreight make:scaffold Product \
  name:string sku:str description:text price:float stock:integer \
  published:boolean metadata:json published_at:datetime
```

| CLI type | Aliases | Go type | SQL (SQLite) | SQL (Postgres) | SQL (MySQL) |
|----------|---------|---------|--------------|----------------|-------------|
| `string` | `str` | `string` | `VARCHAR(255) NOT NULL` | `VARCHAR(255) NOT NULL` | `VARCHAR(255) NOT NULL` |
| `text` | — | `string` | `TEXT NOT NULL` | `TEXT NOT NULL` | `TEXT NOT NULL` |
| `email` | — | `string` | `VARCHAR(255) NOT NULL` | `VARCHAR(255) NOT NULL` | `VARCHAR(255) NOT NULL` |
| `url` | — | `string` | `VARCHAR(512) NOT NULL` | `VARCHAR(512) NOT NULL` | `VARCHAR(512) NOT NULL` |
| `integer` | `int` | `int` | `INTEGER NOT NULL` | `INTEGER NOT NULL` | `INTEGER NOT NULL` |
| `bigint` | — | `int64` | `INTEGER NOT NULL` | `BIGINT NOT NULL` | `BIGINT NOT NULL` |
| `float` | `decimal`, `double` | `float64` | `REAL NOT NULL` | `DOUBLE PRECISION NOT NULL` | `DOUBLE NOT NULL` |
| `boolean` | `bool` | `bool` | `INTEGER NOT NULL DEFAULT 0` | `BOOLEAN NOT NULL DEFAULT false` | `TINYINT(1) NOT NULL DEFAULT 0` |
| `datetime` | `timestamp` | `string` | `TEXT` | `TIMESTAMP` | `TIMESTAMP` |
| `date` | — | `string` | `TEXT` | `DATE` | `DATE` |
| `time` | — | `string` | `TEXT` | `TIME` | `TIME` |
| `uuid` | — | `string` | `TEXT NOT NULL` | `UUID` | `CHAR(36) NOT NULL` |
| `json` | `jsonb` | `string` | `TEXT NOT NULL` | `JSONB NOT NULL` | `JSON NOT NULL` |
| `enum` | — | `string` | `TEXT NOT NULL CHECK (...)` | `TEXT NOT NULL CHECK (...)` | `ENUM(...)` or `VARCHAR` + CHECK |
| `references` | `reference`, `belongs_to` | `int64` | `INTEGER NOT NULL` | `BIGINT NOT NULL` | `BIGINT UNSIGNED NOT NULL` |

### Enum columns

```bash
gofreight make:scaffold Article title:string status:enum:draft,published,archived
```

Generates a `CHECK` constraint on SQLite/Postgres:

```sql
status TEXT NOT NULL CHECK (status IN ('draft', 'published', 'archived'))
```

### Foreign key columns

```bash
gofreight make:scaffold Comment body:text post_id:references:posts
```

Creates `post_id INTEGER NOT NULL` and a Go `int64` field. Add association wiring in the model — see **[Models](models.md#foreign-keys)**.

### Unique columns

Append `:unique` to any field type (Laravel-style):

```bash
gofreight make:scaffold User email:email:unique username:string:unique
gofreight make:scaffold Article slug:string:unique status:enum:draft,published:unique
```

Generates `UNIQUE` on the column in the migration:

```sql
email VARCHAR(255) NOT NULL UNIQUE,
slug VARCHAR(255) NOT NULL UNIQUE
```

---

## Blueprint DSL

Programmatic migrations in Go (`database.Blueprint`). `CreateTableBlueprint` automatically adds `id`, `created_at`, and `updated_at`.

### Fluent API (Laravel-style)

Chain column modifiers like Laravel's schema builder:

```go
import "github.com/lsgser/gofreight/database"

up, down := database.CreateTableBlueprint("users", func(b *database.Blueprint) {
    b.String("email").NotNull().Unique()
    b.String("name").NotNull()
    b.Boolean("active").Default("1")
})
database.WriteMigrationPair("db/migrate", "002_create_users", up, down)
```

Generates:

```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    active BOOLEAN DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);
```

### Functional options

Use `StringColumn`, `IntegerColumn`, etc. with exported column options:

```go
up, down := database.CreateTableBlueprint("comments", func(b *database.Blueprint) {
    b.IntegerColumn("post_id", database.ColNotNull())
    b.StringColumn("body", database.ColNotNull())
    b.StringColumn("slug", database.ColNotNull(), database.ColUnique())
    b.BooleanColumn("approved", database.ColDefault("0"))
    b.DateTimeColumn("reviewed_at", database.ColNullable())
    b.Index("post_id")
    b.UniqueIndex("slug") // composite unique index on multiple columns
})
database.WriteMigrationPair("db/migrate", "004_create_comments", up, down)
```

| Blueprint method | SQL type | Notes |
|------------------|----------|-------|
| `String(name)` / `StringColumn(name, opts...)` | `TEXT` | Chain `.NotNull()`, `.Unique()`, `.Default()` |
| `Text(name)` | `TEXT` | Long text |
| `Integer(name)` / `IntegerColumn(name, opts...)` | `INTEGER` | Integers, foreign keys |
| `Boolean(name)` / `BooleanColumn(name, opts...)` | `BOOLEAN` | Normalized per driver |
| `DateTime(name)` / `DateTimeColumn(name, opts...)` | `TEXT` / `TIMESTAMP` | Datetimes stored as text on SQLite |
| `Id()` | `INTEGER PRIMARY KEY` | Auto-increment id |
| `Timestamps()` | `created_at`, `updated_at` | Laravel `$table->timestamps()` |
| `SoftDeletes()` | `deleted_at` | Nullable timestamp — Laravel `$table->softDeletes()` |
| `SoftDeletesTz()` | `deleted_at` | Nullable timestamptz on Postgres — Laravel `$table->softDeletesTz()` |
| `DropColumn(name)` | — | Alter-table rollback helper |
| `Index(columns...)` | — | Creates `CREATE INDEX IF NOT EXISTS ...` |
| `UniqueIndex(columns...)` | — | Creates `CREATE UNIQUE INDEX IF NOT EXISTS ...` |

### Column options

| Option | Effect |
|--------|--------|
| `.NotNull()` / `database.ColNotNull()` | `NOT NULL` |
| `.Nullable()` / `database.ColNullable()` | nullable (default for blueprint columns) |
| `.Default(v)` / `database.ColDefault(v)` | `DEFAULT v` |
| `.Unique()` / `database.ColUnique()` | `UNIQUE` on the column |
| `b.Timestamps()` | `created_at`, `updated_at` |
| `b.SoftDeletes()` | nullable `deleted_at` (Laravel `$table->softDeletes()`) |
| `b.SoftDeletesTz()` | nullable timezone-aware `deleted_at` (Laravel `$table->softDeletesTz()`) |

### Alter table

```go
up, down := database.AlterTableBlueprint("posts", func(b *database.Blueprint) {
    b.String("slug").NotNull().Unique()
    b.IntegerColumn("view_count", database.ColDefault("0"))
})
```

---

## Raw SQL migrations

For types beyond generators and blueprints, write SQL in `db/migrate/`:

```bash
gofreight make:migration create_products_table
```

Edit the generated file:

```sql
-- db/migrate/20260102120000_create_products_table.sql
CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sku VARCHAR(64) NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    price REAL NOT NULL DEFAULT 0,
    stock INTEGER NOT NULL DEFAULT 0,
    metadata TEXT,
    published_at TEXT,
    deleted_at TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);
```

---

## Column type catalog

Organized like a schema builder reference. **Supported** = built into CLI, blueprint, or documented SQL. **SQL only** = use raw migrations.

### Boolean types

| Type | Support | CLI / Blueprint | Example migration (SQLite) |
|------|---------|-----------------|----------------------------|
| `boolean` | Supported | `published:boolean`, `BooleanColumn` | `published INTEGER NOT NULL DEFAULT 0` |

```bash
gofreight make:model Flag name:string enabled:boolean
```

```go
b.BooleanColumn("enabled", colDefault("1"))
```

---

### String & text types

| Type | Support | CLI / Blueprint | Example |
|------|---------|-----------------|---------|
| `string` | Supported | `title:string` | `title VARCHAR(255) NOT NULL` |
| `text` | Supported | `body:text` | `body TEXT NOT NULL` |
| `char(n)` | SQL only | — | `code CHAR(8) NOT NULL` |
| `varchar(n)` | SQL only | — | `slug VARCHAR(128) NOT NULL` |
| `tinyText` | SQL only | Use `text` CLI type | `notes TEXT` |
| `mediumText` | SQL only | Use `text` CLI type | `content TEXT` |
| `longText` | SQL only | Use `text` CLI type | `content TEXT` |
| `email` | Supported | `contact:email` | `email VARCHAR(255) NOT NULL` |
| `url` | Supported | `website:url` | `website VARCHAR(512) NOT NULL` |

```bash
gofreight make:scaffold Page title:string slug:str content:text excerpt:text
```

```sql
-- Raw SQL: fixed-width char
ALTER TABLE users ADD COLUMN country_code CHAR(2) NOT NULL DEFAULT 'US';
```

---

### Numeric types

| Type | Support | CLI / Blueprint | Example |
|------|---------|-----------------|---------|
| `id` | Auto | Added by `CreateTableBlueprint` | `id INTEGER PRIMARY KEY AUTOINCREMENT` |
| `increments` | Auto | Same as `id` with auto-increment | See above |
| `integer` | Supported | `views:integer`, `IntegerColumn` | `views INTEGER NOT NULL DEFAULT 0` |
| `bigint` | Supported | `legacy_id:bigint` | `legacy_id INTEGER NOT NULL` |
| `smallInteger` | SQL only | — | `sort_order SMALLINT NOT NULL DEFAULT 0` |
| `tinyInteger` | SQL only | — | `priority TINYINT NOT NULL DEFAULT 0` |
| `unsignedInteger` | SQL only | — | `quantity INTEGER UNSIGNED NOT NULL` (MySQL) |
| `float` / `double` | Supported | `price:float`, `rate:double` | `price REAL NOT NULL` |
| `decimal(p,s)` | SQL only | Use `float` CLI or raw SQL | `amount DECIMAL(10,2) NOT NULL` |

```bash
gofreight make:scaffold Order total:float quantity:integer discount:decimal
```

```sql
-- Postgres precise decimal
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    amount DECIMAL(12, 2) NOT NULL,
    tax_rate DOUBLE PRECISION NOT NULL DEFAULT 0
);
```

---

### Date & time types

| Type | Support | CLI / Blueprint | Example |
|------|---------|-----------------|---------|
| `date` | Supported | `starts_on:date` | `starts_on TEXT` (SQLite) / `DATE` (Postgres) |
| `time` | Supported | `opens_at:time` | `opens_at TEXT` / `TIME` |
| `datetime` | Supported | `published_at:datetime` | `published_at TEXT` / `TIMESTAMP` |
| `timestamp` | Supported | Alias for `datetime` | Same as datetime |
| `timestamps` | Auto | `b.Timestamps()` | `created_at`, `updated_at` |
| `softDeletes` | Supported | `b.SoftDeletes()` | `deleted_at TEXT` (SQLite) / `TIMESTAMP` (Postgres/MySQL) |
| `softDeletesTz` | Supported | `b.SoftDeletesTz()` | `deleted_at TIMESTAMPTZ` (Postgres) |
| `year` | SQL only | — | `birth_year INTEGER CHECK (birth_year >= 1900)` |

```bash
gofreight make:scaffold Event name:string starts_on:date opens_at:time published_at:datetime
```

**Soft deletes** — add the column in a migration, then enable on the repository:

```go
database.SchemaCreate(ctx, "posts", func(b *database.Blueprint) {
    b.Id()
    b.String("title").NotNull()
    b.SoftDeletes()    // or b.SoftDeletesTz() for timezone-aware deleted_at
    b.Timestamps()
})
```

```go
var Posts = model.NewRepository[Post]("posts").EnableSoftDelete()
```

See **[Models — Soft deletes](models.md#soft-deletes)**.

---

### Binary types

| Type | Support | Example (SQLite) |
|------|---------|------------------|
| `binary` / `blob` | SQL only | `avatar BLOB` |

```sql
CREATE TABLE attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT NOT NULL,
    data BLOB NOT NULL
);
```

Store large files in **[Storage](storage.md)** when possible; use `BLOB` only for small binary payloads.

---

### JSON types

| Type | Support | CLI | Example |
|------|---------|-----|---------|
| `json` | Supported | `metadata:json` | `metadata TEXT NOT NULL` (SQLite) |
| `jsonb` | Supported | `settings:jsonb` | `settings JSONB NOT NULL` (Postgres) |

```bash
gofreight make:model Setting key:string value:json
```

Parse and validate JSON in application code or Vine rules.

---

### UUID types

| Type | Support | CLI | Example |
|------|---------|-----|---------|
| `uuid` | Supported | `token:uuid` | `token TEXT NOT NULL` (SQLite) / `UUID NOT NULL` (Postgres) |

```bash
gofreight make:model Session token:uuid user_id:references:users
```

**ULID** — SQL only; store as `TEXT` or `CHAR(26)`:

```sql
public_id TEXT NOT NULL UNIQUE  -- store ULID string from Go
```

---

### Relationship types

| Type | Support | CLI | Example |
|------|---------|-----|---------|
| `foreignId` / `references` | Supported | `user_id:references:users` | `user_id INTEGER NOT NULL` |
| `foreignUuid` | SQL only | — | `user_uuid TEXT NOT NULL REFERENCES users(uuid)` |
| `morphs` | SQL only | — | `commentable_type TEXT`, `commentable_id INTEGER` |

```bash
gofreight make:scaffold Comment body:text post_id:references:posts
```

Polymorphic relations (manual SQL):

```sql
CREATE TABLE comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    body TEXT NOT NULL,
    commentable_type TEXT NOT NULL,
    commentable_id INTEGER NOT NULL
);
CREATE INDEX idx_comments_morph ON comments (commentable_type, commentable_id);
```

---

### Enum & set types

| Type | Support | CLI | Example |
|------|---------|-----|---------|
| `enum` | Supported | `status:enum:draft,published` | `TEXT NOT NULL CHECK (...)` |
| `set` | SQL only | — | MySQL `SET('a','b')` or use junction table |

```bash
gofreight make:scaffold Task title:string status:enum:todo,doing,done priority:enum:low,medium,high
```

---

### Specialty types

| Type | Support | Notes |
|------|---------|-------|
| `macAddress` | SQL only | `mac TEXT NOT NULL` — validate in app |
| `ipAddress` | SQL only | `ip TEXT NOT NULL` — use `VARCHAR(45)` for IPv6 |
| `rememberToken` | SQL only | `remember_token TEXT` on users table |
| `vector` | SQL only | Postgres `vector` extension — raw SQL only |

```sql
-- Auth remember token (common pattern)
ALTER TABLE users ADD COLUMN remember_token TEXT;
CREATE INDEX idx_users_remember_token ON users (remember_token);
```

---

### Spatial types

Not built into generators or blueprints. Use database-native types in raw SQL:

```sql
-- PostgreSQL + PostGIS
CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    coords GEOGRAPHY(POINT, 4326)
);

-- SQLite — store lat/lng as REAL columns
CREATE TABLE stores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL
);
```

---

## Admin UI column types

The development admin panel (`/admin`) exposes these types when creating tables or columns:

`INTEGER`, `BIGINT`, `VARCHAR(255)`, `TEXT`, `BOOLEAN`, `TIMESTAMP`, `DATE`, `FLOAT`, `JSON`, `UUID`

See **[Admin Dashboard](admin.md)**.

---

## Driver normalization

Gofreight normalizes some types per driver when using the schema API:

| Input | SQLite | Postgres / MySQL |
|-------|--------|-------------------|
| `VARCHAR`, `CHAR` | `TEXT` | unchanged |
| `BOOLEAN` | `INTEGER` | `BOOLEAN` / `TINYINT(1)` |
| `JSON`, `JSONB` | `TEXT` | `JSON` / `JSONB` |
| `UUID` | `TEXT` | `UUID` / `CHAR(36)` |
| `TIMESTAMP` | `TEXT` | `TIMESTAMP` |

Primary keys use `INTEGER PRIMARY KEY AUTOINCREMENT` on SQLite and `SERIAL` / `AUTO_INCREMENT` on other drivers via `database.AutoIncrement()`.

---

## Complete migration examples

### SQLite (default)

```sql
CREATE TABLE articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    body TEXT NOT NULL,
    excerpt TEXT,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    view_count INTEGER NOT NULL DEFAULT 0,
    featured INTEGER NOT NULL DEFAULT 0,
    metadata TEXT,
    published_at TEXT,
    deleted_at TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX idx_articles_user_id ON articles (user_id);
CREATE INDEX idx_articles_status ON articles (status);
```

### PostgreSQL

```sql
CREATE TABLE articles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    body TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    view_count INTEGER NOT NULL DEFAULT 0,
    featured BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB NOT NULL DEFAULT '{}',
    published_at TIMESTAMP,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### MySQL

```sql
CREATE TABLE articles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    body TEXT NOT NULL,
    status ENUM('draft', 'published', 'archived') NOT NULL DEFAULT 'draft',
    view_count INT UNSIGNED NOT NULL DEFAULT 0,
    featured TINYINT(1) NOT NULL DEFAULT 0,
    metadata JSON NOT NULL,
    published_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_articles_user_id (user_id)
);
```

---

## Related

- **[Models](models.md)** — define structs and repositories
- **[Database](database.md)** — migrations, seeding, blueprint overview
- **[Generators](generators.md)** — CLI field syntax and examples
- **[ORM](orm.md)** — query builder and associations
