# CLI commands

Gofreight ships a comprehensive CLI for scaffolding, migrations, queues, and day-to-day development. Run from your **application root** (where `main.go` and `.env` live).

```bash
gofreight list          # all commands grouped by namespace
gofreight help migrate  # short help for one command
gofreight               # banner + full command list
```

Legacy forms still work: `gofreight generate …`, `gofreight make …`, `gofreight db:migrate`, and `gofreight routes`.

---

## Production safety

When `GOFREIGHT_ENV=production`, mutating commands (`migrate`, `db:wipe`, `tinker`, `make:*`, `queue:clear`, and others) show a red **PRODUCTION ENVIRONMENT** banner and require typing `yes` to continue.

For CI and deploy scripts, pass **`--force`** to skip the prompt:

```bash
gofreight migrate --force
gofreight db:seed --force
```

---

## Application commands

### `gofreight new <name>`

Creates a new Gofreight application in a folder named `<name>` with SQLite by default, route files, bootstrap wiring, `.env.example`, and starter views.

```bash
gofreight new blog
cd blog
gofreight key:generate
gofreight db:create
gofreight migrate
gofreight serve
```

**What you get:** `main.go`, `bootstrap/app.go`, `routes/`, `app/`, `db/migrate/`, `public/`, `tests/`, and a branded welcome message with next steps.

---

### `gofreight serve`

Loads `.env`, prints the development server banner (local URL and admin link in development), then runs `go run .`.

```bash
gofreight serve
# Local   http://localhost:5000
# Admin   http://localhost:5000/admin  (development only)
```

Use this for normal local development. The server reads `PORT` and `HOST` from `.env`.

---

### `gofreight dev [dir]` / `gofreight watch`

Runs the app with **file watching**. When `.go`, `.html`, `.sql`, or `.css` files change under `dir` (default `.`), the dev server restarts automatically.

```bash
gofreight dev
gofreight dev ./cmd   # watch a subdirectory only
```

Equivalent to `watch` (hidden alias). Uses `go run .` under the hood.

---

### `gofreight tinker` / `gofreight db` / `gofreight console`

Opens an **interactive SQL console** connected to your application database. Loads `.env`, connects using `DATABASE_URL` / `DB_*` settings, then accepts SQL one statement at a time.

```bash
gofreight tinker
```

**Session example:**

```
Gofreight console — type SQL and press Enter (exit to quit)
gofreight> SELECT id, title FROM posts LIMIT 5;
map[id:1 title:Hello]
map[id:2 title:World]
gofreight> INSERT INTO posts (title, body) VALUES ('Draft', 'Notes');
OK
gofreight> SELECT COUNT(*) AS n FROM posts;
map[n:3]
gofreight> exit
```

**How it works:**

1. Reads `.env` and resolves the database URL from config.
2. Opens a connection via the framework ORM layer.
3. Each line you type is executed with `database.ExecQuery`.
4. `SELECT` results print as row maps; writes print `OK`.
5. Type `exit` or `quit` to leave.

**Tips:**

- Run from your app root so `.env` is found.
- Works with SQLite, PostgreSQL, and MySQL drivers configured in `.env`.
- For complex exploration, prefer small `SELECT` queries; there is no transaction wrapper — `INSERT`/`UPDATE`/`DELETE` commit immediately.
- Aliases: `gofreight db`, `gofreight console`.

---

### `gofreight test [packages]`

Runs Go tests. Default: `go test ./...`. Pass package paths to narrow scope.

```bash
gofreight test
gofreight test ./tests/...
gofreight test ./app/models/...
```

---

### `gofreight about`

Prints framework version, current environment, working directory, and masked database URL.

```bash
gofreight about
# Gofreight 0.3.0
# Environment development
# Path /Users/you/projects/blog
# Database sqlite://***@/db/development.db
```

---

### `gofreight env`

Prints the current `GOFREIGHT_ENV` / `APP_ENV` value only (useful in shell scripts).

```bash
gofreight env
# development
```

---

### `gofreight down [message]` / `gofreight up`

**Maintenance mode.** `down` writes `storage/framework/maintenance` with an optional custom message; middleware can serve a maintenance page. `up` removes that file.

```bash
gofreight down
gofreight down Deploying v2.1 — back in 10 minutes
gofreight up
```

---

### `gofreight inspire`

Prints a random programming quote (morale boost for long deploy days).

```bash
gofreight inspire
```

---

### `gofreight list` / `gofreight help`

- **`list`** — grouped, colorized command list in the terminal.
- **`help <command>`** — description and usage for one command.

```bash
gofreight list
gofreight list migrate    # filter by name or category
gofreight help make:scaffold
```

---

### `gofreight version` / `-v` / `--version`

Shows the branded Gofreight banner and installed CLI version.

```bash
gofreight version
gofreight -v
```

---

## Database commands

### `gofreight db:create`

Creates the database. For **SQLite**, creates the file and parent directories. For PostgreSQL/MySQL, prints driver info and reminds you to create the database on the server.

```bash
# .env
DB_CONNECTION=sqlite
DB_DATABASE=db/development.db

gofreight db:create
# Created SQLite database: db/development.db
```

---

### `gofreight db:show`

Displays connection name, host, port, database name, driver, and masked URL.

```bash
gofreight db:show
# Connection sqlite
# Database db/development.db
# Driver sqlite
# URL sqlite://***@/db/development.db
```

---

### `gofreight db:seed` / `gofreight db:seed --class=Name`

Seeds the database in two passes:

1. **SQL seeds** — runs all `*.sql` files in `db/seeds/`.
2. **Go seeders** — runs `cmd/seed/main.go` if present (all registered seeders, or one with `--class`).

```bash
gofreight make:seeder DatabaseSeeder
# register in cmd/seed/main.go, then:
gofreight db:seed
gofreight db:seed --class=DatabaseSeeder
```

**Example Go seeder** (`db/seeders/database_seeder.go`):

```go
func (s *DatabaseSeeder) Run(ctx context.Context) error {
    _, err := database.DB().ExecContext(ctx,
        `INSERT INTO users (email, password) VALUES (?, ?)`,
        "admin@example.com", hashedPassword,
    )
    return err
}
```

---

### `gofreight db:wipe`

**Destructive.** Drops all tables in the connected database. Use in development only.

```bash
gofreight db:wipe
# Dropping all tables...
# Done.
```

In production, you must confirm or pass `--force`.

---

## Migration commands

Migrations live in `db/migrate/` as numbered SQL files (e.g. `001_create_posts.sql`).

### `gofreight migrate` / `gofreight db:migrate`

Runs all **pending** migrations (tracks applied files in a migrations table). `db:migrate` is a legacy alias.

```bash
gofreight make:migration create_posts
# edit db/migrate/001_create_posts.sql
gofreight migrate
# Running migrations...
# Done.
```

---

### `gofreight migrate:status`

Lists each migration file and whether it is **up** or **down**.

```bash
gofreight migrate:status
# Migration                                Status
# -------------------------------------------------------
# 001_create_posts.sql                     up
# 002_add_status_to_posts.sql              down
```

Alias: `gofreight db:status`.

---

### `gofreight migrate:rollback`

Rolls back the **last applied** migration only.

```bash
gofreight migrate:rollback
```

Alias: `gofreight db:rollback`.

---

### `gofreight migrate:reset`

Rolls back **all** migrations (empty schema, migration history cleared).

```bash
gofreight migrate:reset
```

---

### `gofreight migrate:refresh`

Rolls back all migrations, then runs them again from scratch.

```bash
gofreight migrate:refresh
```

---

### `gofreight migrate:fresh` / `migrate:fresh --seed`

**Destructive.** Drops all tables, re-runs every migration, optionally seeds.

```bash
gofreight migrate:fresh
gofreight migrate:fresh --seed
```

Common local workflow after schema experiments:

```bash
gofreight migrate:fresh --seed --force   # in CI only
```

---

## Make commands (generators)

All `make:*` commands accept **`name:type`** field pairs for models and scaffolds. See [Generators & field types](generators.md) for the full type reference.

Legacy: `gofreight make scaffold Post title:string` or `gofreight generate model Post title:string`.

### `gofreight make:model <Name> [fields]`

Creates `app/models/<name>.go` and a migration.

```bash
gofreight make:model Post title:string body:text published:boolean
gofreight migrate
```

---

### `gofreight make:controller <Name>`

Creates a REST-style controller skeleton in `app/controllers/`.

```bash
gofreight make:controller Posts
```

---

### `gofreight make:migration <name>`

Creates a timestamped SQL file in `db/migrate/` (empty up/down template).

```bash
gofreight make:migration add_status_to_posts
# edit db/migrate/002_add_status_to_posts.sql
gofreight migrate
```

---

### `gofreight make:scaffold` / `make:resource`

Full **HTML CRUD**: model, controller, GFT views, migration, factory, test, and route registration hint in `routes/web.go`.

```bash
gofreight make:scaffold Post title:string body:text status:enum:draft,published
gofreight migrate
# add controllers.RegisterPostRoutes(r) in routes/web.go if not auto-inserted
gofreight serve
# visit http://localhost:5000/posts
```

`make:resource` is an alias for `make:scaffold`.

---

### `gofreight make:api <Name> [fields]`

JSON API controller + `app/resources/` serializer. Register routes in `routes/api.go` with `ApiResource`.

```bash
gofreight make:api Post title:string body:text
gofreight migrate
```

---

### `gofreight make:auth`

Installs a complete auth starter:

- `User` model + migrations (password resets, API tokens, email verification column)
- `app/controllers/auth_controller.go` — login, register, logout
- `app/views/auth/login.gft` and `register.gft`
- `routes/auth.go` wired into `routes/register.go`
- `app/auth/users.go` — `FindUserByEmail`, `RegisterUser`
- Seed snippet in `db/seeds/users.sql` (admin@example.com / `secret123`)

```bash
gofreight make:auth
gofreight migrate
gofreight db:seed
gofreight serve
# visit /login and /register
```

---

### `gofreight make:graphql`

Installs GraphQL scaffolding:

- `graphql/register.go` — mounts `/graphql` with playground
- `graphql/modules.go` — module registry and loader wiring
- `bootstrap/app.go` — calls `graphql.Mount(app)`
- Adds `graphql-go` and `dataloader` to `go.mod`

```bash
gofreight make:graphql
gofreight make:graphql-module Post title:string body:text
go mod tidy
gofreight serve
# visit /graphql/playground
```

---

### `gofreight make:graphql-module <Name> [field:type ...]`

Generates a GraphQL module file (`graphql/<name>_module.go`) with list/show queries, a `create` mutation, and a dataloader. Registers the module in `graphql/modules.go`. Runs `make:graphql` first if the folder does not exist yet.

```bash
gofreight make:graphql-module User name:string email:email
gofreight make:graphql-module Post title:string body:text author_id:references:users
```

---

### `gofreight make:service <Name>`

Business logic class in `app/services/`.

```bash
gofreight make:service PostPublishing
# register in bootstrap/app.go:
# app.Singleton("postPublishing", func() any { return services.NewPostPublishingService() })
```

---

### `gofreight make:mail` / `make:mailable <Name>`

Mailable class + GFT email view.

```bash
gofreight make:mail WelcomeEmail
gofreight make:mailable OrderShipped   # alias
```

---

### `gofreight make:job <Name>`

Queueable job in `app/jobs/`, registered by name for Redis workers.

```bash
gofreight make:job SendNewsletter
# dispatch from app code; process with gofreight queue:work
```

---

### `gofreight make:middleware <Name>`

HTTP middleware in `app/middleware/`.

```bash
gofreight make:middleware RequestLogger
# register in bootstrap/app.go or on route groups
```

---

### `gofreight make:policy <Name>`

Authorization policy in `app/policies/`.

```bash
gofreight make:policy PostPolicy
```

---

### `gofreight make:request <Name>`

Form request / Vine validator in `app/requests/`.

```bash
gofreight make:request StorePostRequest
```

---

### `gofreight make:seeder <Name>`

Go seeder in `db/seeders/` and `cmd/seed/main.go` (first run only).

```bash
gofreight make:seeder DatabaseSeeder
gofreight db:seed --class=DatabaseSeeder
```

---

### `gofreight make:factory <Name>`

Test factory with faker defaults in `tests/factories/`.

```bash
gofreight make:factory Post
```

---

### `gofreight make:test <Name>`

HTTP feature test skeleton in `tests/`.

```bash
gofreight make:test Posts
```

---

## Queue commands (Redis)

Requires `REDIS_URL` or `REDIS_HOST` in `.env`. Default queue key: `gofreight:jobs`.

### `gofreight queue:work [--queue=key]`

Long-running worker. Processes one job at a time until Ctrl+C.

```bash
# .env
REDIS_URL=redis://127.0.0.1:6379/0
QUEUE_DRIVER=redis

gofreight queue:work
gofreight queue:work --queue=emails
```

Start this in a separate terminal alongside `gofreight serve`.

---

### `gofreight queue:failed`

Lists failed jobs with ID, name, attempt count, and exception message.

```bash
gofreight queue:failed
# abc123  SendNewsletter  attempts=3  connection refused
```

---

### `gofreight queue:retry <id>`

Re-queues a single failed job by ID.

```bash
gofreight queue:retry abc123
```

---

### `gofreight queue:flush`

Removes all **failed** job records from the failed queue.

```bash
gofreight queue:flush
```

---

### `gofreight queue:clear`

Removes all **pending** jobs from the queue (does not run them).

```bash
gofreight queue:clear
```

---

## Routes, cache, views, config

### `gofreight route:list`

Scans `routes/*.go` and prints method + pattern lines found in source (static analysis, not a live router dump).

```bash
gofreight route:list
# METHOD  PATTERN
# GET     /
# GET     /posts
# POST    /posts
```

Legacy alias: `gofreight routes`.

---

### `gofreight route:clear`

Deletes `bootstrap/cache/routes.gob` if present.

```bash
gofreight route:clear
```

---

### `gofreight cache:clear`

Clears files under `storage/framework/cache/`.

```bash
gofreight cache:clear
```

---

### `gofreight view:clear`

Clears compiled GFT views under `storage/framework/views/`.

```bash
gofreight view:clear
```

Use after editing `.gft` templates if cached views look stale.

---

### `gofreight config:show [key]`

Prints loaded configuration. Secrets are masked.

```bash
gofreight config:show
gofreight config:show environment
gofreight config:show database_url
gofreight config:show port
```

Supported keys: `environment`, `env`, `host`, `port`, `database_url`, `database`, `db_connection`, `db_database`, `log_level`, `app_key`, `secret_key`.

---

### `gofreight key:generate [--show] [--force]`

Generates a random `APP_KEY` and writes it to `.env`. Required for sessions, CSRF, and encrypted cookies.

```bash
gofreight key:generate
gofreight key:generate --show    # print only, do not write
gofreight key:generate --force   # overwrite existing key
```

Run once after `gofreight new`. Never commit production keys.

---

### `gofreight optimize` / `gofreight optimize:clear`

Writes or clears a bootstrap cache marker under `bootstrap/cache/` (production performance hint).

```bash
gofreight optimize
gofreight optimize:clear
```

---

### `gofreight auth:clear-resets`

Deletes expired rows from `password_reset_tokens` (when the table exists).

```bash
gofreight auth:clear-resets
```

Schedule periodically in production or run after auth-heavy releases.

---

## Schedule commands

### `gofreight schedule:run`

Boots the app with `GOFREIGHT_SCHEDULE_RUN=1`, executes due tasks from `bootstrap/schedule.go`, and exits.

```bash
gofreight schedule:run
```

**Cron example:**

```cron
* * * * * cd /path/to/app && gofreight schedule:run
```

### `gofreight schedule:list`

Prints a reminder that tasks are defined in `bootstrap/schedule.go`. See [Scheduling](scheduling.md).

---

## Route cache

### `gofreight route:cache`

Writes route metadata to `bootstrap/cache/routes.json` for faster URL generation.

```bash
gofreight route:cache
```

Runs the app with `GOFREIGHT_ROUTE_CACHE=1` and exits after writing the cache file.

---

## End-to-end workflows

### New app from zero

```bash
gofreight new shop
cd shop
gofreight key:generate
gofreight db:create
gofreight migrate
gofreight make:scaffold Product name:string price:float
gofreight migrate
gofreight serve
```

### Debug data with tinker

```bash
gofreight tinker
gofreight> SELECT * FROM products WHERE price > 10;
gofreight> UPDATE products SET price = 9.99 WHERE id = 1;
gofreight> exit
```

### Reset local database

```bash
gofreight migrate:fresh --seed
```

### Background jobs

```bash
# terminal 1
gofreight serve

# terminal 2
gofreight queue:work
```

---

## Legacy command forms

Older tutorials and scripts may use these equivalents:

| Legacy | Modern |
|--------|--------|
| `gofreight generate model Post title:string` | `gofreight make:model Post title:string` |
| `gofreight make scaffold Post title:string` | `gofreight make:scaffold Post title:string` |
| `gofreight db:migrate` | `gofreight migrate` |
| `gofreight routes` | `gofreight route:list` |
| `gofreight db` | `gofreight tinker` |

The `generate` and `make` commands (without colon) dispatch to the same generators as `make:*`.

---

## Related guides

- [Generators & field types](generators.md) — full `name:type` reference
- [Database & migrations](database.md) — blueprint DSL and seeders
- [Jobs & queues](jobs.md) — dispatching and workers
- [Routing](routing.md) — route groups and API resources
- [Getting started](getting-started.md) — install the CLI
