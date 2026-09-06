# Date & time (`support/datetime`)

Gofreight includes a thin wrapper around [Carbon](https://github.com/dromara/carbon) for fluent date manipulation in Go.

```go
import "github.com/lsgser/gofreight/support/datetime"

now := datetime.Now()
tomorrow := datetime.Now().AddDay()
published := datetime.Parse("2026-03-01 09:00:00")

// Store in the database (Gofreight datetime columns use strings)
record.PublishedAt = datetime.DateTimeString(published)
record.StartsOn = datetime.DateString(datetime.Now())

// Convert back to time.Time when needed
t := datetime.ToTime(now)
```

Common Carbon methods are available on the returned value: `AddDays`, `SubMonth`, `StartOfDay`, `EndOfMonth`, `DiffInDays`, `IsPast`, `Format`, etc.

See the [Carbon documentation](https://github.com/dromara/carbon) for the full API.

## Timezone

Pass a timezone name to constructors:

```go
datetime.Now("America/New_York")
datetime.Parse("2026-01-01", "UTC")
```

Default timezone follows Carbon's global configuration.
