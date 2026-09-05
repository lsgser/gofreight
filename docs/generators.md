# Generators & field types

Gofreight generators accept fields as `name:type` pairs:

```bash
gofreight make scaffold Article title:string body:text status:enum:draft,published price:float published:boolean
```

Use **`gofreight make scaffold`**, **`gofreight generate scaffold`**, **`gofreight generate model`**, or **`gofreight generate resource`** — they share the same field type vocabulary.

---

## Field type reference

| CLI type | Aliases | Go type | SQL (SQLite default) | Scaffold form |
|----------|---------|---------|----------------------|---------------|
| `string` | `str` | `string` | `VARCHAR(255)` | text input |
| `text` | — | `string` | `TEXT` | textarea |
| `email` | — | `string` | `VARCHAR(255)` | email input |
| `url` | — | `string` | `VARCHAR(512)` | url input |
| `integer` | `int` | `int` | `INTEGER` | number input |
| `bigint` | — | `int64` | `INTEGER` | number input |
| `float` | `decimal`, `double` | `float64` | `REAL` | number input |
| `boolean` | `bool` | `bool` | `INTEGER` (0/1) | checkbox |
| `datetime` | `timestamp` | `string` | `TEXT` | datetime-local input |
| `date` | — | `string` | `TEXT` | date input |
| `time` | — | `string` | `TEXT` | time input |
| `uuid` | — | `string` | `TEXT` | text input |
| `json` | `jsonb` | `string` | `TEXT` | textarea |
| `enum` | — | `string` | `TEXT` + `CHECK (...)` | `<select>` dropdown |
| `references` | `reference`, `belongs_to` | `int64` | `INTEGER` | number input (foreign key id) |

> **Note:** New apps use **SQLite** by default. Types map cleanly to SQLite; PostgreSQL and MySQL accept the same migration SQL in most cases (`REAL`, `TEXT`, `INTEGER`, `VARCHAR`).

---

## Examples by type

### Strings & text

```bash
gofreight make scaffold Post title:string slug:str body:text summary:text
```

### Numbers

```bash
gofreight make scaffold Product name:string sku:string price:float stock:integer legacy_id:bigint
```

### Booleans

```bash
gofreight make scaffold Post title:string published:boolean featured:bool
```

### Dates & times

Stored as `TEXT` in SQLite (ISO-8601 strings). Use `datetime`, `date`, or `time`:

```bash
gofreight make scaffold Event name:string starts_on:date opens_at:time published_at:datetime
```

### Email & URL

```bash
gofreight make scaffold Contact name:string email:email website:url
```

### JSON

Stored as `TEXT`; validate/parse in application code:

```bash
gofreight make scaffold Setting key:string metadata:json config:jsonb
```

### Enums

Comma-separated allowed values. Generates a `CHECK` constraint and a `<select>` in forms:

```bash
gofreight make scaffold Article title:string status:enum:draft,published,archived
```

Migration column:

```sql
status TEXT NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
```

### Foreign keys (`references`)

Use for `belongs_to` associations. Column name should follow `*_id` convention:

```bash
gofreight make scaffold Comment body:text post_id:references:posts author_id:belongs_to:users
```

Syntax:

- `post_id:references:posts`
- `post_id:reference:posts`
- `user_id:belongs_to:users`

Generates `INTEGER NOT NULL` and a number input in forms. Add explicit `FOREIGN KEY` constraints in migrations manually if required.

### UUID

```bash
gofreight make scaffold ApiToken name:string token:uuid
```

---

## Full scaffold example

```bash
gofreight make scaffold Article \
  title:string \
  slug:string \
  body:text \
  status:enum:draft,published,archived \
  published:boolean \
  views:integer \
  price:float \
  published_at:datetime \
  metadata:json
```

Creates:

| Output | Description |
|--------|-------------|
| `app/models/article.go` | Model + presence validators |
| `app/controllers/article_controller.go` | REST controller |
| `app/views/articles/*.gft` | GFT views (enum → select, text → textarea, etc.) |
| `db/migrate/NNN_create_articles.sql` | SQLite migration |
| `tests/article_test.go` | HTTP tests |
| `tests/factories/article_factory.go` | Test factory |
| Route registration in `config/routes.go` | REST routes |

---

## Syntax rules

| Rule | Example |
|------|---------|
| Field format | `name:type` |
| Enum values | `name:enum:value1,value2,value3` |
| Foreign key | `post_id:references:posts` |
| Unknown type | Falls back to `string` / `VARCHAR(255)` |

Field names are lowercased for database columns (`title` → `db:"title"`) and title-cased for Go struct fields (`Title`).

---

## Generators

| Command | Description |
|---------|-------------|
| `gofreight make scaffold Name fields...` | Full CRUD (recommended) |
| `gofreight generate scaffold Name fields...` | Same as above |
| `gofreight generate resource Name fields...` | Alias for scaffold |
| `gofreight generate model Name fields...` | Model + migration only |
| `gofreight generate controller Name` | Controller only |
| `gofreight generate migration name` | Empty migration stub |
| `gofreight generate auth` | User model + auth scaffolding |

See also [Getting Started](getting-started.md) and [ORM](orm.md).
