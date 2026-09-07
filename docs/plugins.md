# Plugins & lifecycle hooks

Gofreight provides a lightweight plugin registry for lifecycle hooks — useful for packages, internal modules, or app extensions that need to run code at boot time.

## Registering hooks

```go
import "github.com/lsgser/gofreight/plugins"

func init() {
    plugins.Register("boot", func(args ...any) error {
        log.Println("Plugin booted")
        return nil
    })

    plugins.Register("before_run", func(args ...any) error {
        app := args[0].(*application.Application)
        log.Printf("Starting %s on port %d", app.Config.AppName, app.Config.Port)
        return nil
    })
}
```

## Built-in events

| Event | When fired | Arguments |
|-------|------------|-----------|
| `boot` | Early in `app.Run()`, before views load | `*application.Application` |
| `before_run` | Just before HTTP server starts | `*application.Application` |

The framework calls `plugins.Run(event, app)` internally during startup.

## Listing registered events

```go
for _, name := range plugins.Events() {
    log.Println("hook:", name)
}
```

## Error handling

If any hook returns an error, `plugins.Run` stops and propagates the error. A failing `boot` hook prevents the server from starting.

## Example: register custom health check

```go
// app/plugins/health.go
package plugins

import (
    "context"
    "github.com/lsgser/gofreight/health"
    "github.com/lsgser/gofreight/plugins"
)

func init() {
    plugins.Register("before_run", func(args ...any) error {
        // Custom health checks can be added when mounting health manually
        _ = context.Background()
        return nil
    })
}
```

## Example: third-party package integration

```go
// mypackage/register.go
package mypackage

import "github.com/lsgser/gofreight/plugins"

func init() {
    plugins.Register("boot", func(args ...any) error {
        return setupMetrics()
    })
}
```

Import the package from `main.go` or `bootstrap/app.go`:

```go
import _ "myapp/app/plugins/metrics"
```

## Plugins vs service container

| Use case | Approach |
|----------|----------|
| Run code at startup | `plugins.Register` |
| Resolve services in controllers | Service container (`app.Singleton`) |
| HTTP middleware | `app.Router.Use` |
| Route registration | `routes/` directory |

## Custom events

You can register and fire custom events from your own code:

```go
plugins.Register("post.created", func(args ...any) error {
    postID := args[0].(int64)
    log.Printf("Post %d created", postID)
    return nil
})

// After creating a post:
_ = plugins.Run("post.created", postID)
```

There is no built-in event bus beyond this hook registry — for complex pub/sub, use your own service or Redis.

## Related

- [Extending Gofreight](extending.md) — integrations and custom drivers
- [Application wiring](application-wiring.md) — bootstrap lifecycle
- [Services & Container](services.md) — dependency injection
