<!--
| --------------------------------------------------------------------------
| Tutorial-auth-jwt
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
| Tutorial-auth-jwt
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

<p align="center">
  <img src="assets/gofreight-icon.png" alt="Gofreight" width="48">
</p>

# JWT Authentication

This tutorial adds JSON Web Token (JWT) authentication to your API — a login endpoint that returns a Bearer token, and middleware that protects private routes.

## What you'll build

- `POST /api/login` — authenticate and receive a JWT
- Protected routes under `/api/v1` that require `Authorization: Bearer <token>`

## Prerequisites

- An app with `APP_KEY` set (`gofreight key:generate`)
- A users table (use `gofreight make:auth` or scaffold a User model)

See **[Build a REST API](tutorial-rest-api.md)** for the API setup.

## Step 1 — Generate auth scaffolding

```bash
gofreight make:auth
gofreight migrate
```

This creates user models, login handlers, and route stubs you can customize.

## Step 2 — Create a JWT manager

JWTs are signed with `APP_KEY`:

```go
import "github.com/lsgser/gofreight/auth"

jwtMgr := auth.JWTFromEnv(app.Config.AppKey)
```

Optional: set token lifetime in `.env`:

```env
JWT_TTL=24h
```

## Step 3 — Add a login endpoint

`LoginWithJWT` validates credentials and returns a JSON token:

```go
r.Post("/api/login", controller.Handler(auth.LoginWithJWT(
    auth.DefaultLoginConfig(findUserByEmail),
    jwtMgr,
)))
```

Example response:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "type": "Bearer"
}
```

Test with curl:

```bash
curl -X POST http://localhost:5000/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret"}'
```

## Step 4 — Protect API routes

Wrap your API group with JWT middleware:

```go
jwtMgr := auth.JWTFromEnv(app.Config.AppKey)

r.Group(func(api *router.Router) {
    API(api)
}).Prefix("/api/v1").Use(auth.JWTMiddleware(jwtMgr)).Apply()
```

Clients must send the token on every request:

```bash
curl http://localhost:5000/api/v1/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

Requests without a valid token receive `401 Unauthorized`.

## Step 5 — Read the current user in handlers

Extract the authenticated user ID from the request context:

```go
func profileHandler(w http.ResponseWriter, r *http.Request) {
    userID, ok := auth.UserIDFromRequest(r, "current_user_id")
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // load user by ID...
}
```

## Step 6 — Unified guard (JWT + session + API token)

If you need one middleware that accepts JWT, opaque API tokens, or session cookies:

```go
guard := auth.Guard{
    JWT:        jwtMgr,
    TokenStore: tokenStore,
    SessionKey: "current_user_id",
}

r.Group(func(api *router.Router) {
    API(api)
}).Prefix("/api/v1").Use(guard.Middleware).Apply()
```

## Step 7 — Role-based authorization

Combine JWT auth with policies:

```go
policy := auth.NewPolicy()
policy.Define("admin", func(r *http.Request) bool {
    role, _ := auth.RoleFromRequest(r)
    return role == "admin"
})

api.Delete("/posts/:id", postsDestroy).Use(auth.RequireRole("admin"))
```

Policies work with JWT claims, session data, and API token metadata.

## Security checklist

- Run `gofreight key:generate` and keep `APP_KEY` secret
- Use HTTPS in production
- Set a reasonable `JWT_TTL` (default 24h)
- Never store JWTs in localStorage if you can use httpOnly cookies for browser apps

## Next steps

- **[Security](../docs/security.md)** — CSRF, policies, and production defaults
- **[Build a REST API](tutorial-rest-api.md)** — route groups and ApiResource
- **[Testing](../docs/testing.md)** — test authenticated endpoints with `gftest`
