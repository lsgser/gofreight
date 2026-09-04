# Getting Started

## Install the CLI

```bash
go install github.com/gofreight/gofreight/cmd/gofreight@latest
# or from source:
go install ./cmd/gofreight
```

## Create a new app

```bash
gofreight new myapp
cd myapp
go mod tidy
gofreight db:create
gofreight db:migrate
GOFREIGHT_ENV=development go run .
```

Visit `http://localhost:3000`. In development, the database admin is at `http://localhost:3000/admin`.

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
| `DATABASE_URL` | postgres://... | PostgreSQL, SQLite, MySQL, or MariaDB URL |
| `SECRET_KEY` | change-me | Session and CSRF secret |

See [Integrations](integrations.md) for third-party service configuration.
