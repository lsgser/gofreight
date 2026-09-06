# Framework features

Built-in capabilities that ship with Gofreight — no extra packages required for a production-grade Go web stack.

> **Routing:** Route groups, nested prefixes, group and route middleware, web `Resources`, and API `ApiResource` — see **[Routing](routing.md)**.

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
gofreight make:api Post title:string body:text
```

```go
api.Group(r, "v1", func(api *router.Router) {
    api.ApiResource("posts", router.ApiResourceHandlers{ ... })
}, auth.APITokenMiddleware(store))
```

## 4. Auth tokens, JWT & password reset

```go
jwtMgr := auth.JWTFromEnv(app.Config.AppKey)

r.Post("/api/login", controller.Handler(auth.LoginWithJWT(
    auth.DefaultLoginConfig(findUserByEmail),
    jwtMgr,
)))

api.Group(r, "v1", func(api *router.Router) {
    api.ApiResource("posts", router.ApiResourceHandlers{ ... })
}, auth.JWTMiddleware(jwtMgr))

// Or combine JWT, opaque tokens, and session:
guard := auth.Guard{JWT: jwtMgr, TokenStore: tokenStore, SessionKey: "current_user_id"}
// r.Group(...).Use(guard.Middleware).Apply()
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

## 11. Real-time WebSockets

Socket.io-style rooms and events over WebSockets:

```go
app.MountSocket("/socket")

app.Channels.OnConnect(func(c *channels.Connection) {
    c.Join("chat:lobby")
})

app.Channels.On("chat:message", func(c *channels.Connection, raw json.RawMessage) {
    app.Channels.To("chat:lobby").Emit("chat:message", raw)
})
```

Browser client (`channels/gofreight-socket.ts`):

```typescript
const socket = new GofreightSocket('/socket')
socket.join('chat:lobby')
socket.on('chat:message', (msg) => render(msg))
socket.emit('chat:message', { text: 'Hello' })
```

Legacy channel subscribe still works: `{"action":"subscribe","channel":"posts"}` with `app.Channels.Broadcast(...)`.

Full guide: **[Real-time WebSockets](realtime.md)**

## 12. Application wiring

All features integrate via `application.Application`:

```go
app := application.New()
app.UseRedisSessions(os.Getenv("REDIS_URL"))
app.UseLocale()
app.LoadLocales("")
app.MountSocket("/socket")
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
