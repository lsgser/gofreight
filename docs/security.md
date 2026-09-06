# Security

Production security defaults, CSRF, headers, and rate limiting. For login and access control, see **[Authentication](authentication.md)** and **[Authorization](authorization.md)**.

## Production defaults

- Admin panel (`/admin`) is **never mounted** when `GOFREIGHT_ENV=production`
- Set `ADMIN_PASSWORD` in development if your dev server is network-accessible
- Run `gofreight key:generate` and keep `APP_KEY` secret — it signs sessions, CSRF tokens, and encrypted cookies

## Authentication & authorization

Session login, JWT, API tokens, OAuth, password reset, policies, and roles:

- **[Authentication](authentication.md)**
- **[Authorization](authorization.md)**

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
