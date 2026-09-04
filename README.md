# Gofreight

**v0.1.0** · [github.com/lsgser/gofreight](https://github.com/lsgser/gofreight)

A **Ruby on Rails / Laravel-inspired web framework for Go**. Gofreight brings MVC architecture, **Gofreight Templates (GFT)**, ActiveRecord-style models, RESTful routing, code generators, and batteries-included tooling to the Go ecosystem.

---

## Table of contents

- [Philosophy](#philosophy)
- [Quick start](#quick-start)
- [Project structure](#project-structure)
- [Setup guide](#setup-guide)
- [Templating (GFT)](#templating-gofreight-templates-gft)
- [Generators (Artisan-style)](#generators-artisan-style)
- [Routes & controllers](#routes--controllers)
- [ORM](#orm-activerecord--eloquent)
- [Database & migrations](#database--migrations)
- [Testing](#testing-gftest--pest-inspired)
- [Admin dashboard](#database-admin-local-only)
- [Integrations](#integrations)
- [Example app](#example-app)
- [Configuration](#configuration)
- [Documentation](#documentation)

---

## Philosophy

- **Convention over configuration** — sensible defaults, minimal boilerplate
- **Familiar ergonomics** — GFT views, Laravel-style generators, Rails-style routing
- **MVC architecture** — models, views, controllers with clear separation
- **RESTful by default** — `resources :posts` generates seven standard routes
- **Idiomatic Go** — leverages Go's strengths under the hood

---

## Quick start

```bash
# 1. Install the CLI
go install github.com/lsgser/gofreight/cmd/gofreight@latest
# or clone https://github.com/lsgser/gofreight and from source:
git clone https://github.com/lsgser/gofreight.git && cd gofreight && go install ./cmd/gofreight

# 2. Create a new app
gofreight new myapp
cd myapp
go mod tidy

# 3. Configure environment
cp .env.example .env   # or edit .env created by generator

# 4. Database
gofreight db:create
gofreight db:migrate

# 5. Generate your first resource (model + controller + views + tests)
gofreight make scaffold Post title:string body:text published:boolean

# 6. Run
GOFREIGHT_ENV=development go run .
```

Visit **http://localhost:3000** — admin at **http://localhost:3000/admin** (development only).

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
├── config/
│   ├── routes.go           # Routes (like routes.rb / web.php)
│   └── database.go
├── app/
│   ├── controllers/
│   ├── models/
│   └── views/              # GFT templates (.gft)
│       ├── layouts/
│       ├── partials/
│       ├── components/
│       └── posts/
├── public/                 # Static files → /assets/*
├── db/
│   ├── migrate/
│   └── seeds/
├── storage/uploads/
└── tests/
```

Reference implementation: **[examples/blog/](examples/blog/)** — same structure as a generated app.  
Repository: **[github.com/lsgser/gofreight](https://github.com/lsgser/gofreight)**

### Framework packages (library — not copied into your app)

| Package | Purpose |
|---------|---------|
| `application` | App bootstrap, wiring |
| `router` | RESTful routing |
| `controller` | HTTP controllers |
| `model` | ORM |
| `view` | Gofreight Templates (GFT) engine |
| `database` | Migrations, schema |
| `middleware` | Sessions, CSRF, CORS, auth |
| `generator` | Code generators |
| `gftest` | Pest-style testing |
| `admin` | Dev database dashboard |
| `integrations` | S3, SMTP, Stripe, Redis |
| `auth` | Bcrypt, login, policies |
| `validation` | Request validation |
| `cache`, `mail`, `jobs` | Infrastructure |

---

## Setup guide

### Prerequisites

- Go 1.22+
- PostgreSQL, SQLite, MySQL, or MariaDB
- (Optional) Redis for caching/queues

### Step-by-step

**1. Create the application**

```bash
gofreight new blog
cd blog
go mod tidy
```

**2. Configure `.env`**

```env
GOFREIGHT_ENV=development
PORT=3000
DATABASE_URL=postgres://localhost/blog_development?sslmode=disable
SECRET_KEY=your-secret-key-here
REDIS_URL=redis://localhost:6379
MAIL_DRIVER=log
```

**3. Set up the database**

```bash
gofreight db:create      # creates SQLite file or prints instructions
gofreight db:migrate     # runs db/migrate/*.sql
gofreight db:status      # shows migration status
gofreight db:seed        # runs db/seeds/*.sql
```

**4. Generate resources**

```bash
# Full CRUD (recommended — like php artisan make:scaffold)
gofreight make scaffold Post title:string body:text

# Or individual pieces:
gofreight generate model Comment body:text post_id:integer
gofreight generate controller Comment
gofreight generate migration add_index_to_posts
gofreight generate auth
```

**5. Bootstrap in `main.go`**

```go
app := application.New()
app.ConnectDatabase()
app.Draw(config.Routes)
app.Run() // loads .gft views, mounts /health, /assets, /admin (dev)
```

**6. Development workflow**

```bash
gofreight watch          # auto-reload on file changes
gofreight console        # interactive SQL REPL
gofreight test           # run test suite
```

**7. Docker (optional)**

```bash
docker compose up --build
```

See [docs/deployment.md](docs/deployment.md) for production deployment.

---

## Templating (Gofreight Templates — GFT)

**GFT** is Gofreight's native view language — inspired by Laravel and Rails, with its own syntax. See **[docs/templating.md](docs/templating.md)**.

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

## Generators (Artisan-style)

Gofreight includes Laravel/Rails-inspired generators via the CLI.

### Commands

```bash
# Laravel-style alias
gofreight make scaffold Post title:string body:text published:boolean
gofreight make model User email:string
gofreight make controller Post

# Rails-style
gofreight generate scaffold Post title:string body:text
gofreight generate model Post title:string
gofreight generate controller Post
gofreight generate migration add_slug_to_posts
gofreight generate auth
```

### What `scaffold` generates

Running `gofreight make scaffold Post title:string body:text` creates:

| File | Description |
|------|-------------|
| `app/models/post.go` | Model + validations |
| `app/controllers/post_controller.go` | Full REST controller (index, show, new, create, edit, update, destroy) |
| `app/views/posts/*.gft` | GFT views with layouts |
| `db/migrate/001_create_posts.sql` | Migration |
| `tests/post_test.go` | Pest-style tests |
| `tests/factories/post_factory.go` | Test factory |
| Updates `config/routes.go` | Registers routes automatically |

### Field types

```
string    → VARCHAR(255)
text      → TEXT (textarea in forms)
integer   → INTEGER
boolean   → BOOLEAN (checkbox)
float     → DOUBLE PRECISION
```

Example:

```bash
gofreight make scaffold Article title:string body:text published:boolean views:integer
```

---

## Routes & controllers

### RESTful routes

```go
func Routes(r *router.Router) {
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

## ORM (ActiveRecord / Eloquent)

Gofreight includes a full-featured ORM with chainable queries, associations, scopes, and more.

#### Query Builder

```go
// Chainable queries (like ActiveRecord::Relation / Eloquent Builder)
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

```go
Posts.EnableSoftDelete()
Posts.Query(ctx).SoftDelete().Find(1)       // excludes deleted
Posts.Query(ctx).WithTrashed().Find(1)      // includes deleted
Posts.Query(ctx).OnlyTrashed().Get()        // only deleted
Posts.Destroy(ctx, record)                  // sets deleted_at
```

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

Supports **PostgreSQL**, **SQLite**, **MySQL**, and **MariaDB** — auto-detected from `DATABASE_URL`:

```
DATABASE_URL=postgres://localhost/myapp_development?sslmode=disable
DATABASE_URL=sqlite://db/development.db
DATABASE_URL=mysql://user:pass@localhost/myapp_development
DATABASE_URL=mariadb://user:pass@localhost/myapp_development
```

```bash
gofreight db:create      # Create database (SQLite)
gofreight db:migrate     # Run pending migrations
gofreight db:rollback    # Rollback last migration
gofreight db:status      # Show migration status
gofreight db:seed        # Run db/seeds/*.sql
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

## Testing (gftest — Pest-inspired)

```go
func TestPosts(t *testing.T) {
    gftest.Describe(t, "Posts", func(d *gftest.DescribeContext) {
        var app *gftest.App

        d.BeforeEach(func(t *testing.T) {
            app = gftest.NewApp(t, gftest.WithDatabase("sqlite://:memory:"))
            app.Draw(config.Routes)
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
# http://localhost:3000/admin
```

See [docs/admin.md](docs/admin.md).

---

## Integrations

```go
app.ConfigureIntegrations() // auto in Run()
storage, _ := integrations.AsStorage(integrations.MustGet("storage"))
```

Built-in: S3/R2/MinIO, SMTP/SendGrid, Stripe, Redis, webhooks, analytics.

See [docs/integrations.md](docs/integrations.md) and [.env.example](.env.example).

---

## Example app

Full blog with GFT views, validations, and tests:

```bash
cd examples/blog
gofreight db:migrate
GOFREIGHT_ENV=development go run .
# http://localhost:3000/posts
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
| `PORT` | 3000 | Server port |
| `HOST` | 0.0.0.0 | Bind address |
| `DATABASE_URL` | postgres://... | Database connection |
| `SECRET_KEY` | change-me | Session/CSRF secret |
| `ADMIN_PASSWORD` | — | Protect /admin in dev |
| `REDIS_URL` | — | Redis cache/queues |
| `MAIL_DRIVER` | log | `log`, `smtp`, `sendgrid` |

Full list: [.env.example](.env.example)

---

## Documentation

| Guide | Description |
|-------|-------------|
| [docs/project-structure.md](docs/project-structure.md) | Application vs framework layout |
| [docs/README.md](docs/README.md) | Documentation index |
| [docs/getting-started.md](docs/getting-started.md) | Detailed setup |
| [docs/templating.md](docs/templating.md) | GFT reference |
| [docs/admin.md](docs/admin.md) | Database admin |
| [docs/integrations.md](docs/integrations.md) | Cloud services |
| [docs/testing.md](docs/testing.md) | gftest guide |
| [docs/deployment.md](docs/deployment.md) | Docker & production |
| [docs/security.md](docs/security.md) | Auth & security |

---

## Roadmap

- [x] Gofreight Templates (GFT) engine
- [x] Laravel-style generators (`make scaffold`)
- [x] Database migration runner CLI
- [x] Validations and callbacks on models
- [x] Middleware stack (CSRF, sessions, CORS, rate limit)
- [x] Testing helpers (`gofreight test`)
- [x] Multi-database support (Postgres, SQLite, MySQL, MariaDB)
- [x] Database admin dashboard
- [x] Integrations registry
- [x] Auth package (bcrypt, login, policies)
- [x] Production: health check, Docker, structured logging

## License

MIT — see [LICENSE](LICENSE).
