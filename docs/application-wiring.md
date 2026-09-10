<!--
| --------------------------------------------------------------------------
| Application-wiring
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
| Application wiring
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Application wiring

Every Gofreight app boots through `bootstrap/app.go`. This guide documents every wiring helper on `*application.Application` — what it does, when to call it, and how environment variables connect.

## Bootstrap flow

```
main.go
  └── bootstrap.Application()   ← configure middleware, drivers, container
  └── app.ConnectDatabase()
  └── app.Draw(routes.Register)
  └── app.Run()
```

New apps scaffold with sensible defaults:

```go
func Application() *application.Application {
    app := application.New()

    if config.ResolveSessionDriver() == "redis" {
        _ = app.UseRedisSessions(config.ResolveRedisURL())
    } else if config.ResolveSessionDriver() == "file" {
        _ = app.UseFileSessions("")
    }

    if config.ResolveQueueConnection() == "redis" {
        _ = app.UseRedisQueue(config.ResolveRedisURL())
    }

    if config.ResolveRedisURL() != "" {
        _ = app.UseRedisBroadcast(config.ResolveRedisURL())
    }

    _ = app.ConfigureStorage()
    _ = app.ConfigureIntegrations()
    app.UseExceptionHandler()
    app.UseVite()

    _ = app.LoadLocales("config/locales")
    app.UseLocale()
    app.UseCSRF()

    return app
}
```

## One-call defaults: `WireDefaults()`

For minimal bootstrap or tests:

```go
app := application.New()
app.WireDefaults()
```

`WireDefaults()` configures from environment:

| Env | Action |
|-----|--------|
| `SESSION_DRIVER=file` | `UseFileSessions("")` |
| `SESSION_DRIVER=redis` | `UseRedisSessions(REDIS_URL)` |
| `FILESYSTEM_DISK=local` | `ConfigureStorage()` |
| `CACHE_STORE=file` / `redis` | via `ConfigureIntegrations()` |

## Database

```go
if err := app.ConnectDatabase(); err != nil {
    log.Printf("warning: database not connected: %v", err)
}
```

Uses `DATABASE_URL` or `DB_*` from config. See [Database](database.md).

## Sessions

| Method | Driver | Storage |
|--------|--------|---------|
| `UseFileSessions(dir)` | file | `storage/framework/sessions/` (default) |
| `UseRedisSessions(url, prefix)` | redis | Redis keys |

```go
_ = app.UseFileSessions("")                              // default dir
_ = app.UseRedisSessions("redis://127.0.0.1:6379", "")  // optional prefix
```

See [Sessions](sessions.md).

## Cache & mail integrations

```go
_ = app.ConfigureIntegrations()
```

Wires cache and mailer from env:

- `CACHE_STORE=file` → `cache.NewFileStore`
- `CACHE_STORE=redis` → Redis cache
- `MAIL_DRIVER=smtp` / `sendgrid` → SMTP or SendGrid mailer

See [Integrations](integrations.md) and [Cache](cache.md).

## Local storage

```go
_ = app.ConfigureStorage()
```

When `FILESYSTEM_DISK=local`, sets `app.Storage` to `*storage.LocalDisk` rooted at `STORAGE_LOCAL_ROOT` (default `storage/app`).

See [Storage](storage.md).

## Job queue

```go
_ = app.UseRedisQueue(config.ResolveRedisURL())
app.StartJobs(2)  // start worker with 2 goroutines
```

Default queue is in-process. Redis queue requires named jobs for cross-process workers. See [Jobs & Queues](jobs.md).

## Exception handling

```go
app.UseExceptionHandler()
```

Registers panic recovery that renders `errors/404.gft` / `errors/500.gft` or JSON errors. Works alongside the built-in `middleware.Recovery` — the exception handler runs in the middleware stack order you register it.

See [Error handling](error-handling.md).

## Vite frontend

```go
app.UseVite()
```

- Registers `{{vite "entry"}}` template helper
- Proxies `/@vite`, `/resources/`, `/node_modules/` when `public/hot` exists
- Respects `VITE_DEV_SERVER_URL` (default `http://localhost:5173`)

See [Templating](templating.md).

## WebSocket broadcast

```go
_ = app.UseRedisBroadcast(config.ResolveRedisURL())
```

Multi-instance WebSocket fan-out via Redis pub/sub. See [Real-time](realtime.md).

## Locale / i18n

```go
_ = app.LoadLocales("config/locales")
app.UseLocale()
```

Loads JSON translation files and adds locale detection middleware. See [Localization](localization.md).

## CSRF, CORS, rate limiting

```go
app.UseCSRF()
app.UseCORS("https://app.example.com")
app.UseRateLimit(100, time.Minute)
```

CSRF is enabled by default in new app bootstrap. See [Middleware](middleware.md) and [Security](security.md).

## Real-time channels

```go
app.MountChannels("/socket")   // WebSocket endpoint
app.MountSocket("/socket")     // alias
```

See [Real-time WebSockets](realtime.md).

## GraphQL

```go
err := app.MountGraphQL("/graphql", graphql.Config{ /* ... */ })
// or
app.MountGraphQLApplication("/graphql", gqlApp)
```

See [GraphQL](graphql.md).

## Health check

`app.Run()` automatically mounts `GET /health` with a database ping check. Customize via `health.New()` if needed. See [Deployment](deployment.md).

## Route cache (CLI)

```bash
gofreight route:cache
```

Sets `GOFREIGHT_ROUTE_CACHE=1`, writes `bootstrap/cache/routes.json`, exits. Called automatically in `app.Run()` when env is set.

## Scheduler (CLI)

```bash
gofreight schedule:run
```

Sets `GOFREIGHT_SCHEDULE_RUN=1`. Tasks defined in `bootstrap/schedule.go`. See [Scheduling](scheduling.md).

## Service container

```go
app.Singleton("payment", func() any {
    return services.NewPaymentService()
})

svc := app.Make("payment").(*services.PaymentService)
```

See [Services & Container](services.md).

## Notifier

```go
store := notification.NewMemoryStore()
notifier := app.NewNotifier(store)
_ = notifier.Send(ctx, userID, WelcomeNotification{})
```

See [Notifications](notifications.md).

## Production checklist

```go
func Application() *application.Application {
    app := application.New()
    app.WireDefaults()

    if app.Config.IsProduction() {
        _ = app.LoadAssetManifest("public/manifest.json")
    }

    app.UseExceptionHandler()
    app.UseCSRF()
    app.UseRateLimit(120, time.Minute)

    return app
}
```

Additional production guidance: [Deployment](deployment.md), [Security](security.md).

## Related

- [Configuration](configuration.md) — environment variables
- [Project structure](project-structure.md) — where files live
- [Getting started](getting-started.md) — first app
