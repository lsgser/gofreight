<!--
| --------------------------------------------------------------------------
| Sessions
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
| Sessions
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Sessions

HTTP session management with pluggable storage drivers, flash messages, and CSRF integration.

## Drivers

| Driver | Env | Storage |
|--------|-----|---------|
| In-memory | default (no config) | Process RAM — lost on restart |
| File | `SESSION_DRIVER=file` | `storage/framework/sessions/` |
| Redis | `SESSION_DRIVER=redis` | Redis keys |

### Configure in bootstrap

```go
// File (default for new apps)
_ = app.UseFileSessions("")  // empty = storage/framework/sessions

// Redis (multi-process production)
_ = app.UseRedisSessions(config.ResolveRedisURL())
```

Or use the convenience helper:

```go
app.WireDefaults() // reads SESSION_DRIVER from env
```

---

## Session API in controllers

Sessions are available on `controller.Base`:

```go
func (c AuthController) Login(base controller.Base) error {
    base.Session.Set("current_user_id", user.ID)
    base.Session.Set("role", user.Role)
    return base.Redirect("/dashboard")
}

func (c AuthController) Logout(base controller.Base) error {
    base.Session.Delete("current_user_id")
    return base.Redirect("/login")
}

func (c DashboardController) Show(base controller.Base) error {
    userID, ok := base.Session.GetInt64("current_user_id")
    if !ok {
        return base.Redirect("/login")
    }
    // ...
}
```

### Common methods

| Method | Purpose |
|--------|---------|
| `Set(key, value)` | Store a value |
| `Get(key)` | Retrieve (any type) |
| `GetInt64(key)` | Retrieve as int64 |
| `GetString(key)` | Retrieve as string |
| `Delete(key)` | Remove a key |
| `Has(key)` | Check existence |
| `Flash(key, value)` | Set flash (one request) |
| `GetFlash(key)` | Read and consume flash |

---

## Flash messages

Flash data survives one redirect — ideal for success/error messages:

```go
// After creating a post:
base.Session.Flash("success", "Post created!")
return base.Redirect("/posts")

// On the next request:
if msg, ok := base.Session.GetFlash("success"); ok {
    data["FlashSuccess"] = msg
}
```

### GFT flash partial

New apps include `app/views/partials/flash.gft`:

```html
#when .FlashSuccess
  <div class="alert alert-success">{= .FlashSuccess }</div>
#endwhen
#when .FlashError
  <div class="alert alert-error">{= .FlashError }</div>
#endwhen
```

Include in layouts with `#partial "partials/flash"`.

---

## CSRF protection

Enable CSRF middleware:

```go
app.UseCSRF()
```

In GFT forms, use the `#token` directive:

```html
#form action="/posts" method="POST"
  #field "title" label="Title"
  #token
  <button type="submit">Save</button>
#endform
```

`#token` renders:

```html
<input type="hidden" name="authenticity_token" value="...">
```

For AJAX, read the token from a meta tag or cookie and send as header.

See [Security](security.md) and [Forms & Validation](forms-validation.md).

---

## Session lifetime

```env
SESSION_LIFETIME=120   # minutes
```

Configured via `config.Config` — sessions expire after inactivity.

---

## Authentication integration

Session login helpers in `auth` package:

```go
auth.Login(auth.DefaultLoginConfig(findUser))(base)
auth.Logout("current_user_id", "/login")(base)
```

Retrieve authenticated user:

```go
userID, ok := auth.UserIDFromRequest(r, "current_user_id")
role, ok := auth.RoleFromRequest(r)
```

See [Authentication](authentication.md).

---

## File session storage

```go
_ = app.UseFileSessions("storage/framework/sessions")
```

Each session is a file on disk — suitable for single-server deployments. Directory is created automatically.

---

## Redis session storage

```go
_ = app.UseRedisSessions("redis://127.0.0.1:6379")
```

Required when running **multiple app processes** (load balancer, multiple `gofreight serve` instances) so sessions are shared.

---

## Testing sessions

```go
func TestLoginSetsSession(t *testing.T) {
    app := gftest.NewApp(t, gftest.WithMigrateDir("db/migrate"))
    app.Draw(routes.Register)

    app.Post("/login", strings.NewReader("email=a@b.com&password=secret")).
        AssertRedirect()

    // Follow-up authenticated request...
}
```

---

## Security notes

- Always use `SESSION_DRIVER=redis` or `file` in production (not in-memory)
- Set a strong `APP_KEY` — sessions are encrypted/signed with it
- Use HTTPS in production so session cookies are secure
- Regenerate session ID on login to prevent fixation

---

## Related

- [Authentication](authentication.md) — login/logout
- [Security](security.md) — CSRF, cookie settings
- [Application wiring](application-wiring.md) — `UseFileSessions`, `UseRedisSessions`
- [Configuration](configuration.md) — `SESSION_DRIVER`
