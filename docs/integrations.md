# Integrations

Gofreight provides a pluggable integration registry for cloud services and third-party APIs. Integrations are configured via environment variables and initialized at startup.

## Built-in integrations

| Name | Purpose | Enable with |
|------|---------|-------------|
| `storage` | S3, Cloudflare R2, MinIO | `STORAGE_BUCKET`, credentials |
| `email` | SMTP, SendGrid | `EMAIL_FROM`, `SMTP_HOST` or `SENDGRID_API_KEY` |
| `stripe` | Payments | `STRIPE_SECRET_KEY` |
| `redis` | Cache / pub-sub | `REDIS_URL` or `UPSTASH_REDIS_URL` |
| `webhook` | Outbound webhooks | `WEBHOOK_SECRET` |
| `analytics` | Mixpanel, Segment, PostHog | `ANALYTICS_API_KEY` |

## Setup

```go
app := application.New()
app.ConfigureIntegrations() // reads os.Getenv
// or call integrations.ConfigureAll(integrations.OsEnv{}) directly
```

Integrations are also configured automatically when you call `app.Run()`.

## Using an integration

```go
import "github.com/gofreight/gofreight/integrations"

storage, ok := integrations.AsStorage(integrations.MustGet("storage"))
if ok {
    url := storage.URL("uploads/photo.jpg")
}

stripe, _ := integrations.AsStripe(integrations.MustGet("stripe"))
_ = stripe.SecretKey
```

Use `integrations.Get("name")` when the integration is optional:

```go
if email, ok := integrations.Get("email"); ok && email.Enabled() {
    // send mail
}
```

## Environment variables

### Storage (S3 / R2 / MinIO)

```env
STORAGE_PROVIDER=s3       # s3, r2, minio
STORAGE_BUCKET=my-bucket
STORAGE_REGION=us-east-1
STORAGE_ENDPOINT=          # required for MinIO/R2
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
```

### Email

```env
EMAIL_PROVIDER=smtp       # smtp, sendgrid, ses, mailgun
EMAIL_FROM=noreply@example.com
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SENDGRID_API_KEY=         # alternative to SMTP
```

### Stripe

```env
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

### Redis

```env
REDIS_URL=redis://localhost:6379
# or
UPSTASH_REDIS_URL=...
```

### Webhooks

```env
WEBHOOK_SECRET=your-signing-secret
```

### Analytics

```env
ANALYTICS_PROVIDER=posthog  # mixpanel, segment, posthog
ANALYTICS_API_KEY=...
```

## Admin status page

In development, visit `/admin/integrations` to see which integrations are active.

JSON API (development):

```go
r.Get("/integrations/status", integrations.StatusHandler())
```

## Custom integrations

See [Extending Gofreight](extending.md#custom-integrations).
