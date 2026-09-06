<p align="center">
  <img src="docs/assets/gofreight-logo.png" alt="Gofreight" width="520">
</p>

<p align="center">
  <strong>v0.2.0</strong> · <a href="https://github.com/lsgser/gofreight">github.com/lsgser/gofreight</a>
</p>

<p align="center">
  <strong>Gofreight</strong> is a batteries-included web framework for Go - routing, controllers, an ORM, migrations, <strong>Gofreight Templates (GFT)</strong>, CLI, jobs, auth, cache, mail, and tests in one framework.
</p>

<p align="center">
  Ship a complete web application as a <strong>single Go binary</strong>.
</p>

---

## Table of contents

- [Philosophy](#philosophy)
- [Quick start](#quick-start)
- [Project structure](#project-structure)
- [Setup guide](#setup-guide)
- [Templating (GFT)](#templating-gofreight-templates-gft)
- [Generators & CLI](#generators--cli)
- [Routes & controllers](#routes--controllers)
- [ORM](#orm)
- [Database & migrations](#database--migrations)
- [Testing](#testing)
- [Admin dashboard](#database-admin-local-only)
- [Integrations](#integrations)
- [Example app](#example-app)
- [Configuration](#configuration)
- [Documentation](#documentation)

---

## Philosophy

- **Go-first** — stdlib HTTP, explicit types, goroutines, and `go mod` dependencies; compile to one binary
- **Batteries included** — routing, ORM, views, migrations, CLI, queues, sessions, auth, mail, cache, and tests in one framework
- **Convention over configuration** — predictable folders, generators, and defaults so you write product code, not plumbing
- **MVC architecture** — models, views (GFT), and controllers with clear boundaries
- **RESTful by default** — resource routing for web (`Resources`) and JSON APIs (`ApiResource`)
- **Route groups** — prefix, middleware, and nested groups for APIs
- **Vendor-neutral** — no bundled payment or SaaS clients; register the integrations your app needs
- **Production-ready paths** — health checks, graceful shutdown, multi-database support, Redis queues, and Docker examples

---

## Quick start

```bash
# 1. Install the CLI
go install github.com/lsgser/gofreight/cmd/gofreight@latest

# Add Go's bin directory to PATH (once per machine; usually ~/go/bin)
export PATH="$PATH:$(go env GOPATH)/bin"
# macOS zsh — persist: echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc

# or clone https://github.com/lsgser/gofreight and from source:
git clone https://github.com/lsgser/gofreight.git && cd gofreight && go install ./cmd/gofreight

# 2. Create a new app
gofreight new myapp
cd myapp
go mod tidy
gofreight key:generate

# 3. Database (SQLite — no server required)
gofreight db:create
gofreight migrate

# 4. Generate your first resource (optional)
gofreight make:scaffold Post title:string body:text published:boolean
gofreight migrate

# 5. Run the server
gofreight serve
```

Open **http://localhost:5000** — admin at **http://localhost:5000/admin** (development only).

New apps ship with SQLite via `DB_CONNECTION=sqlite` — the database file is created at `db/development.db` automatically; no database server required.

---

## Run your app

After `gofreight new myapp`:

```bash
cd myapp
go mod tidy
gofreight key:generate
gofreight db:create
gofreight migrate
gofreight serve          # or: go run .
```

| Command | What it does |
|---------|----------------|
| `gofreight list` | List all CLI commands by namespace |
| `gofreight key:generate` | Generate a unique `APP_KEY` in `.env` |
| `gofreight serve` | Start the dev server (`go run .`) |
| `gofreight dev` | Auto-restart on file changes |
| `gofreight test` | Run the test suite |
| `gofreight migrate` | Apply pending migrations |
| `gofreight db:seed` | Seed the database |
| `gofreight make:seeder Name` | Create a Go seeder class |

Full command reference: **[docs/commands.md](docs/commands.md)**

**Switch to PostgreSQL/MySQL later** — in `.env`, set `DB_CONNECTION`, uncomment the host/credential lines, and update `DB_DATABASE`:

```env
DB_CONNECTION=pgsql
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=myapp
DB_USERNAME=postgres
DB_PASSWORD=
DB_SSLMODE=disable
```

For MySQL, use `DB_CONNECTION=mysql` and `DB_PORT=3306`.

---

## Project structure

**Your app** and **the Gofreight framework** are separate:

- Run `gofreight new myapp` → creates `myapp/` in your **current directory** (your product code).
- Install `github.com/lsgser/gofreight` as a **Go module dependency** — you do not copy framework folders into your app.

Full guide: **[docs/project-structure.md](docs/project-structure.md)** — application layout, framework repo layout, where to put new code, and local development.

### Application layout (summary)

```
myapp/
├── main.go                 # Boot Application, draw routes, run
├── bootstrap/app.go        # Service container, middleware, bindings
├── routes/
│   ├── register.go         # Route groups: web + /api/v1
│   ├── web.go              # Browser routes
│   └── api.go              # JSON API routes
├── config/
│   ├── app.yaml            # App configuration
│   └── database.go
├── app/
│   ├── controllers/
│   ├── models/
│   ├── services/
│   ├── views/              # GFT templates (.gft)
│   └── …                   # mail, jobs, middleware, policies, requests
├── public/                 # Static files → /assets/*
├── db/
│   ├── migrate/
│   ├── seeds/
│   └── seeders/
└── tests/
```

Reference implementation: **[examples/blog/](examples/blog/)** — same structure as a generated app.  
Repository: **[github.com/lsgser/gofreight](https://github.com/lsgser/gofreight)**

### Framework packages (library — not copied into your app)

| Package | Purpose |
|---------|---------|
| `application` | App bootstrap, server, wiring |
| `router` | RESTful routing |
| `controller` | HTTP controllers |
| `model` | ORM — queries, associations, validations |
| `view` | Gofreight Templates (GFT) engine |
| `database` | Migrations, schema, seeding |
| `middleware` | Sessions, CSRF, CORS, auth, locale |
| `generator` | Code generators (used by CLI) |
| `gftest` | HTTP and database testing helpers |
| `admin` | Development database dashboard |
| `integrations` | Pluggable mail, storage, cache, and custom APIs |
| `auth` | Passwords, tokens, OAuth, verification |
| `jobs` | Background jobs (memory or Redis) |
| `cache`, `mail`, `i18n`, `channels` | Infrastructure |

---

## Setup guide

### Prerequisites

- Go 1.22+
- (Optional) PostgreSQL, MySQL, or MariaDB — **SQLite is the default** for new apps; no extra install required
- (Optional) Redis for caching/queues

### Step-by-step

**1. Create the application**

```bash
gofreight new blog
cd blog
go mod tidy
```

**2. Configure `.env`** (created automatically with SQLite defaults)

```env
GOFREIGHT_ENV=development
PORT=5000
DB_CONNECTION=sqlite
APP_KEY=
```

Run `gofreight key:generate` to set the application encryption key.

**3. Set up the database**

```bash
gofreight db:create
gofreight migrate
gofreight migrate:status
gofreight db:seed
```

**4. Run the app**

```bash
gofreight serve            # http://localhost:5000
```

**5. Generate resources**

```bash
gofreight make:scaffold Post title:string body:text

# Or individual pieces:
gofreight make:model Comment body:text post_id:integer
gofreight make:controller Comment
gofreight make:migration add_index_to_posts
gofreight make:auth
gofreight migrate
```

**6. Bootstrap in `main.go`**

```go
app := application.New()
app.ConnectDatabase()
app.Draw(routes.Register)
app.Run() // views, /health, /assets, /admin (dev), graceful shutdown
```

**7. Development workflow**

```bash
gofreight dev              # auto-reload on file changes
gofreight tinker           # interactive SQL REPL
gofreight test             # run test suite
gofreight route:list       # inspect route definitions
```

**8. Docker (optional)**

```bash
docker compose up --build
```

See [docs/deployment.md](docs/deployment.md) for production deployment.

---

## Templating (Gofreight Templates — GFT)

**GFT** is Gofreight's native view language. It compiles to Go `html/template` at load time — layouts, partials, loops, and conditionals with a syntax designed for Gofreight. See **[docs/templating.md](docs/templating.md)**.

### Example (`app/views/posts/index.gft`)

```gft
#layout "layouts.application"

#slot "content"
  <h1>Posts</h1>
  #partial "partials.flash"

  #eachor .Posts as post
    <article>
      <h2>{= .Title }</h2>
      <p>{= .Body }</p>
    </article>
  #otherwise
    <p>No posts yet.</p>
  #endeach
#endslot
```

### Layout (`app/views/layouts/application.gft`)

```gft
<!DOCTYPE html>
<html>
<head><title>#place "title" "My App"</title></head>
<body>
  <main>#place "content"</main>
</body>
</html>
```

### GFT quick reference

| Directive | Purpose |
|-----------|---------|
| `#layout "path"` | Layout inheritance |
| `#slot "name"` … `#endslot` | Content region |
| `#place "name"` | Render slot in layout |
| `#partial "path"` | Include partial |
| `#when` / `#orwhen` / `#otherwise` / `#endwhen` | Conditionals |
| `#each` … `#endeach` | Loop |
| `#eachor` … `#otherwise` … `#endeach` | Loop with empty state |
| `#token` | CSRF field |
| `{= .Field }` | Escaped output |
| `{! .HTML !}` | Raw output |
| `{# comment #}` | Comment |

---

## Generators & CLI

The `gofreight` CLI scaffolds models, controllers, migrations, services, mail, jobs, tests, and full CRUD resources. Run **`gofreight list`** for every command.

### Commands

```bash
gofreight make:scaffold Post title:string body:text published:boolean
gofreight make:model User email:string
gofreight make:controller Post
gofreight make:service PaymentProcessing
gofreight make:seeder DatabaseSeeder
gofreight make:auth
```

Legacy: `gofreight generate …` and `gofreight make …` still work.

### What `make:scaffold` generates

Running `gofreight make:scaffold Post title:string body:text` creates:

| File | Description |
|------|-------------|
| `app/models/post.go` | Model + validations |
| `app/controllers/post_controller.go` | Full REST controller (index, show, new, create, edit, update, destroy) |
| `app/views/posts/*.gft` | GFT views with layouts |
| `db/migrate/001_create_posts.sql` | Migration |
| `tests/post_test.go` | HTTP tests |
| `tests/factories/post_factory.go` | Test factory |
| Updates `routes/web.go` | Registers routes automatically |

### Field types

Generators accept `name:type` pairs. Full reference: **[docs/generators.md](docs/generators.md)**.

| CLI type | Aliases | Go type | SQL (SQLite) | Form widget |
|----------|---------|---------|--------------|-------------|
| `string` | `str` | `string` | `VARCHAR(255)` | text |
| `text` | — | `string` | `TEXT` | textarea |
| `email` | — | `string` | `VARCHAR(255)` | email |
| `url` | — | `string` | `VARCHAR(512)` | url |
| `integer` | `int` | `int` | `INTEGER` | number |
| `bigint` | — | `int64` | `INTEGER` | number |
| `float` | `decimal`, `double` | `float64` | `REAL` | number |
| `boolean` | `bool` | `bool` | `INTEGER` (0/1) | checkbox |
| `datetime` | `timestamp` | `string` | `TEXT` | datetime-local |
| `date` | — | `string` | `TEXT` | date |
| `time` | — | `string` | `TEXT` | time |
| `uuid` | — | `string` | `TEXT` | text |
| `json` | `jsonb` | `string` | `TEXT` | textarea |
| `enum` | — | `string` | `TEXT` + `CHECK` | select |
| `references` | `reference`, `belongs_to` | `int64` | `INTEGER` | number (FK) |

**Examples:**

```bash
# Basic
gofreight make:scaffold Post title:string body:text published:boolean views:integer

# Enum + JSON + datetime
gofreight make:scaffold Article title:string status:enum:draft,published,archived metadata:json published_at:datetime

# Foreign key
gofreight make:scaffold Comment body:text post_id:references:posts
```

---

## Routes & controllers

Routing uses **route groups** — define routes in a callback, then chain prefix and middleware:

```go
// routes/register.go
func Register(r *router.Router) {
    Web(r)

    r.Group(func(api *router.Router) {
        API(api)
    }).Prefix("/api/v1").Use(authMw).Name("api.").Apply()
}
```

See **[docs/routing.md](docs/routing.md)** for nested groups, API resources, and middleware.

### RESTful routes (web)

```go
// routes/web.go
func Web(r *router.Router) {
    controllers.RegisterPostRoutes(r)
}
```

| Method | Path | Action | Helper |
|--------|------|--------|--------|
| GET | /posts | index | posts.index |
| GET | /posts/new | new | posts.new |
| POST | /posts | create | posts.create |
| GET | /posts/:id | show | posts.show |
| GET | /posts/:id/edit | edit | posts.edit |
| PUT/PATCH | /posts/:id | update | posts.update |
| DELETE | /posts/:id | destroy | posts.destroy |

### API routes (JSON)

Use **`ApiResource`** inside a route group (no `new`/`edit` HTML routes):

```go
// routes/api.go — prefix applied in routes/register.go
func API(r *router.Router) {
    r.ApiResource("posts", router.ApiResourceHandlers{
        Index:   controller.Handler(c.Index),
        Store:   controller.Handler(c.Store),
        Show:    controller.Handler(c.Show),
        Update:  controller.Handler(c.Update),
        Destroy: controller.Handler(c.Destroy),
    })
}
```

| Method | Path | Action |
|--------|------|--------|
| GET | /api/v1/posts | index |
| POST | /api/v1/posts | store |
| GET | /api/v1/posts/:id | show |
| PUT/PATCH | /api/v1/posts/:id | update |
| DELETE | /api/v1/posts/:id | destroy |

Or use the helper: `api.Group(r, "v1", fn, authMw)`. Full guide: **[docs/routing.md](docs/routing.md)**.

### Controller example

```go
func (c PostController) Index(base controller.Base) error {
    posts, err := models.Posts.All(context.Background())
    if err != nil {
        return err
    }
    return base.RenderView("posts/index", map[string]any{"Posts": posts})
}
```

---

## ORM

Gofreight includes a type-safe ORM with chainable queries, associations, scopes, validations, and callbacks — built for Go structs and `context.Context`.

#### Query Builder

```go
// Chainable queries
posts, _ := Posts.Query(ctx).
    WhereEq("published", true).
    Where("views", model.OpGt, 100).
    OrderDesc("created_at").
    Limit(10).
    Get()

post, _ := Posts.Query(ctx).Find(1)
post, _ := Posts.Query(ctx).FindBy("slug", "hello-world")
post, _ := Posts.Query(ctx).WhereIn("id", []any{1, 2, 3}).First()

// Aggregates
count, _ := Posts.Query(ctx).WhereEq("published", true).Count()
sum, _   := Posts.Query(ctx).Sum("views")
avg, _   := Posts.Query(ctx).Avg("rating")
titles, _ := Posts.Query(ctx).PluckStrings("title")

// Batch operations
Posts.Query(ctx).WhereEq("draft", true).UpdateAll(map[string]any{"published": true})
Posts.Query(ctx).WhereEq("draft", true).DeleteAll()
Posts.Query(ctx).FindEach(100, func(p Post) error { ... })

// Pagination
page, _ := Posts.Query(ctx).Paginate(1, 20) // Page[Post] with metadata

// Find or create
post, created, _ := Posts.Query(ctx).FirstOrCreate(map[string]any{"title": "Hello"})
```

#### Scopes

```go
var Posts = model.NewRepository[Post]("posts").
    Scope("published", func(q *model.Query[Post]) *model.Query[Post] {
        return q.WhereEq("published", true)
    }).
    Scope("recent", func(q *model.Query[Post]) *model.Query[Post] {
        return q.Latest()
    })

Posts.Query(ctx).Scope("published").Scope("recent").Get()
```

#### Associations & Eager Loading

```go
Posts.Association(model.HasManyAssociation("Comments", "comments", "post_id"))
Comments.Association(model.BelongsToAssociation("Post", "posts", "post_id"))

// Eager load (prevents N+1)
posts, _ := Posts.Query(ctx).With("Comments").Get()
comments, _ := model.GetAssociation(post, "Comments")
```

Supported: `belongs_to`, `has_one`, `has_many`, `many_to_many`

#### Models, Validations & Callbacks

```go
type Post struct {
    model.Record
    Title string `db:"title" json:"title"`
    Body  string `db:"body" json:"body"`
}

func (p *Post) Validators() []model.Validator {
    return []model.Validator{
        model.Presence("Title"),
        model.Length("Title", 3, 255),
        model.Email("AuthorEmail"),
        model.Uniqueness("Title", "title", Posts, p.ID),
        model.Inclusion("Status", []string{"draft", "published"}),
    }
}

func (p *Post) Save(ctx context.Context) error {
    return Posts.Save(ctx, p) // validates + callbacks + insert/update
}
```

**Validators:** `Presence`, `Length`, `Format`, `Numericality`, `Uniqueness`, `Inclusion`, `Email`, `Accepted`

**Callbacks:** `BeforeValidate`, `AfterValidate`, `BeforeSave`, `AfterSave`, `BeforeCreate`, `AfterCreate`, `BeforeUpdate`, `AfterUpdate`, `BeforeDelete`, `AfterDelete`

#### Transactions

```go
model.Transaction(ctx, func(txCtx context.Context) error {
    Posts.Create(txCtx, &post)
    Comments.Create(txCtx, &comment)
    return nil
})
```

#### Soft Deletes

Add a nullable `deleted_at` column, then enable on the repository:

```go
var Posts = model.NewRepository[Post]("posts").EnableSoftDelete()

Posts.Query(ctx).Find(1)                    // excludes deleted
Posts.Query(ctx).WithTrashed().Find(1)      // includes deleted
Posts.Query(ctx).OnlyTrashed().Get()        // only deleted
Posts.Destroy(ctx, record)                  // sets deleted_at
Posts.Restore(ctx, record)                  // clears deleted_at
Posts.ForceDestroy(ctx, record)             // permanent DELETE
```

See **[docs/orm.md](docs/orm.md#soft-deletes)** for the full guide.

#### Dirty Tracking

```go
d := &model.Dirty{}
d.Snapshot(post)
post.Title = "Changed"
d.Changed()        // true
d.ChangedFields()  // ["title"]
d.Changes()        // map with [original, current] pairs
```

#### Upsert (PostgreSQL)

```go
Posts.Upsert(ctx, &post, []string{"slug"})
```

---

## Database & migrations

Supports **SQLite** (default), **PostgreSQL**, **MySQL**, and **MariaDB**.

Configure with discrete **`DB_*` variables** (recommended) or `DATABASE_URL`:

```env
# SQLite (default — file path resolved automatically)
DB_CONNECTION=sqlite

# PostgreSQL
DB_CONNECTION=pgsql
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=myapp
DB_USERNAME=postgres
DB_PASSWORD=
DB_SSLMODE=disable
```

```bash
gofreight db:create
gofreight migrate
gofreight migrate:rollback
gofreight migrate:status
gofreight migrate:fresh --seed
gofreight db:seed
```

Migration files are plain SQL in `db/migrate/`:

```sql
-- db/migrate/20260101120000_create_posts.sql
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

---

## Testing

`gftest` provides readable HTTP tests with database setup, factories, and assertions:

```go
func TestPosts(t *testing.T) {
    gftest.Describe(t, "Posts", func(d *gftest.DescribeContext) {
        var app *gftest.App

        d.BeforeEach(func(t *testing.T) {
            app = gftest.NewApp(t, gftest.WithDatabase("sqlite://:memory:"))
            app.Draw(routes.Register)
        })

        d.It("lists posts", func(t *testing.T) {
            app.Get("/posts").AssertOk().AssertSee("Posts")
        })
    })
}
```

```bash
gofreight test
gofreight test ./tests/... -v -cover
```

See [docs/testing.md](docs/testing.md).

---

## Database Admin (local only)

Built-in database dashboard at `/admin` (development only). Schema management, CRUD, SQL import/export.

```bash
GOFREIGHT_ENV=development ADMIN_PASSWORD=secret go run .
# http://localhost:5000/admin
```

See [docs/admin.md](docs/admin.md).

---

## Integrations

```go
app.ConfigureIntegrations() // auto in Run() — wires cache + mail from .env

storage, _ := integrations.ActiveStorage(integrations.OsEnv{})
email, _ := integrations.ActiveEmail(integrations.OsEnv{})

// Register your own payment, SMS, CRM, or any API — then resolve by category:
payment, _ := integrations.ActivePayment(integrations.OsEnv{})
```

The framework ships **no vendor-specific clients** (no Stripe, PayFast, etc.). Register third-party APIs in your app — see [Integrations](docs/integrations.md).

---

## Example app

Full blog with GFT views, validations, and tests:

```bash
cd examples/blog
gofreight migrate
GOFREIGHT_ENV=development go run .
# http://localhost:5000/posts
```

The blog uses Gofreight Templates:

```gft
#layout "layouts.application"
#slot "content"
  #eachor .Posts as post
    <h2>{= .Title }</h2>
  #otherwise
    <p>No posts yet.</p>
  #endeach
#endslot
```

---

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GOFREIGHT_ENV` | development | `development`, `test`, `production` |
| `PORT` | 5000 | Server port |
| `HOST` | 0.0.0.0 | Bind address |
| `DB_CONNECTION` | `sqlite` | `sqlite`, `pgsql`, `mysql`, `mariadb` |
| `DB_DATABASE` | `db/development.db` (SQLite) | Database name (Postgres/MySQL/MariaDB); optional for SQLite |
| `DB_HOST` / `DB_PORT` | — | Server host and port (Postgres/MySQL) |
| `DB_USERNAME` / `DB_PASSWORD` | — | Credentials (Postgres/MySQL) |
| `DATABASE_URL` | — | Optional full URL override |
| `APP_KEY` | (empty) | Application encryption key — run `gofreight key:generate` |
| `SECRET_KEY` | — | Legacy alias for `APP_KEY` |
| `ADMIN_PASSWORD` | — | Protect /admin in dev |

Integration variables (`SESSION_DRIVER`, `QUEUE_CONNECTION`, `CACHE_STORE`, `FILESYSTEM_DISK`, `MAIL_MAILER`, `REDIS_HOST`, `AWS_*`, …): see [.env.example](.env.example).

---

## Documentation

| Guide | Description |
|-------|-------------|
| [docs/README.md](docs/README.md) | Documentation index |
| [docs/changelog.md](docs/changelog.md) | Release history |
| [docs/getting-started.md](docs/getting-started.md) | Detailed setup |
| [docs/commands.md](docs/commands.md) | Full CLI reference |
| [docs/generators.md](docs/generators.md) | Generators & field types |
| [docs/project-structure.md](docs/project-structure.md) | Application vs framework layout |
| [docs/routing.md](docs/routing.md) | Route groups & API resources |
| [docs/features.md](docs/features.md) | Sessions, queues, API, i18n, channels |
| [docs/templating.md](docs/templating.md) | GFT reference |
| [docs/forms-validation.md](docs/forms-validation.md) | Vine schemas, GFT forms, flash errors |
| [docs/orm.md](docs/orm.md) | Models, queries, associations |
| [docs/testing.md](docs/testing.md) | gftest & faker guide |
| [docs/datetime.md](docs/datetime.md) | Fluent date helpers |
| [docs/integrations.md](docs/integrations.md) | Pluggable services |
| [docs/admin.md](docs/admin.md) | Database admin |
| [docs/deployment.md](docs/deployment.md) | Docker & production |
| [docs/security.md](docs/security.md) | Auth & security |

---

## Roadmap

- [x] Gofreight Templates (GFT) engine
- [x] Full CLI with generators (`make:*`, migrations, seeders, queues)
- [x] Database migration runner
- [x] Validations and callbacks on models
- [x] Middleware stack (CSRF, sessions, CORS, rate limit, locale)
- [x] Testing helpers (`gftest`, factories)
- [x] Multi-database support (Postgres, SQLite, MySQL, MariaDB)
- [x] Development database admin
- [x] Integrations registry (vendor-neutral)
- [x] Auth (bcrypt, tokens, OAuth, email verification)
- [x] Redis sessions & job queues with retries
- [x] Migration blueprint DSL
- [x] API resources, form requests, mailables
- [x] i18n, YAML config, HTTP/fragment cache, WebSocket channels
- [x] Route groups (nested prefixes, group & route middleware, API resources)
- [x] Service container, split routes, failed job handling

## License

MIT — see [LICENSE](LICENSE).
