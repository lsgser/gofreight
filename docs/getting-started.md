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
