# Services & Container

Keep controllers thin by moving business logic into services. Gofreight includes a lightweight service container for dependency injection.

## Service layer

Services live in `app/services/`:

```go
// app/services/order_service.go
package services

import (
    "context"

    "myapp/app/models"
)

type OrderService struct{}

func NewOrderService() *OrderService {
    return &OrderService{}
}

func (s *OrderService) PlaceOrder(ctx context.Context, userID int64, items []Item) (*models.Order, error) {
    // validation, inventory checks, payment...
    order := &models.Order{UserID: userID}
    return order, models.Orders.Create(ctx, order)
}
```

Use from controllers:

```go
order, err := services.NewOrderService().PlaceOrder(base.Request.Context(), userID, items)
```

Generate a service stub:

```bash
gofreight make:service OrderProcessing
```

## Service container

Register services in `bootstrap/app.go`:

```go
app.Singleton("order", func() any {
    return services.NewOrderService()
})

app.Bind("notifier", func() any {
    return services.NewNotifier() // new instance each resolve
})
```

Resolve in controllers or other services:

```go
svc := app.Make("order").(*services.OrderService)
```

`Singleton` returns the same instance every time. `Bind` creates a new instance on each resolve.

## Container API

Direct access to the underlying container:

```go
import "github.com/lsgser/gofreight/container"

c := container.New()
c.Singleton("db", func() any { return database.DB() })
c.Bind("logger", func() any { return newLogger() })

svc, ok := c.Resolve("db")
svc = c.MustResolve("db") // panics if missing
c.Has("db")
```

The application container is available as `app.Container`.

## Example: demoapp bootstrap

```go
func Application() *application.Application {
    app := application.New()

    app.Singleton("example", func() any {
        return services.NewExampleService()
    })

    return app
}
```

## When to use services vs models

| Layer | Responsibility |
|-------|----------------|
| **Models** | Data persistence, validations, associations |
| **Services** | Business logic spanning multiple models or external APIs |
| **Controllers** | HTTP concerns — parse input, call services, render response |
| **Jobs** | Async work triggered by services or controllers |

## Related

- [Controllers](controllers.md) — keeping handlers thin
- [ORM](orm.md) — model layer
- [Jobs & Queues](jobs.md) — async service work
- [Project Structure](project-structure.md) — folder layout
