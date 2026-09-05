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
gofreight db:create      # creates db/development.db
gofreight db:migrate     # runs db/migrate/*.sql
go run .                 # http://localhost:3000
```

Visit **http://localhost:3000**. In development, the database admin is at **http://localhost:3000/admin**.

### `.env` defaults (created by `gofreight new`)

```env
GOFREIGHT_ENV=development
PORT=3000
DATABASE_URL=sqlite://db/development.db
SECRET_KEY=change-me-in-production
```

To use PostgreSQL or MySQL instead, change `DATABASE_URL` in `.env` and ensure the server is running before `gofreight db:migrate`.

See **[Project structure](project-structure.md)** for the full application layout, where each kind of file belongs, and how the framework repo differs from your app.

## Application bootstrap

```go
app := application.New()
app.ConnectDatabase()
app.ConfigureIntegrations() // optional: load S3, Stripe, etc. from env
app.Draw(config.Routes)
app.Run()
```

`Run()` automatically mounts assets, the admin panel (development only), and configures integrations from environment variables.

## Generate code

```bash
gofreight make scaffold Post title:string body:text published:boolean
gofreight db:migrate
```

Field types (`string`, `text`, `integer`, `boolean`, `enum`, `json`, `datetime`, `references`, …): see **[Generators & field types](generators.md)**.

Or individual generators:

```bash
gofreight generate model Post title:string body:text published:boolean
gofreight generate controller Post
gofreight generate migration add_slug_to_posts
gofreight db:migrate
```

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GOFREIGHT_ENV` | development | `development`, `test`, or `production` |
| `PORT` | 3000 | HTTP port |
| `DATABASE_URL` | `sqlite://db/development.db` | SQLite, PostgreSQL, MySQL, or MariaDB URL |
| `SECRET_KEY` | change-me | Session and CSRF secret |

See [Integrations](integrations.md) for third-party service configuration.
