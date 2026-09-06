# Sessions

Sessions store per-user state across HTTP requests — login status, flash messages, CSRF tokens, and form validation errors.

## Setup

Sessions are enabled automatically when you call `app.UseCSRF()`. Configure explicitly:

```go
sessions := middleware.NewSessions(app.Config.AppKey)
app.Router.Use(sessions.Middleware)
```

For production with multiple server processes, use Redis:

```go
app.UseRedisSessions(os.Getenv("REDIS_URL"))
```

```env
SESSION_DRIVER=redis
REDIS_URL=redis://localhost:6379
```

## Session API

Access the session from the request context:

```go
session := middleware.SessionFromContext(r.Context())

session.Set("current_user_id", userID)
session.Get("current_user_id")
session.GetString("locale")
session.Delete("current_user_id")
```

In controllers:

```go
session := middleware.SessionFromContext(base.Request.Context())
session.Set("cart_id", cartID)
```

## Flash messages

Flash data is available on the **next** request only — ideal for success notices after redirects:

```go
session := middleware.SessionFromContext(r.Context())
session.Flash("success", "Post created!")
session.Flash("error", "Something went wrong")

// On the next request:
flashes := session.Flashes()
// flashes["success"] == "Post created!"
// flashes map is cleared after reading
```

In GFT templates, render flashed messages in your layout when present.

## Validation errors and old input

When form validation fails, Gofreight flashes errors and old input to the session:

```go
// Automatic via controller.RedirectBackWithErrors or ValidateUsing
session.FlashValidationErrors(map[string][]string{"title": {"is required"}})
session.FlashOldInput(map[string]string{"title": "draft title"})
```

Templates read them via `{{ error "title" }}` and old input helpers. The controller pulls and clears them through `ViewData`.

## CSRF tokens

CSRF tokens live in the session under `_csrf_token`. Templates use:

```html
{{ csrfField }}
```

Or read the token programmatically:

```go
token := middleware.CSRFTokenFromSession(session)
```

See **[Middleware](middleware.md#csrf-protection)**.

## Cookie settings

Default session cookie:

| Setting | Value |
|---------|-------|
| Name | `_gofreight_session` |
| Max age | 24 hours |
| HttpOnly | true |
| SameSite | Lax |

Customize via `middleware.NewSessions`:

```go
sm := middleware.NewSessions(appKey)
sm.CookieName = "_myapp_session"
sm.MaxAge = 7 * 24 * time.Hour
app.Router.Use(sm.Middleware)
```

## Authentication integration

Session login stores the user ID:

```go
session.Set("current_user_id", user.ID)
```

Logout clears it:

```go
session.Delete("current_user_id")
```

See **[Authentication](authentication.md)**.

## Related

- [Middleware](middleware.md) — session and CSRF middleware
- [Forms & Validation](forms-validation.md) — flash errors on failed validation
- [Authentication](authentication.md) — session login
