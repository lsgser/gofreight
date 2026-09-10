<!--
| --------------------------------------------------------------------------
| Integrations
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
| Integrations
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Integrations

Gofreight provides a **generic integration registry** for connecting third-party APIs. The framework ships **no vendor-specific payment, CRM, or SaaS clients**. You register what your app needs via categories and drivers.

## How it works

| Concept | Description |
|---------|-------------|
| **Category** | Service type: mail, storage, cache, payment, analytics, webhook, or your own |
| **Driver** | A registered implementation for that category |
| **Env selection** | Standard keys: `MAIL_MAILER`, `FILESYSTEM_DISK`, `CACHE_STORE`, `QUEUE_CONNECTION`, `REDIS_HOST`, … |
| **Registry** | `integrations.Register()` adds drivers at boot |

Built-in **category connectors** (protocol-level, not vendor-branded):

| Category | Primary env | Also accepts |
|----------|-------------|--------------|
| Mail | `MAIL_MAILER` | `MAIL_DRIVER`, `EMAIL_PROVIDER` |
| Storage | `FILESYSTEM_DISK` | `STORAGE_PROVIDER` |
| Cache | `CACHE_STORE` | `CACHE_DRIVER` |
| Queue | `QUEUE_CONNECTION` | `QUEUE_DRIVER` |
| Session | `SESSION_DRIVER` | — |
| Analytics | `ANALYTICS_PROVIDER` | — |
| Webhooks | `WEBHOOK_DRIVER` | — |

**Payments, SMS, CRM, billing, and every other API** — register in your application:

```go
integrations.Register("my_gateway", func() integrations.Integration {
    return &MyPaymentGateway{}
})
```

```env
PAYMENT_PROVIDER=my_gateway
MY_GATEWAY_API_KEY=...
```

## Setup

```go
app := application.New()
app.ConfigureIntegrations() // wires cache + mail from .env
// or: integrations.BootProviders(integrations.OsEnv{})
```

Integrations configure automatically when you call `app.Run()`.

## Resolving the active driver

```go
import "github.com/lsgser/gofreight/integrations"

storage, _ := integrations.ActiveStorage(integrations.OsEnv{})
email, _ := integrations.ActiveEmail(integrations.OsEnv{})
redis, _ := integrations.ActiveRedis(integrations.OsEnv{})

// Any category — including payment after you register a driver
payment, ok := integrations.ActivePayment(integrations.OsEnv{})
if ok {
    cd, _ := integrations.AsCategorizedDriver(payment)
    log.Println("payments via", cd.DriverID())
}

// Or generic resolution
i, ok := integrations.Active(integrations.CategoryPayment, integrations.OsEnv{})
```

## Registering a third-party API

Implement `Integration` and optionally `CategorizedDriver`:

```go
type MyGateway struct {
    apiKey  string
    enabled bool
}

func (g *MyGateway) Name() string { return "my_gateway" }
func (g *MyGateway) Category() integrations.Category { return integrations.CategoryPayment }
func (g *MyGateway) DriverID() string { return "my_gateway" }

func (g *MyGateway) Configure(env integrations.EnvReader) error {
    g.apiKey = env.Get("MY_GATEWAY_API_KEY")
    g.enabled = g.apiKey != ""
    return nil
}

func (g *MyGateway) Enabled() bool { return g.enabled }

func init() {
    integrations.Register("my_gateway", func() integrations.Integration {
        return &MyGateway{}
    })
}
```

For quick one-off clients without a full type, use `RegisterCustom`:

```go
integrations.RegisterCustom("slack", func(cfg map[string]string) error {
    // SLACK_API_KEY, SLACK_URL from env
    return nil
})
```

See [Extending Gofreight](extending.md) for service providers and the event bus.

## Environment variables

### Mail

```env
MAIL_MAILER=smtp
MAIL_HOST=smtp.example.com
MAIL_PORT=587
MAIL_USERNAME=
MAIL_PASSWORD=
MAIL_FROM_ADDRESS=noreply@example.com
MAIL_FROM_NAME="${APP_NAME}"
```

Use `MAIL_MAILER=log` for development (default in `.env.example`).

### Storage (S3-compatible)

```env
FILESYSTEM_DISK=s3
AWS_BUCKET=my-bucket
AWS_DEFAULT_REGION=us-east-1
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
```

Use `FILESYSTEM_DISK=local` for local disk (default).

### Cache, queue, session, Redis

```env
CACHE_STORE=redis
QUEUE_CONNECTION=redis
SESSION_DRIVER=redis

REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=null
```

Or set `REDIS_URL` directly. For local dev defaults use `CACHE_STORE=file`, `QUEUE_CONNECTION=sync`, `SESSION_DRIVER=file`.

### Webhooks

```env
WEBHOOK_DRIVER=webhook
WEBHOOK_SECRET=your-signing-secret
```

### Your integrations

Use `{NAME}_{KEY}` env vars (e.g. `TWILIO_API_KEY`, `MY_GATEWAY_API_KEY`). Set `{CATEGORY}_PROVIDER` to the registered driver name.

## Admin status page

In development, visit `/admin/integrations` to see registered integrations and their status.

```go
r.Get("/integrations/status", integrations.StatusHandler())
```
