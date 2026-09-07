# Notifications

Multi-channel notifications inspired by Laravel Notifications — send mail and database inbox messages from a single notification class.

## Overview

```
Controller / Service
       │
       ▼
  app.NewNotifier(store)
       │
       ▼
  notifier.Send(ctx, userID, notification)
       │
       ├── mail channel  → app.Mailer
       └── database      → DatabaseStore
```

## The Notification interface

Every notification must implement:

```go
type Notification interface {
    Via() []string                              // channels: "mail", "database"
    ToMail(ctx context.Context) *mail.Message   // nil to skip
    ToDatabase(ctx context.Context) map[string]any // nil to skip
}
```

### Channel names

| `Via()` value | Routed to |
|---------------|-----------|
| `"mail"` or `"email"` | `Sender.Mailer.Send()` |
| `"database"` or `"db"` | `Sender.Store.Store()` |

## Complete example

```go
package notifications

import (
    "context"
    "github.com/lsgser/gofreight/mail"
)

type OrderShipped struct {
    OrderID int64
    Tracking string
}

func (n OrderShipped) Via() []string {
    return []string{"mail", "database"}
}

func (n OrderShipped) ToMail(ctx context.Context) *mail.Message {
    return &mail.Message{
        To:      []string{"customer@example.com"},
        Subject: "Your order has shipped",
        Body:    "Order #" + fmt.Sprint(n.OrderID) + " — tracking: " + n.Tracking,
    }
}

func (n OrderShipped) ToDatabase(ctx context.Context) map[string]any {
    return map[string]any{
        "type":    "order_shipped",
        "order_id": n.OrderID,
        "tracking": n.Tracking,
        "message": "Your order has shipped",
    }
}
```

## Sending notifications

```go
// In bootstrap or a service:
store := notification.NewMemoryStore() // dev/testing
notifier := app.NewNotifier(store)

// In a controller:
if err := notifier.Send(r.Context(), userID, OrderShipped{
    OrderID: 42,
    Tracking: "1Z999AA10123456784",
}); err != nil {
    return err
}
```

## BaseNotification helper

For simple notifications, embed `BaseNotification`:

```go
n := notification.BaseNotification{
    Channels: []string{"mail"},
    MailMsg: &mail.Message{
        To:      []string{"user@example.com"},
        Subject: "Hello",
        Body:    "Welcome!",
    },
}
_ = notifier.Send(ctx, userID, n)
```

## Database store

### Interface

```go
type DatabaseStore interface {
    Store(userID int64, data map[string]any) error
}
```

Implement for your `notifications` table:

```go
type SQLNotificationStore struct{}

func (s SQLNotificationStore) Store(userID int64, data map[string]any) error {
    raw, _ := json.Marshal(data)
    _, err := database.DB().Exec(
        `INSERT INTO notifications (user_id, data, created_at) VALUES (?, ?, datetime('now'))`,
        userID, string(raw),
    )
    return err
}

notifier := app.NewNotifier(SQLNotificationStore{})
```

### Suggested migration

```sql
CREATE TABLE notifications (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  type TEXT,
  data TEXT NOT NULL,
  read_at TEXT,
  created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX idx_notifications_user ON notifications(user_id);
```

### MemoryStore (development)

```go
store := notification.NewMemoryStore()
notifier := app.NewNotifier(store)

_ = notifier.Send(ctx, 1, WelcomeNotification{})

// Inspect in tests:
items := store.All(1)
```

## Mail-only notifications

```go
type PasswordResetNotice struct {
    ResetURL string
}

func (n PasswordResetNotice) Via() []string { return []string{"mail"} }

func (n PasswordResetNotice) ToMail(ctx context.Context) *mail.Message {
    return &mail.Message{
        To:      []string{"user@example.com"},
        Subject: "Reset your password",
        Body:    "Click here: " + n.ResetURL,
    }
}

func (n PasswordResetNotice) ToDatabase(ctx context.Context) map[string]any {
    return nil // skip database channel
}
```

## Database-only notifications

```go
func (n InAppAlert) Via() []string { return []string{"database"} }

func (n InAppAlert) ToMail(ctx context.Context) *mail.Message { return nil }

func (n InAppAlert) ToDatabase(ctx context.Context) map[string]any {
    return map[string]any{"type": "alert", "body": n.Body}
}
```

## Queued notifications

Queue mail delivery via the job system:

```go
app.Singleton("notifier", func() any {
    store := notification.NewMemoryStore()
    return app.NewNotifier(store)
})

// In a job:
jobs.RegisterJob("send-welcome", func(ctx context.Context) error {
    notifier := app.Make("notifier").(*notification.Sender)
    return notifier.Send(ctx, userID, WelcomeNotification{})
})

app.Jobs.Dispatch(jobs.NamedJobFunc{
    Name: "send-welcome",
    Fn:   func(ctx context.Context) error { /* ... */ },
})
```

See [Jobs & Queues](jobs.md).

## Testing

```go
func TestOrderShippedNotification(t *testing.T) {
    store := notification.NewMemoryStore()
    sender := &notification.Sender{
        Mailer: gftest.FakeMailer(),
        Store:  store,
    }

    err := sender.Send(context.Background(), 1, OrderShipped{OrderID: 5})
    gftest.Expect(err).ToBeNil()
    gftest.Expect(len(store.All(1))).ToEqual(1)
}
```

## API reference

| Type / Method | Purpose |
|---------------|---------|
| `notification.Sender` | Delivers notifications |
| `Sender.Send(ctx, userID, n)` | Send to all channels in `Via()` |
| `NewMemoryStore()` | In-memory store for dev/tests |
| `BaseNotification` | Embed for quick notifications |
| `DatabaseStore` | Interface for persistence |

## Limitations

| Feature | Status |
|---------|--------|
| Mail + database channels | Supported |
| SMS, Slack, push | Not built-in — implement custom channel via jobs |
| Notification classes generator | Not yet — create manually |
| Mark as read / pagination | App responsibility |

## Related

- [Mail](mail.md) — mailables and SMTP
- [Jobs & Queues](jobs.md) — async delivery
- [Application wiring](application-wiring.md) — `NewNotifier`
