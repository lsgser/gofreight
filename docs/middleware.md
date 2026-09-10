<!--
| --------------------------------------------------------------------------
| Middleware
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
| Middleware
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Middleware

Middleware wraps HTTP handlers in a pipeline — each layer can inspect or modify the request before passing it to the next handler. Gofreight ships built-in middleware for sessions, CSRF, CORS, rate limiting, logging, and more.

## How middleware works

Middleware has the signature `func(http.Handler) http.Handler`. Apply it globally on the application, on route groups, or on individual routes.

```go
// Global — every request
app.Router.Use(middleware.Logger)
app.Router.Use(middleware.Recovery)

// Route group
r.Group(func(api *router.Router) {
    API(api)
}).Prefix("/api/v1").Use(auth.JWTMiddleware(jwtMgr)).Apply()

// Single route
r.Get("/admin", adminHandler).Use(customMiddleware)
```

Middleware runs in registration order: first registered runs first on the way in, last on the way out.

## Built-in middleware

| Middleware | Package | Purpose |
|------------|---------|---------|
| `Logger` | `middleware` | Logs method, path, status, duration |
| `Recovery` | `middleware` | Catches panics, returns 500 |
| `StructuredLog` | `middleware` | JSON structured request logging |
| `Sessions` | `middleware` | Cookie-based session storage |
| `CSRF` | `middleware` | Cross-site request forgery protection |
| `CORS` | `middleware` | Cross-origin resource sharing |
| `RateLimit` | `middleware` | Per-IP request throttling |
| `Locale` | `middleware` | Sets locale from header or session |
| `MethodSpoof` | `middleware` | Supports `_method=PUT/DELETE` in HTML forms |
| `SecurityHeaders` | `middleware` | HSTS, X-Frame-Options, etc. (production) |
| `JWTMiddleware` | `auth` | Bearer JWT authentication |
| `APITokenMiddleware` | `auth` | Opaque API token authentication |
| `Guard.Middleware` | `auth` | JWT + API token + session unified auth |
| `RequirePolicy` | `auth` | Authorization policy check |
| `RequireRole` | `auth` | Role-based access control |
| `HTTPCache` | `cache` | Cache-Control headers for static responses |

## Application helpers

`application.Application` wraps common setup:

```go
app := application.New()
app.UseCSRF()
app.UseCORS("https://app.example.com")
app.UseRateLimit(60, time.Minute)
app.UseLocale()
app.UseHTTPCache(time.Hour)
app.UseRedisSessions(os.Getenv("REDIS_URL"))
```

## Sessions

Session middleware loads or creates a session for each request and stores it in the request context:

```go
sessions := middleware.NewSessions(app.Config.AppKey)
app.Router.Use(sessions.Middleware)
```

Access the session in controllers:

```go
session := middleware.SessionFromContext(base.Request.Context())
session.Set("current_user_id", userID)
userID := session.Get("current_user_id")
```

See **[Sessions](sessions.md)** for flash messages, Redis driver, and validation error storage.

## CSRF protection

Enable CSRF for HTML forms:

```go
app.UseCSRF()
```

In templates, include the CSRF field:

In GFT forms, use the `#token` directive:

```html
#form action="/posts" method="POST"
  #field "title" label="Title"
  #token
  <button type="submit">Save</button>
#endform
```

API routes using Bearer tokens typically skip CSRF. Safe methods (GET, HEAD, OPTIONS) are excluded automatically.

## CORS

```go
app.UseCORS("http://localhost:3000", "https://app.example.com")
```

Or configure manually:

```go
app.Router.Use(middleware.CORS([]string{"https://app.example.com"}))
```

## Rate limiting

```go
app.UseRateLimit(120, time.Minute) // 120 requests per IP per minute
```

Returns 429 when the limit is exceeded.

## Method spoofing

HTML forms cannot send PUT or DELETE. Gofreight middleware reads `_method` from the form body:

```html
#form action="/posts/1" method="PUT"
  #token
  #field "title" label="Title"
  <button type="submit">Update</button>
#endform
```

Enabled automatically when sessions are configured.

## JWT and API token middleware

Protect API route groups:

```go
jwtMgr := auth.JWTFromEnv(app.Config.AppKey)

r.Group(func(api *router.Router) {
    api.Get("/profile", profileHandler)
}).Prefix("/api/v1").Use(auth.JWTMiddleware(jwtMgr)).Apply()
```

For opaque API tokens:

```go
store := auth.NewMemoryTokenStore()
r.Use(auth.APITokenMiddleware(store))
```

See **[Authentication](authentication.md)**.

## Custom application middleware

Create middleware in `app/middleware/`:

```go
// app/middleware/request_id.go
package middleware

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "net/http"
)

type ctxKey struct{}

func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := make([]byte, 8)
        rand.Read(id)
        w.Header().Set("X-Request-ID", hex.EncodeToString(id))
        ctx := context.WithValue(r.Context(), ctxKey{}, hex.EncodeToString(id))
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

Register in `bootstrap/app.go`:

```go
app.Router.Use(appmiddleware.RequestID)
```

## Middleware order

Recommended order for a typical web app:

1. `Recovery` — catch panics first
2. `Logger` — log all requests
3. `Sessions` — load session before CSRF
4. `MethodSpoof` — rewrite method before routing
5. `CSRF` — validate tokens on mutating requests
6. `Locale` — set language
7. Route-specific auth middleware

```go
app.Router.Use(middleware.Recovery)
app.Router.Use(middleware.Logger)
// sessions + CSRF configured via app.UseCSRF() which depends on sessions
app.UseCSRF()
app.UseLocale()
```

## Exception handler

Rich panic recovery with HTML/JSON error pages:

```go
app.UseExceptionHandler()
```

Catches panics, logs stack traces, and renders:

- `errors/404.gft` / `errors/500.gft` for HTML clients
- JSON `{ "error": "..." }` when `Accept: application/json`
- Debug details when `APP_DEBUG=true`

Works alongside the built-in `middleware.Recovery`. See [Error handling](error-handling.md).

## File sessions

```go
_ = app.UseFileSessions("") // storage/framework/sessions
```

See [Sessions](sessions.md) and [Application wiring](application-wiring.md).

## Related

- [Sessions](sessions.md) — session API, flash, Redis driver
- [Authentication](authentication.md) — login, JWT, API tokens
- [Authorization](authorization.md) — policies and roles
- [Security](security.md) — production hardening
- [Routing](routing.md) — applying middleware to groups
