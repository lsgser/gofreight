# Security

## Production defaults

- Admin panel (`/admin`) is **never mounted** when `GOFREIGHT_ENV=production`
- Set `ADMIN_PASSWORD` in development if your dev server is network-accessible
- Run `gofreight key:generate` and keep `APP_KEY` secret — it signs sessions, CSRF tokens, and encrypted cookies

## Authentication

Session login (HTML), JWT (API), and opaque API tokens are all supported.

```go
import "github.com/lsgser/gofreight/auth"

hash, _ := auth.HashPassword("user-password")
auth.CheckPassword(storedHash, inputPassword)
```

### Session login (browser)

```go
r.Post("/login", controller.Handler(auth.Login(auth.DefaultLoginConfig(findUserByEmail))))
r.Post("/logout", controller.Handler(auth.Logout("current_user_id", "/")))
```

### JWT (API)

Sign tokens with `APP_KEY` (run `gofreight key:generate`):

```go
jwtMgr := auth.JWTFromEnv(app.Config.AppKey)

r.Post("/api/login", controller.Handler(auth.LoginWithJWT(
    auth.DefaultLoginConfig(findUserByEmail),
    jwtMgr,
)))

r.Group(func(api *router.Router) {
    api.Get("/profile", profileHandler)
}).Prefix("/api/v1").Use(auth.JWTMiddleware(jwtMgr)).Apply()
```

Clients send `Authorization: Bearer <token>`.

Optional env: `JWT_TTL=24h` (default 24 hours).

### Unified guard (JWT + API token + session)

```go
guard := auth.Guard{
    JWT:        jwtMgr,
    TokenStore: tokenStore,
    SessionKey: "current_user_id",
}
api.Use(guard.Middleware)
```

### Opaque API tokens

```go
store := auth.NewMemoryTokenStore()
api.Use(auth.APITokenMiddleware(store))
```

Generate auth scaffolding:

```bash
gofreight make:auth
```

## Authorization (policies)

Policies and role checks work with JWT, session, and API token context:

```go
policy := auth.NewPolicy()
policy.Define("edit-post", func(r *http.Request) bool {
    userID, ok := auth.UserIDFromRequest(r, "current_user_id")
    if !ok {
        return false
    }
    // ownership / role checks using userID
    return true
})
app.Router.Use(policy.RequirePolicy("edit-post"))

// Role middleware — reads role from JWT claims or session
app.Router.Use(auth.RequireRole("editor", "current_user_id"))
```

## CSRF

```go
app.UseCSRF()
```

## Security headers

Enabled automatically in production via `middleware.SecurityHeaders`.

## Rate limiting

```go
app.UseRateLimit(60, time.Minute) // 60 requests per minute per IP
```

## File uploads

Use `upload.SaveFile` with size limits and sanitized filenames. Store files in S3 via integrations when configured.

## SQL injection

- ORM queries use parameterized placeholders
- Admin SQL console is read-only; Import SQL is development-only
- Never expose admin in production
