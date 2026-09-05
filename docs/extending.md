# Extending Gofreight

Gofreight is designed to be extended without forking the framework.

## Third-party agnostic design

The **core framework** (MVC, ORM, routing, GFT, generators) does not ship vendor-specific clients for Stripe, PayFast, Twilio, or any SaaS product. Third-party APIs connect through the **integrations registry**:

- Register any API client with `integrations.Register()` or `RegisterCustom()`.
- Select the active driver per category via env (`PAYMENT_PROVIDER`, `SMS_PROVIDER`, etc.).
- Built-in connectors cover **protocols** (SMTP, S3-compatible storage, Redis) — not branded products.

```go
storage, _ := integrations.ActiveStorage(integrations.OsEnv{})
email, _ := integrations.ActiveEmail(integrations.OsEnv{})
payment, _ := integrations.ActivePayment(integrations.OsEnv{}) // after you register a driver
```

## Integration providers

Register integrations at application boot:

```go
type AppIntegrations struct{}

func (AppIntegrations) Name() string { return "app" }

func (AppIntegrations) Register(r *integrations.Registry) {
    integrations.Register("my_gateway", func() integrations.Integration {
        return &MyPaymentGateway{}
    })
    integrations.RegisterCustom("twilio", func(cfg map[string]string) error {
        return nil
    })
}

func (AppIntegrations) Boot(env integrations.EnvReader) error { return nil }

func init() {
    integrations.RegisterProvider(AppIntegrations{})
}
```

`app.ConfigureIntegrations()` calls `BootProviders()` automatically.

## Custom integrations

```go
integrations.RegisterCustom("slack", func(cfg map[string]string) error {
    return nil
})
```

Environment variables follow `{NAME}_{KEY}`:

```env
SLACK_URL=https://hooks.slack.com/services/...
SLACK_API_KEY=xoxb-...
SLACK_ENABLED=true
```

## Registering a category driver

```go
type MyGateway struct {
    enabled bool
    apiKey  string
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

Your type must implement `Name()`, `Configure(env)`, and `Enabled()`.

## Event bus

```go
integrations.Subscribe("order.shipped", func(e integrations.Event) {
    orderID := e.Payload["order_id"]
    // call your carrier API, send email, etc.
})

integrations.Publish(ctx, integrations.Event{
    Name: "order.shipped",
    Payload: map[string]any{"order_id": 42},
})
```

## Middleware

```go
app := application.New()
app.Router.Use(myMiddleware)
```

## Generators

Extend the CLI with custom generators by adding templates under `generator/templates/`.

## Database admin

The admin panel lives in `admin/` and uses embedded templates. Fork or wrap `admin.Panel` to customize the UI for your organization.

## Models and scopes

```go
func (Posts) Published(ctx context.Context) ([]Post, error) {
    return Posts.Query(ctx).WhereEq("published", true).Get()
}
```

See [ORM](orm.md) for associations, validations, and callbacks.
