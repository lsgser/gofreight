# Cache

Gofreight provides in-memory and Redis cache stores, HTTP response caching, and fragment caching for expensive view partials.

## In-memory cache

```go
import "github.com/lsgser/gofreight/cache"

store := cache.New()

store.Put("user:1", user, time.Hour)
value, ok := store.Get("user:1")
store.Has("user:1")
store.Forget("user:1")
store.Flush()
```

The application exposes a default cache via `app.Cache`.

## Remember pattern

Compute and store on cache miss:

```go
result := app.Cache.Remember("sidebar:nav", time.Minute, func() any {
    return buildNavigation()
})
```

`Forever` stores without expiration:

```go
app.Cache.Forever("config:features", features)
```

## Redis cache

Configure via integrations:

```env
CACHE_DRIVER=redis
REDIS_URL=redis://localhost:6379
```

Redis cache persists across process restarts and works in multi-server deployments.

## HTTP caching

Add cache headers to responses:

```go
app.UseHTTPCache(time.Hour)
```

Or per-route via `cache.HTTPCache` middleware.

## Fragment caching

Cache expensive rendered partials:

```go
fc := cache.NewFragmentCache(app.Cache)
html := fc.Remember("sidebar", time.Minute, func() string {
    return renderSidebar()
})
```

Useful for navigation, category trees, or other slow-to-render fragments.

## CLI

```bash
gofreight cache:clear
```

## Testing

Use a fresh in-memory store per test:

```go
app.Cache.Flush()
```

With `gftest` fakes:

```go
gftest.UseFakes()
// cache operations use in-memory fake
```

## Related

- [Integrations](integrations.md) — Redis cache driver
- [Configuration](configuration.md) — `CACHE_DRIVER`, `REDIS_URL`
- [CLI Commands](commands.md) — `cache:clear`
