<!--
| --------------------------------------------------------------------------
| Jobs
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
| Jobs & Queues
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Jobs & Queues

Background job processing for emails, exports, webhooks, and any work that should not block HTTP requests.

## Overview

| Backend | Env | Use case |
|---------|-----|----------|
| In-process queue | default / `QUEUE_CONNECTION=sync` | Development, single process |
| Redis queue | `QUEUE_CONNECTION=redis` | Production, multiple workers |

```go
// bootstrap/app.go
if config.ResolveQueueConnection() == "redis" {
    _ = app.UseRedisQueue(config.ResolveRedisURL())
}
```

---

## Defining jobs

### Struct jobs

```go
// app/jobs/send_welcome_email.go
package jobs

import "context"

type SendWelcomeEmail struct {
    UserID int64
    Email  string
}

func (j SendWelcomeEmail) Handle(ctx context.Context) error {
    // send email...
    return nil
}
```

### Function jobs

```go
app.Jobs.DispatchFunc(func(ctx context.Context) error {
    log.Println("cleanup complete")
    return nil
})
```

---

## Dispatching

```go
// In-process (runs immediately in sync mode)
app.Jobs.Dispatch(jobs.SendWelcomeEmail{UserID: 1, Email: "a@b.com"})

// Via queued mailer
queued := app.QueuedMailer()
_ = queued.Send(mail.Message{ /* ... */ })
```

---

## Redis queue & named jobs

Redis workers run in **separate processes**. Job structs cannot be serialized directly — use **named jobs**:

### Register handlers

```go
import "github.com/lsgser/gofreight/jobs"

func init() {
    jobs.RegisterJob("send-welcome", func(ctx context.Context) error {
        // load user, send email...
        return nil
    })

    jobs.RegisterJob("prune-logs", func(ctx context.Context) error {
        return pruneOldLogs(ctx)
    })
}
```

Import the package from bootstrap so `init()` runs:

```go
import _ "myapp/app/jobs/handlers"
```

### Dispatch named jobs

```go
app.Jobs.Dispatch(jobs.NamedJobFunc{
    Name: "send-welcome",
    Fn:   func(ctx context.Context) error { /* runs in worker */ },
})

// Or implement NamedJob on a struct:
type WelcomeJob struct{}
func (WelcomeJob) JobName() string { return "send-welcome" }
func (WelcomeJob) Handle(ctx context.Context) error { /* ... */ }

app.Jobs.Dispatch(WelcomeJob{})
```

### Start workers

```bash
# Terminal 1
gofreight serve

# Terminal 2
gofreight queue:work
```

Or programmatically:

```go
app.StartJobs(2) // 2 concurrent workers
```

---

## CLI commands

| Command | Purpose |
|---------|---------|
| `gofreight queue:work` | Process jobs until stopped |
| `gofreight queue:listen` | Alias for `queue:work` |
| `gofreight queue:failed` | List failed jobs |
| `gofreight queue:retry` | Retry failed jobs |
| `gofreight queue:flush` | Clear the queue |
| `gofreight queue:clear` | Clear pending jobs |

Production safety: mutating queue commands require confirmation when `GOFREIGHT_ENV=production`.

---

## Retries & failed jobs

Redis queue jobs retry with exponential backoff (default **3 attempts**):

```go
app.Jobs.Dispatch(jobs.DispatchFuncJob{
    Name: "risky-task",
    Fn:   riskyFunc,
})
```

Failed jobs are stored in Redis at `{queue_key}:failed`. Inspect with:

```bash
gofreight queue:failed
gofreight queue:retry all
```

---

## Generating jobs

```bash
gofreight make:job SendNewsletter
```

Creates `app/jobs/send_newsletter.go` with a `Handle(ctx)` method stub.

---

## Queued mail

```go
queuedMailer := app.QueuedMailer()
_ = queuedMailer.Send(mail.Message{
    To:      []string{"user@example.com"},
    Subject: "Welcome",
    Body:    "Hello!",
})
```

Mail is dispatched as a background job instead of blocking the request.

---

## Testing

```go
func TestJobDispatched(t *testing.T) {
    gftest.UseFakes()
    app := gftest.NewApp(t)

    app.Jobs.Dispatch(jobs.SendWelcomeEmail{UserID: 1})
    // assert with fakes or run worker synchronously in test
}
```

For Redis queue tests, use a test Redis instance or test the job handler function directly.

---

## Architecture

```
HTTP Request
     │
     ▼
Controller dispatches job
     │
     ├── sync queue → Handle() runs immediately
     │
     └── redis queue → JSON payload pushed to Redis
                              │
                              ▼
                     queue:work worker
                              │
                              ▼
                     RegisterJob handler by name
```

---

## Best practices

- Keep jobs **idempotent** — they may retry
- Use **named jobs** for anything dispatched to Redis
- Register all job handlers in an `init()` package imported by bootstrap
- Log errors inside `Handle` — failures go to the failed queue
- Use `context.Context` for cancellation and timeouts

---

## Limitations

| Feature | Status |
|---------|--------|
| In-process queue | Supported |
| Redis queue + retries | Supported |
| Delayed/scheduled jobs | Use [Scheduling](scheduling.md) or manual `RunAt` |
| Job batches / chains | Not built-in |
| Horizon-style dashboard | Not built-in |

---

## Related

- [Scheduling](scheduling.md) — cron-style recurring tasks
- [Mail](mail.md) — queued mailables
- [Notifications](notifications.md) — async notification delivery
- [CLI Commands](commands.md) — queue commands
- [Application wiring](application-wiring.md) — `UseRedisQueue`, `StartJobs`
