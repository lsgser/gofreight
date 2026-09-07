# Task scheduling

Cron-style recurring tasks for maintenance, reports, and periodic cleanup — inspired by Laravel's task scheduler.

## How it works

```
Cron (every minute)
       │
       ▼
gofreight schedule:run
       │
       ▼
GOFREIGHT_SCHEDULE_RUN=1 → bootstrap.RunSchedule()
       │
       ▼
scheduler.RunDue(ctx) → executes due tasks
```

The scheduler does **not** start the HTTP server — it boots the app, runs tasks, and exits.

---

## Defining tasks

`bootstrap/schedule.go` (included in new apps):

```go
package bootstrap

import (
    "context"
    "log"
    "time"

    "github.com/lsgser/gofreight/application"
)

var scheduler = application.NewScheduler()

func init() {
    scheduler.Every(time.Minute, "heartbeat", func(ctx context.Context) error {
        log.Println("scheduler: heartbeat")
        return nil
    })

    scheduler.Daily("09:00", "daily-report", func(ctx context.Context) error {
        return sendDailyReport(ctx)
    })

    scheduler.Daily("02:00", "prune-sessions", func(ctx context.Context) error {
        return pruneExpiredSessions(ctx)
    })
}

func RunSchedule() []error {
    return scheduler.RunDue(context.Background())
}
```

---

## Scheduler API

### `Every(interval, name, fn)`

Run at a fixed interval:

```go
scheduler.Every(5*time.Minute, "sync-inventory", func(ctx context.Context) error {
    return syncInventory(ctx)
})

scheduler.Every(time.Hour, "clear-cache", func(ctx context.Context) error {
    app.Cache.Flush()
    return nil
})
```

### `Daily(at, name, fn)`

Run once per day at `HH:MM` (24-hour, server local timezone):

```go
scheduler.Daily("00:00", "midnight-cleanup", func(ctx context.Context) error {
    return cleanupTempFiles()
})

scheduler.Daily("06:30", "morning-digest", func(ctx context.Context) error {
    return sendMorningDigest(ctx)
})
```

### `Tasks()` / `RunDue(ctx)`

```go
tasks := scheduler.Tasks()       // list registered tasks
errs := scheduler.RunDue(ctx)    // run all due tasks, collect errors
```

---

## Running from CLI

```bash
gofreight schedule:run
gofreight schedule:list   # prints where tasks are defined
```

### Cron setup

Every minute (Laravel-style):

```cron
* * * * * cd /var/www/myapp && /usr/local/bin/gofreight schedule:run >> storage/logs/scheduler.log 2>&1
```

Every 5 minutes:

```cron
*/5 * * * * cd /var/www/myapp && gofreight schedule:run
```

### Systemd timer (alternative)

```ini
# /etc/systemd/system/gofreight-schedule.service
[Unit]
Description=Gofreight Scheduler

[Service]
Type=oneshot
WorkingDirectory=/var/www/myapp
ExecStart=/usr/local/bin/gofreight schedule:run
User=www-data
```

```ini
# /etc/systemd/system/gofreight-schedule.timer
[Unit]
Description=Run Gofreight scheduler every minute

[Timer]
OnCalendar=*:*:00
Persistent=true

[Install]
WantedBy=timers.target
```

---

## Dispatching jobs from scheduled tasks

Prefer dispatching heavy work to the job queue:

```go
scheduler.Every(time.Hour, "queue-newsletter", func(ctx context.Context) error {
    app.Jobs.Dispatch(jobs.NamedJobFunc{
        Name: "send-newsletter",
        Fn:   sendNewsletter,
    })
    return nil
})
```

Ensure `gofreight queue:work` is running separately.

---

## Error handling

`RunDue` returns a slice of errors — one per failed task:

```go
// main.go
if os.Getenv("GOFREIGHT_SCHEDULE_RUN") == "1" {
    for _, err := range bootstrap.RunSchedule() {
        log.Printf("schedule error: %v", err)
    }
    return
}
```

Tasks should return errors for monitoring — the scheduler logs but continues other tasks.

---

## Testing scheduled tasks

Test the task function directly:

```go
func TestDailyReport(t *testing.T) {
    err := sendDailyReport(context.Background())
    gftest.Expect(err).ToBeNil()
}
```

Or test via scheduler:

```go
sched := schedule.New()
sched.Every(time.Millisecond, "test", func(ctx context.Context) error {
    called = true
    return nil
})
sched.RunDue(context.Background())
gftest.Expect(called).ToBeTrue()
```

---

## Scheduler vs job queue

| Use scheduler for | Use job queue for |
|-------------------|-------------------|
| Triggering periodic work | Heavy/async processing |
| Lightweight cleanup | Email sending |
| "Run every N minutes" | Retries and failure tracking |
| Cron-style timing | User-triggered background work |

---

## Limitations

| Feature | Status |
|---------|--------|
| Interval & daily tasks | Supported |
| Timezone configuration | Uses server local time |
| `weekly()`, `monthly()` | Not built-in — use `Every` |
| Overlap prevention | Not built-in — make tasks idempotent |
| Scheduler dashboard | Not built-in |

---

## Related

- [Jobs & Queues](jobs.md) — background processing
- [CLI Commands](commands.md) — `schedule:run`
- [Application wiring](application-wiring.md) — bootstrap lifecycle
- [Deployment](deployment.md) — production cron setup
