# Extending Gofreight

Gofreight is designed to be extended without forking the framework.

## Custom integrations

Register your own services in `main.go` or an `init()` in your app package:

```go
import "github.com/gofreight/gofreight/integrations"

func init() {
    integrations.RegisterCustom("slack", func(cfg map[string]string) error {
        // cfg contains SLACK_URL, SLACK_API_KEY, etc.
        return nil
    })
}
```

Environment variables follow the pattern `{NAME}_{KEY}`:

```env
SLACK_URL=https://hooks.slack.com/services/...
SLACK_API_KEY=xoxb-...
SLACK_ENABLED=true
```

## Event bus

Subscribe to integration events for decoupled workflows:

```go
integrations.Subscribe("payment.completed", func(e integrations.Event) {
    orderID := e.Payload["order_id"]
    // fulfill order, send email, etc.
})

integrations.Publish(ctx, integrations.Event{
    Name: "payment.completed",
    Payload: map[string]any{"order_id": 42},
})
```

## Registering low-level integrations

For full control, register a factory:

```go
integrations.Register("my_service", func() integrations.Integration {
    return &MyService{}
})
```

Your type must implement `Name()`, `Configure(env)`, and `Enabled()`.

## Middleware

Add global middleware in your application:

```go
app := application.New()
app.Router.Use(myMiddleware)
```

## Generators

Extend the CLI with custom generators by adding templates under `generator/templates/`.

## Database admin

The admin panel lives in `admin/` and uses embedded templates. Fork or wrap `admin.Panel` to customize the UI for your organization.

## Models and scopes

Define reusable query scopes on your models:

```go
func (Posts) Published(ctx context.Context) ([]Post, error) {
    return Posts.Query(ctx).WhereEq("published", true).Get()
}
```

See [ORM](orm.md) for associations, validations, and callbacks.
