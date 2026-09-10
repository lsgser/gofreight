<!--
| --------------------------------------------------------------------------
| Admin
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
| Admin dashboard
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Admin dashboard

Browser-based database administration for **development and test environments only**. The UI is inspired by phpMyAdmin — sidebar table list, tabbed table views, SQL console, and schema tools.

> **Warning:** The admin panel is **disabled in production** (`GOFREIGHT_ENV=production`). Never expose it publicly.

---

## Access

```bash
gofreight serve
# → http://localhost:5000/admin
```

Optional password protection: set `ADMIN_PASSWORD` in `.env`.

---

## Interface

| Area | Description |
|------|-------------|
| **Top bar** | Server driver, quick links (Databases, New table, Import, SQL, Status) |
| **Left sidebar** | Filterable table list with row counts |
| **Table tabs** | Browse · Structure · SQL · Search · Insert · Export (per table) |

---

## Features

| Feature | Description |
|---------|-------------|
| Table browser | Dashboard + sidebar with row counts |
| Browse rows | Paginated grid, column sort, bulk select/delete |
| Search | Filter rows by column (=, !=, LIKE, >, <, IS NULL) |
| Insert / edit | Full row CRUD with type hints |
| Structure | Columns, indexes, add/rename/drop columns |
| Create table | Visual table builder |
| SQL console | SELECT / PRAGMA / EXPLAIN with query history |
| Import SQL | Multi-statement DDL/DML import |
| Export | Download table as `.sql` or `.csv` |
| Empty table | Truncate all rows |
| Migrations | Save table schema as migration file |
| Integrations | Status page for configured drivers |

Works with SQLite, PostgreSQL, and MySQL.

---

## Enabling / disabling

Admin mounts automatically in development via `app.Run()`:

```go
app.MountAdmin() // no-op in production
```

Custom config:

```go
cfg := admin.DefaultConfig(true)
panel, err := admin.New(cfg)
panel.Mount(app.Router)
```

---

## Security

- **Production guard:** disabled when `GOFREIGHT_ENV=production`
- **Localhost tooling:** not a replacement for pgAdmin in production
- **Optional auth:** `ADMIN_PASSWORD` session login
- **SQL console:** read-only; use Import for writes
- **Destructive actions:** drop/truncate/bulk delete require confirmation

---

## Troubleshooting

**Admin not loading:** confirm `GOFREIGHT_ENV=development` and database connectivity.

**Empty table list:** run `gofreight migrate`.

---

## Related

- [Database](database.md) — migrations and schema
- [CLI Commands](commands.md) — `gofreight tinker`
- [Deployment](deployment.md) — production (admin disabled)
