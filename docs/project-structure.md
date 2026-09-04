# Project structure

Gofreight involves **two different trees**:

1. **The Gofreight framework** — a Go library and CLI ([github.com/lsgser/gofreight](https://github.com/lsgser/gofreight)). You install it; you do not copy it into your app.
2. **Your application** — a separate project directory created with `gofreight new`. This is where your product code lives.

When you run `gofreight new myapp`, a new folder `myapp/` is created **in your current working directory**. That folder is your app. It is not nested inside the framework repo unless you choose to create it there.

---

## Your application (what `gofreight new` creates)

This is the canonical layout every generated app follows. Paths below are **fixed conventions** — the framework looks for views in `app/views`, static files in `public`, migrations in `db/migrate`, and so on.

```
myapp/                              # Go module root (module name = folder name by default)
├── main.go                         # Entry point — boot Application, draw routes, run server
├── go.mod                          # Depends on github.com/lsgser/gofreight
├── .env                            # Local secrets (gitignored)
├── .env.example                    # Committed template for environment variables
├── .gitignore
│
├── config/                         # App configuration (not framework config)
│   ├── routes.go                   # Route map — like routes.rb or routes/web.php
│   └── database.go                 # Optional: programmatic migration runner
│
├── app/                            # Application code (MVC)
│   ├── controllers/                # HTTP handlers
│   │   ├── home_controller.go
│   │   └── post_controller.go
│   ├── models/                     # ORM models + validations
│   │   └── post.go
│   └── views/                      # Gofreight Templates (.gft)
│       ├── layouts/
│       │   └── application.gft     # Default layout
│       ├── partials/
│       │   └── flash.gft           # Reusable fragments (#partial)
│       ├── components/             # Optional UI building blocks (partials convention)
│       ├── home/
│       │   └── index.gft
│       └── posts/                  # One folder per resource
│           ├── index.gft
│           ├── show.gft
│           ├── new.gft
│           └── edit.gft
│
├── public/                         # Static assets served at /assets/*
│   └── app.css
│
├── db/
│   ├── migrate/                    # SQL migrations (001_create_posts.sql, …)
│   └── seeds/                      # Seed SQL run via gofreight db:seed
│
├── storage/
│   └── uploads/                    # Local file uploads (development)
│
└── tests/                          # Integration / HTTP tests (gftest)
    ├── example_test.go
    └── factories/
        └── factories.go            # Test data factories
```

### What each layer does

| Path | Role | Rails / Laravel analogue |
|------|------|---------------------------|
| `main.go` | Bootstraps `application.New()`, DB, routes | `config.ru` / `public/index.php` |
| `config/routes.go` | URL → controller mapping | `config/routes.rb` / `routes/web.php` |
| `app/controllers/` | Request/response logic | `app/controllers/` |
| `app/models/` | Data, validations, associations | `app/models/` |
| `app/views/` | GFT templates | `app/views/` |
| `public/` | CSS, JS, images | `public/` |
| `db/migrate/` | Schema changes | `db/migrate/` |
| `tests/` | HTTP + DB tests | `spec/` / `tests/` |

### Hard-coded paths the framework expects

These paths are wired in `application.New()` and related packages:

| Concern | Path | Override |
|---------|------|----------|
| Views | `app/views` | `view.New("app/views")` in Application (customize in bootstrap if needed) |
| Static assets | `public` | `assets.New("public")` |
| Migrations | `db/migrate/*.sql` | CLI `gofreight db:migrate` |
| Uploads (local) | `storage/uploads` | Upload config |
| Environment | `.env` in module root | Loaded via `godotenv` at startup |

Keep these names unless you intentionally change bootstrap code.

---

## Where to put new code

| I want to… | Put it here |
|------------|-------------|
| Add a page or API endpoint | `config/routes.go` + new method in `app/controllers/` |
| Add a database table | `gofreight make migration …` → `db/migrate/` + model in `app/models/` |
| Add HTML | `app/views/<resource>/` as `.gft` files |
| Share markup across views | `app/views/partials/` or `app/views/components/` |
| Add CSS/JS | `public/` (served under `/assets/`) |
| Add background work | Enqueue in controller; worker runs via Application |
| Add tests | `tests/` with `gftest` |
| Add seed data | `db/seeds/*.sql` |
| Configure services (S3, Stripe, …) | `.env` — see [Integrations](integrations.md) |

### Generators keep structure consistent

```bash
gofreight make scaffold Post title:string body:text
```

Creates, in the right places:

- `app/models/post.go`
- `app/controllers/post_controller.go`
- `app/views/posts/*.gft`
- `db/migrate/NNN_create_posts.sql`
- Route registration snippet for `config/routes.go`

You should not need to invent folder names — use generators and match existing resources.

---

## Framework repository (this repo)

If you clone [github.com/lsgser/gofreight](https://github.com/lsgser/gofreight), you see **many top-level packages**. That is normal for a Go framework: each package is imported by applications via `go.mod`, not copied into your app.

```
gofreight/                          # Framework module (library)
├── cmd/gofreight/                  # CLI binary (gofreight new, db:migrate, make, …)
├── application/                    # App bootstrap, server, wiring
├── router/                         # RESTful routing
├── controller/                     # Base controller, RenderView, JSON helpers
├── model/                          # ORM, queries, associations
├── view/                           # GFT template engine
├── database/                       # Migrations, schema, introspection
├── middleware/                     # Sessions, CSRF, CORS, rate limit, logging
├── generator/                      # Code generators (used by CLI)
├── gftest/                         # Testing helpers
├── admin/                          # Development database dashboard
├── auth/                           # Passwords, login, policies
├── validation/                     # Request validation
├── integrations/                   # S3, SMTP, Stripe, Redis, webhooks
├── cache/, mail/, jobs/, upload/   # Infrastructure
├── assets/, health/, plugins/, dev/
├── docs/                           # Documentation (you are here)
└── examples/
    └── blog/                       # Reference application (same layout as `gofreight new`)
```

**Application developers** typically only install the CLI and import packages — they do not edit these folders.

**Framework contributors** work in this tree and run tests from the repo root:

```bash
go test ./...
```

---

## Reference app: `examples/blog`

The blog under `examples/blog/` uses **the same layout** as a generated app. Use it as a working reference:

```
examples/blog/
├── main.go
├── config/routes.go
├── app/controllers/
├── app/models/
├── app/views/
├── db/migrate/
└── tests/
```

It depends on the local framework via `replace` in `go.mod`:

```go
replace github.com/lsgser/gofreight => ../..
```

---

## Local development vs published framework

### App created anywhere (typical)

```bash
cd ~/projects
gofreight new shop
cd shop
go mod tidy
go run .
```

`go.mod` contains:

```go
require github.com/lsgser/gofreight v0.1.0
```

Go downloads the framework module from the module proxy.

### App next to a cloned framework (contributors)

If you develop the framework and an app side by side:

```go
// shop/go.mod
replace github.com/lsgser/gofreight => ../gofreight
```

Or, as in `examples/blog`, from inside the framework repo:

```go
replace github.com/lsgser/gofreight => ../..
```

---

## Optional additions (not scaffolded by default)

Add these when your app needs them — they are not required for a minimal app:

| Path | When to add |
|------|-------------|
| `Dockerfile` | Production deployment — see [Deployment](deployment.md) |
| `docker-compose.yml` | Local Postgres/Redis/MinIO stack |
| `cmd/` | Extra binaries (workers, one-off tools) |
| `internal/` | Private packages if the app grows beyond MVC folders |
| `scripts/` | Deploy or maintenance scripts |

The framework repo includes `Dockerfile` and `docker-compose.yml` as **examples for deployment**, not as something copied into every new app.

---

## Mental model

```
┌─────────────────────────────────────────────────────────┐
│  Your machine                                             │
│                                                           │
│  ~/projects/shop/          ← your app (gofreight new)     │
│    app/ controllers models views                        │
│    config/ db/ public/ tests/                           │
│         │                                                 │
│         │  import                                         │
│         ▼                                                 │
│  Go module cache / github.com/lsgser/gofreight        │
│    router, model, view, application, …                    │
└─────────────────────────────────────────────────────────┘
```

Your app is a **thin MVC shell**. Gofreight is the **engine** imported as a dependency — similar to how a Rails app lives in one folder and loads the `rails` gem from RubyGems.

---

## See also

- [Getting Started](getting-started.md) — create and run your first app
- [Templating](templating.md) — GFT views under `app/views/`
- [Testing](testing.md) — tests under `tests/`
- [Main README](../README.md) — quick reference
