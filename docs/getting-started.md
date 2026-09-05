# Getting Started

## Install the CLI

```bash
go install github.com/lsgser/gofreight/cmd/gofreight@latest
```

`go install` puts the binary in `$(go env GOPATH)/bin` (usually `~/go/bin`). That folder must be on your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
gofreight version   # should print: gofreight v0.1.0
```

**macOS (zsh)** — add to `~/.zshrc` so it persists:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

Or run without changing PATH:

```bash
$(go env GOPATH)/bin/gofreight new myapp
```

From source:

```bash
git clone https://github.com/lsgser/gofreight.git
cd gofreight && go install ./cmd/gofreight
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Create and run a new app

New apps use **SQLite by default** — no database server to install.

```bash
gofreight new myapp
cd myapp
go mod tidy
gofreight key:generate
gofreight db:create
gofreight migrate
gofreight serve          # http://localhost:5000
```

Visit **http://localhost:5000**. In development, the database admin is at **http://localhost:5000/admin**.

### `.env` defaults (created by `gofreight new`)

```env
APP_NAME=myapp
GOFREIGHT_ENV=development
PORT=5000
DB_CONNECTION=sqlite
APP_KEY=
```

Run `gofreight key:generate` after scaffolding to set a unique encryption key.

See `.env.example` for all variables. SQLite uses `db/development.db` by default — no `DB_DATABASE` line needed. For PostgreSQL, MySQL, or MariaDB, set `DB_CONNECTION` and uncomment `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USERNAME`, and `DB_PASSWORD` (use port `5432` for PostgreSQL or `3306` for MySQL/MariaDB).

See **[Project structure](project-structure.md)** for the full application layout and how the framework module differs from your app.

## Routing

Routes live in `routes/web.go` (HTML) and `routes/api.go` (JSON). **`routes/register.go`** wires them with **route groups**:

```go
func Register(r *router.Router) {
    Web(r)

    r.Group(func(api *router.Router) {
        API(api)
    }).Prefix("/api/v1").Name("api.").Apply()
}
```

Inside the group callback, define paths **without** the prefix (`/health`, not `/api/v1/health`). Chain `.Use(middleware)` for auth or rate limits. See **[Routing](routing.md)** for nested groups, `ApiResource`, and middleware order.

```bash
gofreight route:list
```

## Application bootstrap

```go
app := application.New()
app.ConnectDatabase()
app.ConfigureIntegrations() // optional: wire mail, cache, storage from .env
app.Draw(routes.Register)   // web + API routes
app.Run()
```

`Run()` mounts static assets, the admin panel (development only), health checks, and graceful shutdown.

## Generate code

```bash
gofreight make:scaffold Post title:string body:text published:boolean
gofreight migrate
```

Field types (`string`, `text`, `integer`, `boolean`, `enum`, `json`, `datetime`, `references`, …): see **[Generators & field types](generators.md)**.

Individual generators:

```bash
gofreight make:model Post title:string body:text published:boolean
gofreight make:controller Post
gofreight make:migration add_slug_to_posts
gofreight make:seeder DatabaseSeeder
gofreight migrate
```

Run **`gofreight list`** for the complete CLI. See **[CLI commands](commands.md)**.

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GOFREIGHT_ENV` | development | `development`, `test`, or `production` |
| `PORT` | 5000 | HTTP port |
| `APP_URL` | http://localhost:5000 | Public application URL |
| `DB_CONNECTION` | sqlite | Database driver |
| `DB_DATABASE` | db/development.db (SQLite) | Database name for Postgres/MySQL/MariaDB; optional for SQLite |
| `APP_KEY` | (empty) | Application encryption key — run `gofreight key:generate` |

Optional integration keys match `.env.example`: `SESSION_DRIVER`, `QUEUE_CONNECTION`, `CACHE_STORE`, `FILESYSTEM_DISK`, `MAIL_MAILER`, `REDIS_HOST`, `AWS_*`, etc.

See [Integrations](integrations.md) for configuring mail, storage, cache, and custom API drivers via environment variables.
