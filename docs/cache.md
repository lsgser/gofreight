# Cache

Gofreight provides multiple cache backends and HTTP caching middleware for faster applications.

## Drivers

| Driver | Env | Storage | Best for |
|--------|-----|---------|----------|
| In-memory | default | Process RAM | Development, single process |
| File | `CACHE_STORE=file` | `storage/framework/cache/data/` | Single-server production |
| Redis | `CACHE_STORE=redis` | Redis | Multi-process / distributed |

Configure in `.env`:

```env
CACHE_STORE=file
REDIS_URL=redis://127.0.0.1:6379
```

Bootstrap:

```go
_ = app.ConfigureIntegrations() // wires cache from env
```

Or manually:

```go
store, _ := cache.NewFileStore("")
app.Cache = store
```

---

## Basic operations

All stores implement the `cache.Cacher` interface:

```go
// Set (optional TTL)
app.Cache.Put("posts:featured", posts, time.Hour)

// Get
if val, ok := app.Cache.Get("posts:featured"); ok {
    posts := val.([]models.Post)
}

// Check existence
if app.Cache.Has("posts:featured") { /* ... */ }

// Delete
app.Cache.Forget("posts:featured")

// Clear all
app.Cache.Flush()

// Compute-if-missing
posts := app.Cache.Remember("posts:all", time.Minute, func() any {
    return loadAllPosts()
})
```

---

## File store

Persists cache entries as JSON files:

```go
store, err := cache.NewFileStore("") // default: storage/framework/cache/data/
app.Cache = store
```

Features:

- TTL via expiration timestamp in file metadata
- Automatic cleanup on expired reads
- Safe for single-server deployments
- Survives process restarts

---

## Redis store

```go
store, err := cache.NewRedis("redis://127.0.0.1:6379", "gofreight:")
app.Cache = store
```

Requires `REDIS_URL` and `CACHE_STORE=redis` in production multi-process setups.

---

## HTTP cache middleware

Add cache headers to responses:

```go
app.UseHTTPCache(5 * time.Minute)
```

Sets `Cache-Control: public, max-age=300` on responses passing through the middleware.

Use for static-ish API responses or public pages — not for personalized content.

---

## Route-level caching pattern

```go
func (c PostsController) Index(base controller.Base) error {
    if cached, ok := base.App.Cache.Get("posts:index"); ok {
        return base.JSON(cached)
    }

    posts, _ := models.Posts.All(base.Request.Context())
    base.App.Cache.Put("posts:index", posts, 30*time.Second)
    return base.JSON(posts)
}
```

Invalidate on writes:

```go
func (c PostsController) Store(base controller.Base) error {
    // ... create post ...
    base.App.Cache.Forget("posts:index")
    return base.Created(post)
}
```

---

## Fragment caching in views

Pass cached HTML from controllers:

```go
sidebar := app.Cache.Remember("sidebar:nav", time.Hour, func() any {
    html, _ := app.Views.RenderString("partials/sidebar", data)
    return html
})
return base.RenderView("home/index", base.ViewData(map[string]any{
    "Sidebar": sidebar,
}))
```

---

## CLI commands

```bash
gofreight cache:clear    # flush application cache
```

Requires confirmation in production unless `--force`.

---

## Configuration reference

| Variable | Default | Purpose |
|----------|---------|---------|
| `CACHE_STORE` | `file` (new apps) | `file`, `redis`, or memory |
| `REDIS_URL` | — | Required for Redis cache |
| `CACHE_PREFIX` | `gofreight:` | Redis key prefix |

---

## Testing

```go
func TestCachedResponse(t *testing.T) {
    gftest.UseFakes()
    app := gftest.NewApp(t)
    app.Draw(routes.Register)

    app.Get("/posts").AssertOk()
    app.Get("/posts").AssertOk() // second hit uses cache
}
```

Or test cache store directly:

```go
store, _ := cache.NewFileStore(t.TempDir())
store.Put("key", "value", time.Minute)
val, ok := store.Get("key")
```

---

## Driver selection guide

| Deployment | Recommended |
|------------|-------------|
| Local development | In-memory or file |
| Single VPS | File |
| Multiple app instances | Redis |
| Serverless / ephemeral | Redis |

---

## Related

- [Configuration](configuration.md) — `CACHE_STORE`
- [Integrations](integrations.md) — wiring cache drivers
- [Application wiring](application-wiring.md) — `ConfigureIntegrations`
- [Storage](storage.md) — file paths for file cache
