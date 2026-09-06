# Jobs & Queues

Background jobs defer slow work — sending email, processing uploads, generating reports — so HTTP responses stay fast.

## Defining a job

Jobs implement the `jobs.Job` interface:

```go
// app/jobs/send_welcome_email.go
package jobs

import (
    "context"

    "github.com/lsgser/gofreight/jobs"
    "github.com/lsgser/gofreight/mail"
)

type SendWelcomeEmail struct {
    Email string
    Name  string
}

func (j SendWelcomeEmail) Handle(ctx context.Context) error {
    m := mail.NewMailable("app/views/mail", "welcome.html", "Welcome", j.Email)
    m.With("Name", j.Name)
    return m.Send(appMailer)
}
```

Or use a function:

```go
jobs.JobFunc(func(ctx context.Context) error {
    return processOrder(ctx, orderID)
})
```

Generate a job stub:

```bash
gofreight make:job SendWelcomeEmail
```

## Dispatching jobs

The application exposes an in-memory queue by default:

```go
app.Queue.Dispatch(jobs.SendWelcomeEmail{Email: user.Email, Name: user.Name})

// Or inline:
app.Queue.DispatchFunc(func(ctx context.Context) error {
    return heavyWork(ctx)
})
```

## Running workers

Start background workers with the application:

```go
app.StartJobs(2) // 2 concurrent workers
app.Run()
```

Or run a dedicated worker process:

```bash
gofreight queue:work
gofreight queue:work --concurrency=4
```

## Redis queue (production)

For multi-process deployments, use Redis:

```go
app.UseRedisQueue(os.Getenv("REDIS_URL"))
app.StartJobs(4)
```

```env
QUEUE_DRIVER=redis
REDIS_URL=redis://localhost:6379
```

Jobs persist in Redis and survive process restarts.

### Failed jobs

```bash
gofreight queue:failed
gofreight queue:retry <job-id>
gofreight queue:flush    # clear all failed jobs
gofreight queue:clear    # delete all pending jobs
```

Failed jobs are recorded after max attempts (default 3) with exponential backoff:

```go
failed, _ := redisQueue.Failed(ctx)
redisQueue.RetryFailed(ctx, jobID)
```

## Queued mail

Send email asynchronously via the queued mailer:

```go
app.QueuedMailer().Send(mail.Message{
    To:      []string{"user@example.com"},
    Subject: "Welcome",
    HTML:    "<p>Hello!</p>",
})
```

Or queue a mailable:

```go
m := mail.NewMailable("app/views/mail", "welcome.html", "Welcome", "user@example.com")
mail.QueueMailable(app.QueuedMailer(), m)
```

See **[Mail](mail.md)**.

## Testing

Process all pending jobs synchronously in tests:

```go
errs := app.Queue.Process(ctx)
if len(errs) > 0 {
    t.Fatal(errs[0])
}
```

Check pending count:

```go
if app.Queue.Pending() != 0 {
    t.Fatal("expected no pending jobs")
}
```

Flush between tests:

```go
app.Queue.Flush()
```

## Related

- [Mail](mail.md) — mailables and queued delivery
- [Configuration](configuration.md) — `QUEUE_DRIVER`, `REDIS_URL`
- [Integrations](integrations.md) — queue driver configuration
