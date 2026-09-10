<!--
| --------------------------------------------------------------------------
| Configuration
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
| Configuration
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Configuration

Gofreight loads configuration from environment variables, `.env` files, and layered YAML config files. Environment variables always take precedence.

## Environment variables

Every new app includes a `.env` file. Key variables:

| Variable | Default | Purpose |
|----------|---------|---------|
| `APP_NAME` | `Gofreight` | Application name |
| `APP_KEY` | — | Encryption/signing key (sessions, JWT, CSRF) |
| `APP_URL` | `http://localhost:5000` | Base URL |
| `APP_DEBUG` | `true` | Debug mode |
| `GOFREIGHT_ENV` | `development` | `development`, `test`, or `production` |
| `PORT` | `5000` | HTTP port |
| `HOST` | `0.0.0.0` | Bind address |
| `DB_CONNECTION` | `sqlite` | `sqlite`, `pgsql`, `mysql`, `mariadb` |
| `DB_HOST` | `127.0.0.1` | Database host |
| `DB_PORT` | — | `5432` (Postgres) or `3306` (MySQL) |
| `DB_DATABASE` | — | Database name |
| `DB_USERNAME` | — | Database user |
| `DB_PASSWORD` | — | Database password |
| `DB_SSLMODE` | `disable` | Postgres SSL mode |
| `DATABASE_URL` | — | Full connection URL (overrides `DB_*`) |
| `SESSION_DRIVER` | `file` | `file`, `redis`, or `memory` |
| `CACHE_STORE` | `file` | `file`, `redis`, or in-memory default |
| `QUEUE_DRIVER` / `QUEUE_CONNECTION` | `sync` | `sync` (in-process), `redis` |
| `FILESYSTEM_DISK` | `local` | `local` filesystem disk (`storage/app`) |
| `STORAGE_LOCAL_ROOT` | `storage/app` | Root path for local disk |
| `VITE_DEV_SERVER_URL` | `http://localhost:5173` | Vite dev server (when `public/hot` exists) |
| `REDIS_URL` | — | Redis for sessions, cache, queue, WebSocket broadcast |
| `JWT_TTL` | `24h` | JWT token lifetime |
| `LOG_LEVEL` | `info` | Log verbosity |
| `MAIL_DRIVER` | `log` | `log`, `smtp`, `sendgrid` |
| `MAIL_FROM` | — | Default sender address |

Run `gofreight key:generate` after scaffolding to set a unique `APP_KEY`.

## YAML config files

Layered YAML in `config/`:

```
config/
├── app.yaml           # Base settings
├── development.yaml   # Development overrides
├── production.yaml    # Production overrides
└── test.yaml          # Test overrides
```

Example `config/app.yaml`:

```yaml
app_name: myapp
port: 5000
log_level: info
```

Environment-specific files merge on top. Keys map to env vars via `ApplyEnv`:

```go
files.ApplyEnv(map[string]string{
    "PORT":    "port",
    "APP_KEY": "app_key",
})
```

## Loading config

`application.New()` calls `config.Load()` automatically:

```go
app := application.New()
app.Config.AppName   // from APP_NAME or YAML
app.Config.Port      // from PORT or YAML
app.Config.AppKey    // from APP_KEY
app.Config.DatabaseURL
```

Inspect resolved config:

```bash
gofreight config:show
gofreight config:show database
```

## Database configuration

Discrete `DB_*` variables (recommended):

```env
DB_CONNECTION=pgsql
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=myapp
DB_USERNAME=postgres
DB_PASSWORD=secret
DB_SSLMODE=disable
```

Or a single URL:

```env
DATABASE_URL=postgres://user:pass@localhost:5432/myapp?sslmode=disable
```

Resolution order: `DATABASE_URL` → `DB_URL` → built from `DB_*`.

See **[ORM](orm.md#database-configuration)** and **[Database](database.md)**.

## Integrations via env

Mail, cache, and local storage drivers are configured through environment variables. See **[Integrations](integrations.md)** and **[Storage](storage.md)**.

```env
MAIL_DRIVER=smtp
MAIL_HOST=smtp.example.com
MAIL_PORT=587

SESSION_DRIVER=file
CACHE_STORE=file
FILESYSTEM_DISK=local

# Optional — enables Redis sessions, cache, queue, and WebSocket broadcast
REDIS_URL=redis://127.0.0.1:6379
```

> **Honest defaults:** New apps scaffold with **file** sessions and cache. Cloud storage (S3) is not a built-in driver yet — use the local disk or wire a custom integration.

## Bootstrap wiring

Use `bootstrap/app.go` to configure the application based on config:

```go
func Application() *application.Application {
    app := application.New()

    if os.Getenv("SESSION_DRIVER") == "redis" {
        _ = app.UseRedisSessions(os.Getenv("REDIS_URL"))
    }
    if os.Getenv("QUEUE_DRIVER") == "redis" {
        _ = app.UseRedisQueue(os.Getenv("REDIS_URL"))
    }

    app.UseCSRF()
    _ = app.ConfigureIntegrations()

    return app
}
```

## Production checklist

- Set `GOFREIGHT_ENV=production`
- Set `APP_DEBUG=false`
- Run `gofreight key:generate` and keep `APP_KEY` secret
- Use PostgreSQL or MySQL instead of SQLite
- Set `SESSION_DRIVER=redis` for multi-process deployments
- Never commit `.env` — use `.env.example` as a template

See **[Deployment](deployment.md)** and **[Security](security.md)**.

## Related

- [Database](database.md) — migrations and seeding
- [Integrations](integrations.md) — mail, storage, cache drivers
- [Getting Started](getting-started.md) — first app setup
