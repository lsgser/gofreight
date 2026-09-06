<p align="center">
  <img src="assets/gofreight-icon.png" alt="Gofreight" width="96">
</p>

<h1 align="center">Gofreight Documentation</h1>

<p align="center">
  Batteries-included web framework for Go — compile to a single binary.
</p>

**Repository:** [github.com/lsgser/gofreight](https://github.com/lsgser/gofreight)

---

## Prologue

| Guide | Description |
|-------|-------------|
| [Getting Started](getting-started.md) | Install the CLI, create an app, run migrations |
| [Project structure](project-structure.md) | Framework vs application layout |
| [Configuration](configuration.md) | Environment variables and YAML config |

## The Basics

| Guide | Description |
|-------|-------------|
| [Routes & routing](routing.md) | Route groups, API resources, middleware |
| [Controllers](controllers.md) | Request handling, views, JSON responses |
| [Middleware](middleware.md) | HTTP pipeline, CSRF, CORS, rate limiting |
| [ORM](orm.md) | Models, queries, associations, validations |
| [Database](database.md) | Migrations, seeding, blueprint DSL |
| [Templating](templating.md) | Gofreight Templates (GFT) syntax |
| [Forms & validation](forms-validation.md) | Vine schemas, GFT form components, flash errors |

## Digging Deeper

| Guide | Description |
|-------|-------------|
| [Authentication](authentication.md) | Session login, JWT, API tokens, OAuth |
| [Authorization](authorization.md) | Policies and role-based access |
| [Sessions](sessions.md) | Session storage, flash messages, cookies |
| [Mail](mail.md) | Mailables, SMTP, queued delivery |
| [Jobs & Queues](jobs.md) | Background jobs and workers |
| [Cache](cache.md) | Memory, Redis, HTTP caching |
| [Services & Container](services.md) | Business logic and dependency injection |
| [API Resources](api-resources.md) | JSON serializers for API responses |
| [Real-time WebSockets](realtime.md) | Rooms, events, TypeScript client |
| [GraphQL](graphql.md) | Modular schema, DataLoader, playground, security |

## Advanced

| Guide | Description |
|-------|-------------|
| [Features overview](features.md) | Index of all framework capabilities |
| [Security](security.md) | CSRF, headers, rate limiting, production |
| [Testing](testing.md) | HTTP tests with `gftest` |
| [Date & time](datetime.md) | Carbon-style helpers via `support/datetime` |
| [Integrations](integrations.md) | Pluggable mail, storage, cache, and custom APIs |
| [Extending Gofreight](extending.md) | Custom integrations, events, plugins |
| [Admin Dashboard](admin.md) | Local database admin (development only) |
| [Deployment](deployment.md) | Docker, production checklist |
| [CLI commands](commands.md) | Full `gofreight` command reference |
| [Generators & field types](generators.md) | `make:*` commands and `name:type` fields |

## Quick links

- [Main README](../README.md) — overview and API reference
- [Example blog app](../examples/blog/) — full working application
- Admin panel (development): `http://localhost:5000/admin`
