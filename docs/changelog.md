# Changelog

All notable changes to Gofreight are documented here. The project follows [Semantic Versioning](https://semver.org/).

Install a specific release:

```bash
go install github.com/lsgser/gofreight/cmd/gofreight@v0.3.1
```

---

## Unreleased

---

## 0.3.1 — September 7, 2026

### Added

- **GraphQL generators** — `gofreight make:graphql` and `gofreight make:graphql-module` scaffold `graphql/`, bootstrap wiring, and resource modules with queries, mutations, and dataloaders
- **Scaffold bootstrap migration** — new apps include `db/migrate/0001_init.sql` so the first `gofreight migrate` runs visible migrations
- **Admin panel overhaul** — phpMyAdmin-inspired UI with sidebar table list, tabbed views (Browse, Structure, SQL, Search, Insert, Export), column sort, row search, bulk delete, CSV export, truncate, and SQL query history

### Changed

- **Welcome page** — redesigned home view with dark mode, feature cards, docs-site links (not GitHub), and framework version badge
- **New-app README** — links to the documentation site; includes `key:generate` in getting-started steps
- **Migration output** — clearer messages when no files exist or the database is up to date
- **Admin documentation** — reflects full CRUD, schema tools, and export capabilities

### Fixed

- **Test factories stub** — removed unused import from generated `tests/factories/factories.go`
- **GraphQL module deps** — `make:graphql` adds `graphql-go` and `dataloader` to app `go.mod`

---

## 0.3.0 — September 7, 2026

### Added — Routing

- **Declarative redirect routes** — `Redirect()`, `PermanentRedirect()`
- **Named route URL generation** — `Router.URL()`, `controller.RouteURL()`, `RedirectRoute()`
- **Route constraints** — `.Where()`, `.WhereParam()` (regex on `:id`, `:slug`, etc.)
- **Wildcard & optional parameters** — `{path*}`, `{id?}`, and colon equivalents
- **Domain & subdomain routing** — `.Domain()`, `.Subdomain()` on route groups
- **Route model binding** — `BindModel()`, `BindModelBy()`, custom `.Bind()`
- **Signed URLs** — `URLSigner`, `.Signed()` middleware, `SignedURL()`, `TemporarySignedRoute()`
- **File download & upload helpers** — `Download()`, `File()`, `StoreUpload()`, `UploadedFile()`
- **HTTP status helpers** — `Created()`, `NoContent()`, `Abort()`, symbolic `StatusFromName()`
- **Route-level status** — `.Status()`, `.StatusName()` on route registrars
- **Extra HTTP verbs** — `Any()`, `Match()`, `Head()`, `Options()`, `Fallback()`

See the expanded **[Routing](routing.md)** and **[Controllers](controllers.md)** guides.

### Added — Drivers & infrastructure

- **File session driver** — `SESSION_DRIVER=file` persists to `storage/framework/sessions/`
- **File cache store** — `CACHE_STORE=file` persists to `storage/framework/cache/data/`
- **Local storage disk** — `FILESYSTEM_DISK=local` with `app.Storage` and upload helpers
- **Redis job serialization** — named jobs (`NamedJob`, `RegisterJob`) for Redis workers
- **Task scheduler** — `schedule.Scheduler`, `gofreight schedule:run`, `bootstrap/schedule.go`
- **Notifications** — multi-channel sender (mail + database store callback)
- **Exception handler** — `app.UseExceptionHandler()` with HTML/JSON panic recovery
- **Route cache** — `gofreight route:cache`, `GOFREIGHT_ROUTE_CACHE=1`
- **REST `only` / `except`** — `router.ResourceOptions` on `Resources()` / `ApiResource()`
- **Authorization gates** — model-aware `auth.Gate` with `Define`, `Allows`, `RequireGate`
- **Vite integration** — `#vite` GFT directive, dev proxy, `app.UseVite()`
- **Redis WebSocket broadcast** — `app.UseRedisBroadcast()` for multi-instance realtime
- **Auth starter** — `gofreight make:auth` generates login/register views, routes, and controllers

### Added — Documentation

- **CLI commands** — in-depth reference with examples for every `gofreight` command (including `tinker`, migrations, queues, generators)
- **Scheduling, notifications, storage, error handling** — new guides with honest capability notes
- **Testing guide fixes** — correct `Describe(t, ...)` and `app.Draw()` usage

---

## 0.2.0 — September 6, 2026

### Added

- **GraphQL server** — modular schema modules, SDL string definitions (`GQL()`), DataLoader batching, GraphiQL playground, query depth/complexity limits, and production security defaults. See [GraphQL](graphql.md) and the [GraphQL tutorial](tutorial-graphql.md).
- **Real-time WebSockets** — socket.io-style rooms, events, and broadcasts with a TypeScript client (`GofreightSocket`). See [Real-time WebSockets](realtime.md) and the [real-time tutorial](tutorial-realtime.md).
- **Route groups** — prefix, middleware, and nested groups for clean API versioning. See [Routing](routing.md).
- **JWT authentication guard** — protect API routes with bearer tokens. See [Authentication](authentication.md) and the [JWT tutorial](tutorial-auth-jwt.md).
- **Vine schema validation** — declarative request validation for forms and JSON APIs. See [Forms & Validation](forms-validation.md).
- **Production CLI guard** — mutating commands (`migrate`, `db:wipe`, `db:seed`, `make:*`, queue/cache clears, and more) show a red **PRODUCTION ENVIRONMENT** banner and require typing `yes` to continue when `GOFREIGHT_ENV=production`. Pass `--force` to skip the prompt in CI/deploy scripts.
- **Documentation site** — full framework guides, six step-by-step tutorials, and searchable docs (see the `gofreight-web` repository).

### Changed

- **CLI help output** — running `gofreight` with no arguments shows the banner once; command categories use cyan headers, bold command names, and dimmed descriptions.
- **CLI command list** — section heading is now **Available commands** (no duplicate framework title).

### Fixed

- **CLI banner** — removed duplicate Gofreight name, version, and tagline when invoking the root command.
- **Docs site dark mode** — “Start building” CTA button text is readable on the orange banner in dark theme.

---

## 0.1.0 — September 5, 2026

Initial public release — a batteries-included Go web framework you compile to a single binary.

### Added

- **HTTP routing** — RESTful resources (`Resources`, `ApiResource`), middleware pipeline, and named routes.
- **Controllers & GFT templating** — MVC handlers and **Gofreight Templates** (`.gft`) with layouts, partials, and form helpers.
- **ORM** — chainable queries, associations, validations, lifecycle callbacks, soft deletes, pagination, and transactions.
- **Database layer** — migrations, SQL/Go seeders, blueprint DSL, multi-driver support (SQLite, PostgreSQL, MySQL), SQLite by default for new apps.
- **CLI (`gofreight`)** — `new`, `serve`, `dev`, `migrate`, `make:*` generators, queue/cache/config commands, branded welcome banners, and `gofreight list` grouped by namespace.
- **Authentication** — session login, password hashing, API tokens, OAuth helpers, email verification, and password reset tokens.
- **Authorization** — policies and role-based middleware.
- **API resources** — JSON serializers for REST responses.
- **Jobs & queues** — background jobs with Redis queue driver and `queue:work`.
- **Mail** — mailables and SMTP delivery.
- **Cache** — in-memory and Redis stores, HTTP response caching, fragment caching.
- **Sessions** — cookie sessions, flash messages, encrypted cookies.
- **Service container** — dependency injection and service registration.
- **Configuration** — `.env` loading, YAML config files, `APP_KEY` encryption, and structured database config.
- **Testing (`gftest`)** — HTTP test helpers, factories, fakes, and database seeding for tests.
- **Admin dashboard** — local-only database admin at `/admin` in development.
- **Security** — CSRF, CORS, rate limiting, security headers, and maintenance mode (`gofreight down` / `up`).
- **Generators** — scaffold full CRUD resources, models, controllers, migrations, mail, jobs, policies, factories, and auth scaffolding.
- **Example apps** — `demoapp/` and `examples/blog/`.

---

## Upgrade notes

### From 0.3.0 to 0.3.1

1. Update the module version in your app's `go.mod`:

   ```bash
   go get github.com/lsgser/gofreight@v0.3.1
   go mod tidy
   ```

2. Reinstall the CLI:

   ```bash
   go install github.com/lsgser/gofreight/cmd/gofreight@v0.3.1
   ```

3. Optional: run `gofreight make:graphql` to add GraphQL scaffolding to an existing app.

### From 0.2.0 to 0.3.0

1. Update the module version in your app's `go.mod`:

   ```bash
   go get github.com/lsgser/gofreight@v0.3.0
   go mod tidy
   ```

2. Reinstall the CLI:

   ```bash
   go install github.com/lsgser/gofreight/cmd/gofreight@v0.3.0
   ```

3. **New apps** pick up file session/cache drivers, `bootstrap/schedule.go`, and `make:auth` starter automatically.

4. **Existing apps** — optionally add `app.UseExceptionHandler()`, `app.UseVite()`, and Redis broadcast wiring from the [Application wiring](application-wiring.md) guide.

### From 0.1.0 to 0.2.0

1. Update the module version in your app’s `go.mod`:

   ```bash
   go get github.com/lsgser/gofreight@v0.2.0
   go mod tidy
   ```

2. Reinstall the CLI:

   ```bash
   go install github.com/lsgser/gofreight/cmd/gofreight@v0.2.0
   ```

3. **Production deploys** — if you run CLI commands against production databases, add `--force` to non-interactive scripts (e.g. `gofreight migrate --force`) or expect the new confirmation prompt.

4. **Optional** — add GraphQL or WebSockets using the new guides; existing REST and HTML apps continue to work unchanged.
