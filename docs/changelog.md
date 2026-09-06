# Changelog

All notable changes to Gofreight are documented here. The project follows [Semantic Versioning](https://semver.org/).

Install a specific release:

```bash
go install github.com/lsgser/gofreight/cmd/gofreight@v0.2.0
```

---

## 0.2.0 — September 6, 2026

### Added

- **GraphQL server** — modular schema modules, SDL string definitions (`GQL()`), DataLoader batching, GraphiQL playground, query depth/complexity limits, and production security defaults. See [GraphQL](graphql.md) and the [GraphQL tutorial](tutorial-graphql.md).
- **Real-time WebSockets** — socket.io-style rooms, events, and broadcasts with a TypeScript client (`GofreightSocket`). See [Real-time WebSockets](realtime.md) and the [real-time tutorial](tutorial-realtime.md).
- **Route groups** — prefix, middleware, and nested groups for clean API versioning. See [Routing](routing.md).
- **JWT authentication guard** — protect API routes with bearer tokens. See [Authentication](authentication.md) and the [JWT tutorial](tutorial-auth-jwt.md).
- **Vine schema validation** — declarative request validation for forms and JSON APIs. See [Forms & Validation](forms-validation.md).
- **Production CLI guard** — mutating commands (`migrate`, `db:wipe`, `db:seed`, `make:*`, queue/cache clears, and more) show a red **PRODUCTION ENVIRONMENT** banner and require typing `yes` to continue when `GOFREIGHT_ENV=production`. Pass `--force` to skip the prompt in CI/deploy scripts.
- **Documentation site** — full framework guides, six step-by-step tutorials, and searchable docs.

### Changed

- **CLI help output** — running `gofreight` with no arguments shows the banner once; command categories use cyan headers, bold command names, and dimmed descriptions.
- **CLI command list** — section heading is now **Available commands** (no duplicate framework title).

### Fixed

- **CLI banner** — removed duplicate Gofreight name, version, and tagline when invoking the root command.

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
