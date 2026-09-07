# Admin dashboard

Browser-based database administration for **development and test environments only** — inspect tables, run queries, and browse records without leaving the browser.

> **Warning:** The admin panel is **disabled in production** (`GOFREIGHT_ENV=production`). Never expose it publicly.

---

## Access

Start the dev server:

```bash
gofreight serve
```

Open:

```
http://localhost:5000/admin
```

The URL is printed in the development server banner.

---

## Features

| Feature | Description |
|---------|-------------|
| Table browser | List all tables in the connected database |
| Row viewer | Paginated record browsing |
| Schema inspection | Column names and types |
| SQL runner | Execute read queries (development safety limits apply) |

Works with SQLite, PostgreSQL, and MySQL — whatever database your app connects to.

---

## Enabling / disabling

Admin mounts automatically in development via `app.Run()`:

```go
if app.Config.IsDevelopment() {
    app.MountAdmin()
}
```

Manual mount:

```go
app.MountAdmin() // no-op in production
```

Custom config:

```go
cfg := admin.DefaultConfig(true) // debug mode
panel, err := admin.New(cfg)
panel.Mount(app.Router)
```

---

## Security

- **Production guard:** `MountAdmin()` returns immediately when `GOFREIGHT_ENV=production`
- **Local only:** Intended for localhost development
- **No authentication:** Do not expose on public networks
- **Read-focused:** Destructive operations are limited in development

For production database management, use dedicated tools (pgAdmin, TablePlus, cloud consoles).

---

## Troubleshooting

**Admin not loading:**

- Confirm `GOFREIGHT_ENV=development`
- Check database connection (`gofreight tinker` → `SELECT 1`)
- Look for warnings in server logs: `admin panel failed to load`

**Empty table list:**

- Run migrations: `gofreight migrate`
- Verify `DATABASE_URL` points to the correct database

---

## Related

- [Database](database.md) — migrations and schema
- [CLI Commands](commands.md) — `gofreight tinker` for SQL console
- [Deployment](deployment.md) — production (admin disabled)
