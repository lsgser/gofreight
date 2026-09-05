# Framework features

Built-in capabilities that ship with Gofreight — no extra packages required for a production-grade Go web stack.

## 1. Redis sessions & queues

```go
app.UseRedisSessions(os.Getenv("REDIS_URL"))
app.UseRedisQueue(os.Getenv("REDIS_URL"))
app.StartJobs(2)
```

```env
SESSION_DRIVER=redis
QUEUE_DRIVER=redis
REDIS_URL=redis://localhost:6379
```

## 2. Migration blueprint DSL

Programmatic migrations with auto-generated rollback:

```go
up, down := database.CreateTableBlueprint("comments", func(b *database.Blueprint) {
    b.IntegerColumn("post_id", colNotNull())
    b.StringColumn("body")
})
database.WriteMigrationPair("db/migrate", "004_create_comments", up, down)
```

## 3. API resources

```bash
gofreight generate api Post title:string body:text
```

```go
api.Render(w, http.StatusOK, resources.NewPostResource(post))
r.Group(api.VersionPrefix("v1"), func(api *router.Router) { ... })
```

## 4. Auth tokens & password reset

```go
tokenStore := auth.NewMemoryTokenStore()
resetStore := auth.NewMemoryPasswordResetStore()
r.Post("/password/forgot", controller.Handler(auth.RequestPasswordReset(resetStore, findUser, sendMail)))
api.Use(auth.APITokenMiddleware(tokenStore))
```

Run `gofreight generate auth` for User model + migration stubs.

## 5. Internationalization

```go
app.LoadLocales("config/locales")
app.UseLocale()
msg := app.I18n.T("welcome", map[string]string{"name": "World"})
```

Translation files: `config/locales/en.json`

## 6. Config files

Layered YAML config in `config/app.yaml` and `config/{env}.yaml`. Env vars override files.

## 7. Mailables & queued mail

```go
m := mail.NewMailable("app/views/mail", "welcome.html", "Welcome", "user@example.com")
m.With("Name", "Ada")
app.QueuedMailer().Send(mail.Message{...}) // or m.Send(app.Mailer)
```

## 8. Form requests

```go
fr, _ := request.NewFormRequest(r)
fr.Required("email", "password").MinLength("password", 8)
if !fr.Validate() {
    base.Unprocessable(fr.Errors)
}
```

## 9. Asset manifest (Vite)

```go
app.LoadAssetManifest("public/manifest.json")
url := app.Assets.Path("app.js")
```

## 10. HTTP & fragment caching

```go
app.UseHTTPCache(time.Hour)
cache.NewFragmentCache(app.Cache).Remember("sidebar", time.Minute, renderSidebar)
```

## 11. Real-time channels

```go
app.MountChannels("/cable")
app.Channels.Broadcast("posts", "created", map[string]any{"id": 1})
```

Client: `{"action":"subscribe","channel":"posts"}`

## 12. Application wiring

All features integrate via `application.Application`:

```go
app := application.New()
app.UseRedisSessions(os.Getenv("REDIS_URL"))
app.UseLocale()
app.LoadLocales("")
app.MountChannels("/cable")
app.StartJobs(2)
app.Run()
```

## 13. Date/time & faker

**Carbon-style dates** via `support/datetime`:

```go
import "github.com/lsgser/gofreight/support/datetime"

due := datetime.Now().AddDays(7)
record.DueAt = datetime.DateTimeString(due)
```

**Fake data** for tests and seeders via `gftest/faker`:

```go
import "github.com/lsgser/gofreight/gftest/faker"

name := faker.Name()
email := faker.Email()

// Generated factories use faker automatically:
// gofreight make:factory Post
```

See [Date & time](datetime.md) and [Testing](testing.md).
