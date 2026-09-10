<!--
| --------------------------------------------------------------------------
| Error-handling
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
| Error handling
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Error handling

Consistent error responses for panics, missing pages, and API failures.

## Two layers of protection

Gofreight applications have two panic recovery mechanisms:

| Layer | Source | Behavior |
|-------|--------|----------|
| Built-in recovery | `middleware.Recovery` | Plain 500 response, always active |
| Exception handler | `app.UseExceptionHandler()` | HTML error views or JSON with optional debug |

Register the exception handler in bootstrap (recommended for production apps):

```go
app.UseExceptionHandler()
```

The exception handler middleware runs in your middleware stack and catches panics **before** they reach the client as raw stack traces.

---

## Enabling the exception handler

```go
func Application() *application.Application {
    app := application.New()
    app.UseExceptionHandler()
    return app
}
```

When views are loaded, the handler renders GFT templates. Otherwise it falls back to `http.Error`.

---

## Custom error views

Create templates:

```
app/views/errors/404.gft
app/views/errors/500.gft
```

Example `errors/404.gft`:

```html
#layout "layouts/application"

<h1>Page not found</h1>
<p>Sorry, we couldn't find {= .Message }.</p>
<a href="/">Go home</a>
```

Example `errors/500.gft`:

```html
#layout "layouts/application"

<h1>Something went wrong</h1>
<p>We're working on it. Please try again later.</p>
```

The handler uses view names `errors/404` and `errors/500` by default.

---

## JSON API errors

When the client sends `Accept: application/json`:

```json
{
  "error": "Internal Server Error"
}
```

With `APP_DEBUG=true`:

```json
{
  "error": "Internal Server Error",
  "debug": "runtime error: index out of range"
}
```

Never expose debug details in production — set `APP_DEBUG=false`.

---

## Expected errors (not panics)

Use controller helpers for normal error paths:

```go
// 404
base.NotFound("Post not found")

// 403
base.Unauthorized("Forbidden")

// 422 validation
base.Unprocessable(map[string][]string{
    "email": {"is required"},
})

// Abort with any status
base.Abort(409, "Conflict")

// HTTP status helpers
base.Created(resource)
base.NoContent()
```

See [Controllers](controllers.md).

---

## Fallback route (404)

Register a fallback for unmatched routes:

```go
r.Fallback(func(w http.ResponseWriter, r *http.Request) {
    http.NotFound(w, r)
})
```

Or render a custom view in the fallback handler.

---

## Logging

Panics are always logged with stack traces:

```
panic: runtime error: invalid memory address
goroutine 42 [running]:
...
```

Configure log output via `LOG_LEVEL` and your deployment logging stack.

---

## Production checklist

- [ ] `APP_DEBUG=false`
- [ ] `app.UseExceptionHandler()` enabled
- [ ] Custom `errors/500.gft` without sensitive details
- [ ] Monitor logs for panic stack traces
- [ ] Use controller status helpers for expected errors (don't panic)

---

## Testing error responses

```go
func TestNotFound(t *testing.T) {
    app := gftest.NewApp(t)
    app.Draw(routes.Register)
    app.Get("/nonexistent").AssertNotFound()
}

func TestValidationError(t *testing.T) {
    app := gftest.NewApp(t)
    app.Draw(routes.Register)
    app.PostJSON("/posts", map[string]string{}).AssertUnprocessable()
}
```

---

## Related

- [Middleware](middleware.md) — HTTP pipeline
- [Controllers](controllers.md) — status helpers
- [Security](security.md) — production hardening
- [Application wiring](application-wiring.md) — `UseExceptionHandler`
