# CLI commands

Gofreight ships a comprehensive CLI for scaffolding, migrations, queues, and day-to-day development. Run **`gofreight list`** to see every command grouped by namespace.

Legacy forms still work: `gofreight generate …`, `gofreight make …`, `gofreight db:migrate`, and `gofreight routes`.

---

## Application

| Command | Description |
|---------|-------------|
| `gofreight about` | App name, environment, path, database |
| `gofreight new <name>` | Scaffold a new application |
| `gofreight serve` | `go run .` with `.env` loaded |
| `gofreight dev [dir]` | Dev server with file watching |
| `gofreight test [packages]` | Run `go test ./...` |
| `gofreight tinker` | Interactive SQL console |
| `gofreight env` | Print current environment |
| `gofreight down [message]` | Enable maintenance mode |
| `gofreight up` | Disable maintenance mode |
| `gofreight inspire` | Display a quote |
| `gofreight list` | List all commands |
| `gofreight help <cmd>` | Help for one command |
| `gofreight version` / `-v` / `--version` | Show Gofreight version (branded banner) |

Running **`gofreight`** with no arguments prints the framework name and full command list.

---

## Database & migrations

| Command | Description |
|---------|-------------|
| `gofreight db` | SQL console (alias for tinker) |
| `gofreight db:create` | Create SQLite file or print driver hints |
| `gofreight db:show` | Connection URL and driver |
| `gofreight db:wipe` | Drop all tables |
| `gofreight db:seed` | Run `db/seeds/*.sql` and Go seeders |
| `gofreight db:seed --class=UserSeeder` | Run one Go seeder via `cmd/seed` |
| `gofreight migrate` | Run pending migrations |
| `gofreight migrate:status` | Migration up/down status |
| `gofreight migrate:rollback` | Roll back last migration |
| `gofreight migrate:reset` | Roll back all migrations |
| `gofreight migrate:refresh` | Reset + migrate |
| `gofreight migrate:fresh` | Wipe + migrate |
| `gofreight migrate:fresh --seed` | Fresh + seed |

---

## Make (generators)

`make:*` commands scaffold application code (aliases for `gofreight generate` / `gofreight make`):

```bash
gofreight make:model Post title:string body:text
gofreight make:controller Posts
gofreight make:migration add_status_to_posts
gofreight make:scaffold Post title:string
gofreight make:resource Post title:string    # alias for scaffold
gofreight make:api Post title:string
gofreight make:service PostPublishing
gofreight make:mail WelcomeMail
gofreight make:job SendNewsletter
gofreight make:middleware RequestLogger
gofreight make:policy PostPolicy
gofreight make:request StorePostRequest
gofreight make:seeder DatabaseSeeder
gofreight make:factory Post
gofreight make:test Posts
gofreight make:auth
```

See [generators.md](generators.md) for field types.

### Seeders

`make:seeder` creates `db/seeders/<name>_seeder.go` and `cmd/seed/main.go` (first time). Register seeders in the map inside `cmd/seed/main.go`, then:

```bash
gofreight db:seed
gofreight db:seed --class=DatabaseSeeder
```

---

## Queue (Redis)

Requires `REDIS_URL` in `.env`.

| Command | Description |
|---------|-------------|
| `gofreight queue:work` | Process jobs until Ctrl+C |
| `gofreight queue:failed` | List failed jobs |
| `gofreight queue:retry <id>` | Re-queue a failed job |
| `gofreight queue:flush` | Remove all failed jobs |
| `gofreight queue:clear` | Remove pending jobs |

---

## Routes, cache, config

| Command | Description |
|---------|-------------|
| `gofreight route:list` | Scan `routes/*.go` for route definitions |
| `gofreight route:clear` | Remove bootstrap route cache |
| `gofreight cache:clear` | Clear `storage/framework/cache` |
| `gofreight view:clear` | Clear compiled views cache |
| `gofreight config:show [key]` | Print config values |
| `gofreight key:generate [--show] [--force]` | Generate and write `APP_KEY` to `.env` |
| `gofreight optimize` | Write bootstrap cache marker |
| `gofreight optimize:clear` | Clear bootstrap cache |
| `gofreight auth:clear-resets` | Delete expired password reset tokens |

---

## Examples

```bash
gofreight new blog
cd blog
gofreight key:generate
gofreight db:create
gofreight migrate
gofreight make:seeder DatabaseSeeder
gofreight db:seed
gofreight make:scaffold Post title:string body:text
gofreight route:list
gofreight serve
```

See [Routing](routing.md) for route groups and API resources.

---
