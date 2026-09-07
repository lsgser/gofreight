# Project structure

Gofreight involves **two different trees**:

1. **The Gofreight framework** — a Go library and CLI ([github.com/lsgser/gofreight](https://github.com/lsgser/gofreight)). You install it via `go mod`; you do not copy it into your app.
2. **Your application** — a separate project directory created with `gofreight new`. This is where your product code lives.

When you run `gofreight new myapp`, a new folder `myapp/` is created **in your current working directory**. That folder is your app. It is not nested inside the framework repo unless you choose to create it there.

---

## Your application (what `gofreight new` creates)

This is the canonical layout every generated app follows. Paths below are **fixed conventions** — the framework looks for views in `app/views`, static files in `public`, migrations in `db/migrate`, and so on.

```
myapp/
├── main.go
├── go.mod
├── .env
├── .env.example
├── .gitignore
│
├── bootstrap/
│   └── app.go                      # Service container, middleware, bindings
│
├── routes/                         # HTTP route definitions
│   ├── register.go                 # Wires web + API
│   ├── web.go                      # Browser routes
│   └── api.go                      # JSON API routes
│
├── config/
│   ├── database.go
│   ├── app.yaml
│   └── locales/
│
├── app/
│   ├── controllers/
│   ├── models/
│   ├── services/
│   ├── resources/
│   ├── mail/
│   ├── jobs/
│   ├── middleware/
│   ├── policies/
│   ├── requests/
│   └── views/
│
├── public/
├── db/migrate/
├── db/seeds/
├── db/seeders/
├── cmd/seed/                       # Go seeder runner (created by make:seeder)
├── storage/
└── tests/
```

### What each layer does

| Path | Role |
|------|------|
| `main.go` | Entry point — bootstraps `application.New()`, DB, routes, server |
| `bootstrap/app.go` | Application wiring, service container, middleware registration |
| `routes/web.go` | Browser URL mapping (HTML, sessions, CSRF) |
| `routes/api.go` | JSON API routes (registered inside `/api/v1` group) |
| `routes/register.go` | Wires web + API with route groups |
| `app/controllers/` | HTTP request handlers (`doc.go` explains the folder) |
| `app/models/` | Data layer — structs, validations, associations |
| `app/services/` | Business logic (keep controllers thin) |
| `app/resources/` | API JSON serializers |
| `app/mail/` | Mailable email classes |
| `app/jobs/` | Background job handlers |
| `app/middleware/` | Application HTTP middleware |
| `app/policies/` | Authorization policies |
| `app/requests/` | Form request validation |
| `app/views/` | GFT templates (`.gft`) |
| `config/` | YAML config, locales, database helpers |
| `public/` | Static assets (CSS, JS, images) |
| `db/migrate/` | SQL schema migrations |
| `db/seeds/` | SQL seed files |
| `db/seeders/` | Go seeder classes |
| `tests/` | HTTP and integration tests |

### Hard-coded paths the framework expects

These paths are wired in `application.New()` and related packages:

| Concern | Path | Override |
|---------|------|----------|
| Views | `app/views` | Customize in bootstrap if needed |
| Static assets | `public` | `assets.New("public")` |
| Migrations | `db/migrate/*.sql` | CLI `gofreight migrate` |
| Uploads (local) | `storage/uploads` | Upload config |
| Environment | `.env` in module root | Loaded via `godotenv` at startup |

Keep these names unless you intentionally change bootstrap code.

---

## Where to put new code

| I want to… | Put it here |
|------------|-------------|
| Add a page or API endpoint | `routes/web.go` or `routes/api.go` + controller; see [Routing](routing.md) |
| Add business logic | `gofreight make:service OrderProcessing` → `app/services/` |
| Add a database table | `gofreight make:migration …` → `db/migrate/` + model in `app/models/` |
| Add HTML | `app/views/<resource>/` as `.gft` files |
| Share markup across views | `app/views/partials/` or `app/views/components/` |
| Add CSS/JS | `public/` (served under `/assets/`) |
| Add background work | Job in `app/jobs/`; worker via `queue:work` or Application |
| Add tests | `tests/` with `gftest` |
| Add seed data | `db/seeds/*.sql` or `gofreight make:seeder` → `db/seeders/` |
| Configure services (mail, storage, custom APIs) | `.env` — see [Integrations](integrations.md) |

### Generators keep structure consistent

```bash
gofreight make:scaffold Post title:string body:text
gofreight make:service PaymentProcessing
gofreight make:api Post title:string
```

Creates, in the right places:

- `app/models/post.go`
- `app/controllers/post_controller.go`
- `app/views/posts/*.gft`
- `db/migrate/NNN_create_posts.sql`
- `app/services/payment_processing_service.go` (service generator)
- Route registration snippet for `routes/web.go` (web) or `routes/api.go` (API)

### Route registration (`routes/register.go`)

Every new app uses **route groups** to separate web and API traffic:

```go
func Register(r *router.Router) {
    Web(r)

    r.Group(func(api *router.Router) {
        API(api)
    }).Prefix("/api/v1").Use(/* optional auth middleware */).Name("api.").Apply()
}
```

- **Web** routes in `routes/web.go` — use `r.Resources()` for HTML CRUD (includes `new`/`edit`).
- **API** routes in `routes/api.go` — use `r.ApiResource()` for JSON CRUD (no `new`/`edit`).
- **Nested groups** — `.Prefix()` stacks (e.g. `/api` + `/v1` → `/api/v1`).

Full guide: **[Routing](routing.md)**.

Use generators and match existing resources — you should not need to invent folder names.

Each `app/*` directory includes a **`doc.go`** file with block comments explaining the folder's purpose and the relevant `gofreight make:*` commands — similar to guided defaults in a new project scaffold.

---

## Framework repository (this repo)

If you clone [github.com/lsgser/gofreight](https://github.com/lsgser/gofreight), you see **many top-level packages**. That is normal for a Go framework: each package is imported by applications via `go.mod`, not copied into your app.

```
gofreight/                          # Framework module (library)
├── cmd/gofreight/                  # CLI (gofreight new, migrate, make:*, …)
├── application/                    # App bootstrap, server, wiring
├── router/                         # Route groups, REST & API resources
├── controller/                     # Base controller, RenderView, JSON helpers
├── model/                          # ORM, queries, associations
├── view/                           # GFT template engine
├── database/                       # Migrations, schema, introspection
├── middleware/                     # Sessions, CSRF, CORS, rate limit, logging
├── generator/                      # Code generators (used by CLI)
├── gftest/                         # Testing helpers
├── admin/                          # Development database dashboard
├── auth/                           # Passwords, tokens, OAuth, verification
├── validation/                     # Request validation
├── integrations/                   # Pluggable registry; SMTP, S3, Redis connectors
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
├── bootstrap/app.go
├── routes/
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
require github.com/lsgser/gofreight v0.4.0
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
│    routes/ config/ db/ public/ tests/                   │
│         │                                                 │
│         │  go.mod import                                │
│         ▼                                                 │
│  Go module cache / github.com/lsgser/gofreight        │
│    router, model, view, application, …                    │
└─────────────────────────────────────────────────────────┘
```

Your app is a **thin MVC shell**. Gofreight is the **engine** imported as a Go module dependency — the same pattern as importing any other library, but with a full web stack built in.

---

## See also

- [Getting Started](getting-started.md) — create and run your first app
- [CLI commands](commands.md) — full command reference
- [Templating](templating.md) — GFT views under `app/views/`
- [Testing](testing.md) — tests under `tests/`
- [Main README](../README.md) — quick reference
