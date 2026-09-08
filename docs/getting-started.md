# Getting Started

Install the Gofreight CLI, scaffold an app, and understand the core concepts in under ten minutes.

---

## Install the CLI

```bash
go install github.com/lsgser/gofreight/cmd/gofreight@latest
```

The binary installs to `$(go env GOPATH)/bin` (usually `~/go/bin`). Add it to your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
gofreight version   # gofreight v0.5.0
```

**macOS (zsh)** — persist in `~/.zshrc`:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

From source:

```bash
git clone https://github.com/lsgser/gofreight.git
cd gofreight && go install ./cmd/gofreight
```

---

## Create and run a new app

New apps use **SQLite** by default — no database server required.

```bash
gofreight new myapp
cd myapp
go mod tidy
gofreight key:generate
gofreight db:create
gofreight migrate
gofreight serve
```

| URL | Purpose |
|-----|---------|
| http://localhost:5000 | Your app |
| http://localhost:5000/admin | Database admin (development only) |
| http://localhost:5000/health | Health check JSON |

### Development with hot reload

`gofreight serve` enables hot reload automatically in development — restarts on code changes and reloads `.gft` views on each request.

```bash
gofreight serve    # hot reload in development (default)
gofreight dev      # explicit file watcher (same behavior)
```

### Production build

```bash
gofreight build
GOFREIGHT_ENV=production ./bin/myapp
```

Cross-compile for Linux: `gofreight build --os linux --arch amd64`. See **[Deployment](deployment.md)**.

---

## What you get

```
myapp/
├── main.go                 # Entry point
├── bootstrap/
│   ├── app.go              # Application wiring
│   └── schedule.go         # Scheduled tasks
├── routes/
│   ├── register.go         # Route groups
│   ├── web.go              # HTML routes
│   └── api.go              # JSON API routes
├── app/
│   ├── controllers/        # HTTP handlers
│   ├── models/             # Database models
│   ├── views/              # GFT templates (.gft)
│   ├── services/           # Business logic
│   └── ...
├── db/migrate/             # SQL migrations
├── config/                 # YAML + locales
├── public/                 # Static assets
├── storage/                # Uploads, sessions, cache
└── tests/                  # HTTP tests
```

See [Project structure](project-structure.md) for the full layout.

---

## Environment defaults

`.env` created by `gofreight new`:

```env
APP_NAME=myapp
GOFREIGHT_ENV=development
PORT=5000
DB_CONNECTION=sqlite
SESSION_DRIVER=file
CACHE_STORE=file
FILESYSTEM_DISK=local
APP_KEY=
```

Run `gofreight key:generate` immediately after scaffolding.

Full reference: [Configuration](configuration.md) and `.env.example`.

---

## Routing

Routes live in `routes/web.go` (HTML) and `routes/api.go` (JSON):

```go
// routes/register.go
func Register(r *router.Router) {
    Web(r)
    Auth(r)  // after gofreight make:auth

    r.Group(func(api *router.Router) {
        API(api)
    }).Prefix("/api/v1").Name("api.").Apply()
}
```

List registered routes:

```bash
gofreight route:list
```

The router supports groups, constraints, model binding, signed URLs, redirects, and file uploads. See [Routing](routing.md).

---

## Controllers

```go
type PostController struct{}

func (c PostController) Index(base controller.Base) error {
    posts, _ := models.Posts.All(base.Request.Context())
    return base.RenderView("posts/index", base.ViewData(map[string]any{
        "Posts": posts,
    }))
}

func (c PostController) Show(base controller.Base) error {
    return base.JSON(map[string]string{"status": "ok"})
}
```

See [Controllers](controllers.md) for status codes, redirects, uploads, and JSON responses.

---

## Application bootstrap

`bootstrap/app.go` wires the framework:

```go
func Application() *application.Application {
    app := application.New()
    _ = app.UseFileSessions("")
    _ = app.ConfigureStorage()
    _ = app.ConfigureIntegrations()
    app.UseExceptionHandler()
    app.UseVite()
    app.UseCSRF()
    return app
}
```

See [Application wiring](application-wiring.md) for every helper.

`main.go`:

```go
app := bootstrap.Application()
app.ConnectDatabase()
app.Draw(routes.Register)
app.Run()
```

---

## Generate code

Full CRUD scaffold:

```bash
gofreight make:scaffold Post title:string body:text published:boolean
gofreight migrate
gofreight serve
```

Auth starter:

```bash
gofreight make:auth
gofreight migrate
gofreight db:seed
# visit /login
```

Field types: [Generators & field types](generators.md).  
All commands: [CLI commands](commands.md) (`gofreight list`).

---

## Database

```bash
gofreight migrate              # run pending migrations
gofreight migrate:status         # check status
gofreight migrate:rollback       # undo last batch
gofreight migrate:fresh --seed   # reset + seed
gofreight db:seed              # run seeders
gofreight tinker               # interactive SQL console
```

See [Database](database.md) and [ORM](orm.md).

---

## Templates (GFT)

Gofreight Templates (`.gft`) use `#` directives:

```html
#layout "layouts/application"

<h1>{= .Title }</h1>

#each .Posts as post
  <article>{= post.Title }</article>
#endeach

#form action="/posts" method="POST"
  #field "title" label="Title"
  #token
  <button type="submit">Save</button>
#endform
```

See [Templating](templating.md) and [Forms & Validation](forms-validation.md).

---

## Testing

```bash
gofreight test              # feature tests (gftest in tests/)
gofreight test:unit           # Go unit tests (app/)
go test ./...                 # all packages — raw Go runner
```

```go
app := gftest.NewApp(t, gftest.WithMigrateDir("db/migrate"))
app.Draw(routes.Register)
app.Get("/posts").AssertOk().AssertSee("Hello")
```

See [Testing](testing.md).

---

## Next steps

| Goal | Guide |
|------|-------|
| First app walkthrough | [Tutorial: Your First App](tutorial-first-app.md) |
| REST API | [Tutorial: REST API](tutorial-rest-api.md) |
| HTML forms & CRUD | [Tutorial: HTML CRUD](tutorial-html-crud.md) |
| JWT auth | [Tutorial: JWT Authentication](tutorial-auth-jwt.md) |
| WebSockets | [Tutorial: Real-time](tutorial-realtime.md) |
| GraphQL | [Tutorial: GraphQL](tutorial-graphql.md) |
| All features | [Features overview](features.md) |
| Production deploy | [Deployment](deployment.md) |

---

## Key environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GOFREIGHT_ENV` | development | `development`, `test`, `production` |
| `PORT` | 5000 | HTTP port |
| `APP_URL` | http://localhost:5000 | Public URL |
| `APP_KEY` | (empty) | Encryption key — run `gofreight key:generate` |
| `DB_CONNECTION` | sqlite | Database driver |
| `SESSION_DRIVER` | file | Session storage |
| `CACHE_STORE` | file | Cache backend |
| `REDIS_URL` | — | Redis (sessions, cache, queue, broadcast) |

See [Configuration](configuration.md) and [Integrations](integrations.md).
