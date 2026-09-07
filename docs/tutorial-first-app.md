<p align="center">
  <img src="assets/gofreight-icon.png" alt="Gofreight" width="48">
</p>

# Your First Gofreight App

This tutorial walks you through creating a new Gofreight application from scratch, running migrations, and serving your first page.

## What you'll build

A new Gofreight project with SQLite, a working dev server on port **5000**, and the built-in admin panel for browsing your database.

## Prerequisites

- Go 1.22 or later
- Git (optional, for cloning from source)

## Step 1 — Install the CLI

```bash
go install github.com/lsgser/gofreight/cmd/gofreight@latest
export PATH="$PATH:$(go env GOPATH)/bin"
gofreight version
```

You should see `gofreight v0.4.0` (or newer).

## Step 2 — Create a new app

```bash
gofreight new myapp
cd myapp
go mod tidy
```

Gofreight scaffolds a full project layout: routes, controllers, views, migrations, and a `.env` file with sensible defaults.

### Default `.env`

```env
APP_NAME=myapp
GOFREIGHT_ENV=development
PORT=5000
DB_CONNECTION=sqlite
APP_KEY=
```

SQLite stores data in `db/development.db` — no database server required.

## Step 3 — Generate an app key

Sessions, CSRF tokens, and JWT signing all depend on `APP_KEY`:

```bash
gofreight key:generate
```

This writes a unique key into `.env`. Never commit production keys to version control.

## Step 4 — Prepare the database

```bash
gofreight db:create
gofreight migrate
```

`db:create` ensures the SQLite file exists. `migrate` runs all files in `database/migrations/`.

## Step 5 — Start the dev server

```bash
gofreight serve
```

Open **http://localhost:5000** in your browser. You should see the welcome page.

In development, the database admin is available at **http://localhost:5000/admin**.

## Step 6 — Add your first route

Routes for HTML pages live in `routes/web.go`. Open it and add a simple handler:

```go
r.Get("/hello", func(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "text/plain")
    fmt.Fprintln(w, "Hello from Gofreight!")
})
```

Restart `gofreight serve` and visit **http://localhost:5000/hello**.

## Step 7 — Scaffold a resource (optional)

Generate a full CRUD stack — model, migration, controller, views, and routes — in one command:

```bash
gofreight make:scaffold Post title:string body:text published:boolean
gofreight migrate
```

Visit **http://localhost:5000/posts** to browse your new resource.

## Project layout at a glance

| Path | Purpose |
|------|---------|
| `routes/web.go` | HTML routes |
| `routes/api.go` | JSON API routes |
| `routes/register.go` | Wires web + API route groups |
| `controllers/` | Request handlers |
| `views/` | GFT templates |
| `models/` | Database models |
| `database/migrations/` | Schema migrations |

## Next steps

- **[Build a REST API](tutorial-rest-api.md)** — JSON endpoints with route groups
- **[HTML CRUD with GFT](tutorial-html-crud.md)** — forms, validation, and templates
- **[JWT Authentication](tutorial-auth-jwt.md)** — protect API routes with Bearer tokens
