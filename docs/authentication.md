# Authentication

Gofreight supports session login (HTML), JWT (SPA/mobile API), and opaque API tokens — with a unified `Guard` that accepts any of the three.

## Password hashing

```go
import "github.com/lsgser/gofreight/auth"

hash, err := auth.HashPassword("secret-password")
err = auth.CheckPassword(storedHash, inputPassword)
```

Use bcrypt via the `auth` package for all stored passwords.

## Session login (browser)

Store the authenticated user ID in the session after validating credentials:

```go
cfg := auth.DefaultLoginConfig(findUserByEmail)

r.Get("/login", showLoginForm)
r.Post("/login", controller.Handler(auth.Login(cfg)))
r.Post("/logout", controller.Handler(auth.Logout("current_user_id", "/")))
```

`DefaultLoginConfig` sets:

- `SessionKey`: `"current_user_id"`
- `RedirectTo`: `"/"`
- `FindUser`: your lookup function returning `*auth.User`

Read the current user in controllers:

```go
session := middleware.SessionFromContext(base.Request.Context())
userID := session.Get("current_user_id")
```

See **[Sessions](sessions.md)** for the session API.

## JWT (API)

Sign tokens with `APP_KEY` (run `gofreight key:generate`):

```go
jwtMgr := auth.JWTFromEnv(app.Config.AppKey)

r.Post("/api/login", controller.Handler(auth.LoginWithJWT(
    auth.DefaultLoginConfig(findUserByEmail),
    jwtMgr,
)))
```

Login response:

```json
{
  "token_type": "Bearer",
  "access_token": "...",
  "expires_at": "2026-01-02T10:00:00Z",
  "user": { "id": 1, "email": "user@example.com", "role": "editor" }
}
```

Protect routes:

```go
r.Group(func(api *router.Router) {
    api.Get("/profile", profileHandler)
}).Prefix("/api/v1").Use(auth.JWTMiddleware(jwtMgr)).Apply()
```

Clients send `Authorization: Bearer <token>`.

Optional env: `JWT_TTL=24h` (default 24 hours).

See **[Tutorial: JWT Authentication](../examples/blog/)** (web: `tutorial-auth-jwt.md`).

## Opaque API tokens

For long-lived machine-to-machine tokens:

```go
store := auth.NewMemoryTokenStore() // use a database-backed store in production

token, _ := store.Create(userID, "mobile-app", time.Now().Add(365*24*time.Hour))

r.Group(func(api *router.Router) {
    api.Get("/data", dataHandler)
}).Use(auth.APITokenMiddleware(store)).Apply()
```

Revoke tokens with `store.Revoke(token)`.

## Unified Guard

Accept JWT, API token, or session in one middleware:

```go
guard := auth.Guard{
    JWT:        jwtMgr,
    TokenStore: tokenStore,
    SessionKey: "current_user_id",
}
api.Use(guard.Middleware)
```

Read the authenticated user ID from any auth method:

```go
userID, ok := auth.UserIDFromRequest(r, "current_user_id")
role, ok := auth.RoleFromRequest(r)
```

## OAuth2

Generic OAuth2 flow — works with any provider (Google, GitHub, etc.):

```go
provider := auth.OAuthProvider{
    Name:         "github",
    ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
    ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
    AuthURL:      "https://github.com/login/oauth/authorize",
    TokenURL:     "https://github.com/login/oauth/access_token",
    UserInfoURL:  "https://api.github.com/user",
    RedirectURL:  "http://localhost:5000/auth/github/callback",
    Scopes:       []string{"user:email"},
}

cfg := auth.OAuthConfig{
    Provider:   provider,
    OnUser:     findOrCreateUserFromOAuth,
    SessionKey: "current_user_id",
    RedirectTo: "/",
}

r.Get("/auth/github", controller.Handler(auth.OAuthRedirect(cfg)))
r.Get("/auth/github/callback", controller.Handler(auth.OAuthCallback(cfg)))
```

The callback validates state, exchanges the code, fetches user info, and logs the user in via session.

## Password reset

```go
resetStore := auth.NewMemoryPasswordResetStore()

r.Post("/password/forgot", controller.Handler(auth.RequestPasswordReset(
    resetStore,
    findUserByEmail,
    sendResetEmail, // func(email, token string) error
)))

r.Post("/password/reset", controller.Handler(auth.ResetPassword(
    resetStore,
    updateUserPassword, // func(email, newPassword string) error
)))
```

Tokens expire after one hour. The forgot endpoint always returns success (no email enumeration).

Clear expired tokens:

```bash
gofreight auth:clear-resets
```

## Email verification

```go
verifyStore := auth.NewMemoryVerificationStore()

token, _ := auth.SendVerificationEmail(verifyStore, userID, email, sendVerifyEmail)

r.Get("/email/verify", controller.Handler(auth.VerifyEmailHandler(
    verifyStore,
    markEmailVerified, // func(userID int64, email string) error
)))
```

Verification links: `GET /email/verify?token=...`

## Scaffolding

Generate User model, migration, and auth stubs:

```bash
gofreight make:auth
gofreight generate auth
```

## Production notes

- Run `gofreight key:generate` and keep `APP_KEY` secret — it signs JWTs, sessions, and CSRF tokens
- Replace in-memory token/reset/verification stores with database-backed implementations
- Use HTTPS in production so session cookies are secure
- See **[Security](security.md)** for CSRF, rate limiting, and headers

## Related

- [Authorization](authorization.md) — policies and roles
- [Sessions](sessions.md) — session storage and flash
- [Middleware](middleware.md) — auth middleware
- [Security](security.md) — production checklist
