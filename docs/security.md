# Security

Production security practices for Gofreight applications — CSRF, headers, rate limiting, signed URLs, upload hardening, and authentication.

---

## Environment hardening

```env
GOFREIGHT_ENV=production
APP_DEBUG=false
APP_KEY=base64:...   # run gofreight key:generate
```

| Setting | Production value |
|---------|------------------|
| `APP_DEBUG` | `false` |
| `GOFREIGHT_ENV` | `production` |
| `SESSION_DRIVER` | `redis` or `file` (not in-memory) |
| Database | PostgreSQL or MySQL (not SQLite) |

---

## CSRF protection

Enable on all mutating routes:

```go
app.UseCSRF()
```

GFT forms use `#token`:

```html
#form action="/posts" method="POST"
  #field "title" label="Title"
  #token
  <button type="submit">Save</button>
#endform
```

API routes using JWT/API tokens typically skip CSRF — use token auth instead.

---

## Security headers

Production mode automatically applies security headers via `middleware.SecurityHeaders`:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`

Development uses lighter middleware (`Logger` + `Recovery`).

---

## Rate limiting

Protect login and API endpoints:

```go
loginLimiter := middleware.NewRateLimiter(5, time.Minute)
apiLimiter := middleware.NewRateLimiter(100, time.Minute)

r.Post("/login", controller.Handler(c.Login)).Use(loginLimiter.Middleware)
r.Group(func(api *router.Router) {
    api.Get("/posts", controller.Handler(c.Index))
}).Prefix("/api").Use(apiLimiter.Middleware).Apply()
```

---

## Signed URLs

Time-limited links for password resets, email verification, and secure downloads:

```go
signer := app.URLSigner()
signed := signer.SignRelative("/reset?token=abc", time.Hour)

// Verify in handler:
if !signer.Verify(r.URL.RequestURI()) {
    base.Unauthorized("Invalid or expired link")
    return nil
}
```

Named route signing:

```go
url, _ := base.TemporarySignedRoute("invites.accept", time.Hour, map[string]string{
    "token": token,
})
```

See [Routing](routing.md) and [Controllers](controllers.md).

---

## Authentication security

| Method | Guidance |
|--------|----------|
| Session login | Regenerate session on login; use HTTPS |
| JWT | Set reasonable `JWT_TTL`; validate on every request |
| API tokens | Use `auth.NewDatabaseTokenStore()` in production; support revocation |

```go
store := auth.NewDatabaseTokenStore()
token, _ := store.Create(userID, "mobile-app", time.Now().Add(30*24*time.Hour))
```

See [Authentication](authentication.md).

---

## Authorization

Use policies, gates, or role middleware — never rely on hiding UI alone:

```go
r.Put("/posts/{id}", controller.Handler(c.Update)).
    Use(gate.RequireGate("update-post", "current_user_id", post))
```

See [Authorization](authorization.md).

---

## File upload security

```go
path, err := base.StoreUpload("file", "uploads", 5) // max 5 MB
```

Additional measures:

- Validate MIME types server-side
- Store uploads outside `public/` unless intentionally public
- Serve private files through authorized download routes
- Never execute uploaded files

See [Storage](storage.md).

---

## Production CLI guard

Mutating commands require confirmation in production:

```bash
gofreight migrate          # prompts: type "yes"
gofreight migrate --force  # skip prompt (CI/deploy scripts)
```

Affected: `migrate`, `db:wipe`, `db:seed`, `make:*`, `queue:clear`, `cache:clear`, and others.

---

## WebSocket security

- Validate session/JWT in `OnConnect` before joining sensitive rooms
- Do not expose admin channels without authentication
- Use Redis broadcast only over trusted networks
- Restrict CORS / origin checking before production

See [Real-time WebSockets](realtime.md).

---

## Secrets management

- Never commit `.env` — use `.env.example` as template
- Rotate `APP_KEY` if compromised (invalidates all sessions)
- Use environment variables or secret managers in production
- Keep database credentials out of source control

---

## HTTPS

Always terminate TLS in production:

- Reverse proxy (nginx, Caddy, Cloudflare)
- Set `APP_URL=https://yourdomain.com`
- Mark session cookies as Secure

---

## Security checklist

- [ ] `APP_DEBUG=false`, `GOFREIGHT_ENV=production`
- [ ] Strong unique `APP_KEY`
- [ ] CSRF enabled for HTML forms
- [ ] Rate limiting on auth endpoints
- [ ] HTTPS everywhere
- [ ] Redis sessions for multi-process
- [ ] Database-backed API tokens with expiry
- [ ] Authorization on all mutating routes
- [ ] Upload size limits and MIME validation
- [ ] Error pages don't leak stack traces
- [ ] `--force` only in trusted CI/deploy pipelines

---

## Related

- [Authentication](authentication.md)
- [Authorization](authorization.md)
- [Sessions](sessions.md)
- [Error handling](error-handling.md)
- [Deployment](deployment.md)
