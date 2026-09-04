# Database Admin Dashboard

Gofreight includes a **local-only** database admin panel inspired by phpMyAdmin and Django Admin. It is available at `/admin` when `GOFREIGHT_ENV=development` and is **never mounted in production**.

## Features

### Data management
- Browse all tables with row counts
- Paginated table views
- Create, edit, and delete rows
- Form fields inferred from column types

### Schema management (phpMyAdmin-style)
- **Create Table** — define columns, types, primary keys, defaults
- **Table Structure** — view column definitions
- **Add Column** — alter existing tables
- **Drop Table** — with confirmation (protects `schema_migrations`)
- **Export SQL** — download CREATE TABLE + INSERT statements
- **Import SQL** — run CREATE, ALTER, INSERT, UPDATE, DELETE statements

### Developer tools
- **SQL Console** — read-only SELECT/PRAGMA/EXPLAIN queries
- **Integrations** — view status of configured cloud services

## Usage

```bash
GOFREIGHT_ENV=development go run .
# Open http://localhost:3000/admin
```

## Manual mounting

```go
app := application.New()
app.ConnectDatabase()
app.MountAdmin() // only in development/test
app.Draw(config.Routes)
app.Run()
```

## Supported databases

PostgreSQL, SQLite, MySQL, and MariaDB. DDL is dialect-aware (auto-increment, timestamps, type mapping).

## Security

The admin panel is a **development tool only**. Do not expose it in production. Gofreight refuses to mount admin routes when `GOFREIGHT_ENV=production`.

Import SQL and schema changes run directly against your database — use only on local or staging databases.
