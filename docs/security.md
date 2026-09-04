# Security

## Production defaults

- Admin panel (`/admin`) is **never mounted** when `GOFREIGHT_ENV=production`
- Set `ADMIN_PASSWORD` in development if your dev server is network-accessible
- Use a strong `SECRET_KEY` for sessions and CSRF tokens

## Authentication

```go
import "github.com/gofreight/gofreight/auth"

hash, _ := auth.HashPassword("user-password")
// store hash in database, never store plaintext

auth.CheckPassword(storedHash, inputPassword)
```

Generate auth scaffolding:

```bash
gofreight generate auth
```

## Authorization (policies)

```go
policy := auth.NewPolicy()
policy.Define("edit-post", func(r *http.Request) bool {
    // check session role, ownership, etc.
    return true
})
app.Router.Use(policy.RequirePolicy("edit-post"))
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
