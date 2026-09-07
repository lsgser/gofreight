# Framework features

Overview of built-in Gofreight capabilities. Each topic has a dedicated guide — this page links to them.

## Core MVC

| Feature | Guide |
|---------|-------|
| HTTP routing — groups, constraints, model binding, signed URLs | [Routing](routing.md) |
| Controllers — status codes, files, redirects, JSON/HTML | [Controllers](controllers.md) |
| Middleware pipeline | [Middleware](middleware.md) |
| GFT templating | [Templating](templating.md) |
| Forms & Vine validation | [Forms & Validation](forms-validation.md) |
| Models — structs, repositories, soft deletes | [Models](models.md) |
| ORM — queries, associations, transactions | [ORM](orm.md) |
| Migrations & column types | [Database](database.md), [Column Types](column-types.md) |

### Routing highlights

The router supports these patterns out of the box:

| Capability | API |
|------------|-----|
| Route groups & middleware | `Group().Prefix().Use().Apply()` |
| Named routes & URL generation | `Router.URL()`, `RouteURL()` |
| Redirect routes | `Redirect()`, `PermanentRedirect()` |
| Parameter constraints | `.Where()`, `.WhereParam()` |
| Optional & catch-all params | `{id?}`, `{path*}` |
| Domain / subdomain routing | `.Domain()`, `.Subdomain()` |
| Route model binding | `BindModel()`, `BindModelBy()`, `.Bind()` |
| Signed URLs | `URLSigner`, `.Signed()`, `SignedURL()` |
| File download / upload | `Download()`, `File()`, `StoreUpload()` |
| HTTP status helpers | `Created()`, `NoContent()`, `Abort()` |
| Fallback & extra verbs | `Fallback()`, `Any()`, `Match()` |

## Data layer

| Feature | Guide |
|---------|-------|
| ORM — models, queries, associations | [ORM](orm.md) |
| Migrations, seeding, blueprint DSL | [Database](database.md) |
| API JSON serializers | [API Resources](api-resources.md) |

## Auth & security

| Feature | Guide |
|---------|-------|
| Session login, JWT, API tokens, OAuth | [Authentication](authentication.md) |
| Policies & roles | [Authorization](authorization.md) |
| Sessions & flash messages | [Sessions](sessions.md) |
| CSRF, rate limiting, production hardening | [Security](security.md) |

## Infrastructure

| Feature | Guide |
|---------|-------|
| Background jobs & queues | [Jobs & Queues](jobs.md) |
| Task scheduler | [Scheduling](scheduling.md) |
| Email & mailables | [Mail](mail.md) |
| Notifications (mail + database) | [Notifications](notifications.md) |
| Caching (memory, file, Redis, HTTP) | [Cache](cache.md) |
| Local file storage | [Storage](storage.md) |
| Exception / panic handling | [Error handling](error-handling.md) |
| Localization (i18n) | [Localization](localization.md) |
| Plugin lifecycle hooks | [Plugins](plugins.md) |
| Bootstrap wiring reference | [Application wiring](application-wiring.md) |
| Service container & business logic | [Services & Container](services.md) |
| Configuration & environment | [Configuration](configuration.md) |
| Mail, storage, cache drivers | [Integrations](integrations.md) |

## Real-time & API

| Feature | Guide |
|---------|-------|
| WebSockets & channels | [Real-time WebSockets](realtime.md) |
| Vite dev assets | [Templating](templating.md) — `#vite` directive |
| Modular GraphQL | [GraphQL](graphql.md) |

## Developer tools

| Feature | Guide |
|---------|-------|
| CLI commands | [CLI Commands](commands.md) |
| Code generators | [Generators](generators.md) |
| HTTP & database testing | [Testing](testing.md) |
| Dev database admin | [Admin Dashboard](admin.md) |
| Date/time helpers | [Date & Time](datetime.md) |
| Custom integrations & events | [Extending](extending.md) |
| Docker & production | [Deployment](deployment.md) |

## Quick wiring example

All features integrate via `application.Application`:

```go
app := application.New()
app.ConnectDatabase()
app.UseRedisSessions(os.Getenv("REDIS_URL"))
app.UseRedisQueue(os.Getenv("REDIS_URL"))
app.UseCSRF()
app.UseLocale()
app.LoadLocales("config/locales")
app.MountSocket("/socket")
app.StartJobs(2)
app.Run()
```

## Tutorials

Step-by-step guides on the docs site:

- [Your First App](tutorial-first-app.md) — create and run a project
- [Build a REST API](tutorial-rest-api.md) — JSON endpoints
- [HTML CRUD with GFT](tutorial-html-crud.md) — browser forms
- [JWT Authentication](tutorial-auth-jwt.md) — protected API routes
- [Real-time WebSockets](tutorial-realtime.md) — live chat
- [GraphQL API](tutorial-graphql.md) — SDL schemas, resolvers, DataLoader
