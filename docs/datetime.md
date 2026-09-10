<!--
| --------------------------------------------------------------------------
| Datetime
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
| Date & time
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Date & time

Fluent date/time helpers via `github.com/lsgser/gofreight/support/datetime` — parse, format, and manipulate dates without reaching for external libraries in common cases.

---

## Quick start

```go
import "github.com/lsgser/gofreight/support/datetime"

now := datetime.Now()
today := datetime.Today()

formatted := now.Format("2006-01-02 15:04:05")
parsed, err := datetime.Parse("2006-01-02", "2026-09-07")
```

---

## Creating dates

```go
datetime.Now()                          // current time
datetime.Today()                        // today at midnight
datetime.Date(2026, time.September, 7) // specific date
datetime.Parse(layout, value)         // parse string
datetime.ParseRFC3339("2026-09-07T12:00:00Z")
```

---

## Formatting

```go
dt := datetime.Now()

dt.Format("Jan 2, 2006")       // "Sep 7, 2026"
dt.Format("2006-01-02")          // "2026-09-07"
dt.Format("Monday, January 2")   // "Sunday, September 7"
dt.ToDateString()                // date-only helper
dt.ToTimeString()                // time-only helper
dt.ToDateTimeString()            // combined helper
```

Uses Go's reference time layout (`2006-01-02 15:04:05`).

---

## Manipulation

```go
dt := datetime.Now()

dt.AddDays(7)
dt.SubDays(3)
dt.AddHours(2)
dt.AddMinutes(30)

dt.StartOfDay()
dt.EndOfDay()
dt.StartOfMonth()
dt.EndOfMonth()
```

Chain operations:

```go
due := datetime.Now().AddDays(30).EndOfDay()
```

---

## Comparison

```go
a := datetime.Now()
b := datetime.Now().AddDays(1)

a.IsBefore(b)   // true
a.IsAfter(b)    // false
a.IsSameDay(b)  // true (same calendar day)
a.Equal(b)      // exact time match
```

---

## Differences

```go
a := datetime.ParseRFC3339("2026-09-07T00:00:00Z")
b := datetime.ParseRFC3339("2026-09-14T00:00:00Z")

days := a.DiffInDays(b)       // 7
hours := a.DiffInHours(b)     // 168
minutes := a.DiffInMinutes(b) // 10080
```

---

## Database integration

Store dates as ISO strings (ORM default for `datetime` fields):

```go
post.PublishedAt = datetime.Now().Format(time.RFC3339)
```

Parse when reading:

```go
published, _ := datetime.ParseRFC3339(post.PublishedAt)
if published.IsBefore(datetime.Now()) {
    // already published
}
```

---

## Templates

Pass formatted strings from controllers:

```go
return base.RenderView("posts/show", base.ViewData(map[string]any{
    "Post":       post,
    "PublishedAgo": datetime.ParseRFC3339(post.PublishedAt).DiffForHumans(datetime.Now()),
}))
```

Or format inline in Go before rendering.

---

## Testing

```go
func TestDueDate(t *testing.T) {
    due := datetime.Now().AddDays(7)
    gftest.Expect(due.IsAfter(datetime.Now())).ToBeTrue()
    gftest.Expect(due.DiffInDays(datetime.Now())).ToEqual(7)
}
```

---

## When to use stdlib `time`

Use `time` directly when you need:

- Sub-nanosecond precision
- Complex timezone database operations
- Timer/ticker concurrency primitives

The datetime helpers wrap common patterns — they don't replace the full `time` package.

---

## Related

- [ORM](orm.md) — datetime column types
- [Generators](generators.md) — `datetime`, `date`, `time` field types
