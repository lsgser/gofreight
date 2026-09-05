# Routing

Gofreight uses **route groups** to share prefixes, middleware, and name prefixes across related routes. Define routes in a callback, then chain `.Prefix()`, `.Use()`, `.Name()`, and `.Apply()`.

## Basic routes

```go
r.Get("/", handler)
r.Post("/posts", handler)
r.Get("/posts/:id", handler)  // :id → req.PathValue("id")
```

Paths may omit the leading slash (`"posts"` and `"/posts"` both work).

## Route groups

```go
r.Group(func(api *router.Router) {
    api.Get("/users", usersIndex)
    api.Post("/users", usersStore)
}).Prefix("/api/v1").Use(authMw).Name("api.").Apply()
```

| Method | Purpose |
|--------|---------|
| `Group(fn)` | Start a group — routes defined inside `fn` |
| `Prefix(path)` | URI prefix for all routes in the group |
| `Use(mw...)` / `Middleware(mw...)` | Middleware for routes in the group |
| `Name(prefix)` | Prefix for route names (e.g. `api.users.index`) |
| `Apply()` | Register the group on the parent router |

### Nested groups

Prefixes stack from outer to inner:

```go
r.Group(func(api *router.Router) {
    api.Group(func(v1 *router.Router) {
        v1.Get("/users", handler)    // GET /api/v1/users
        v1.Get("/payments", handler) // GET /api/v1/payments
    }).Prefix("/v1").Apply()
}).Prefix("/api").Apply()
```

### Route-level middleware

Attach middleware to a single route after registration:

```go
r.Group(func(api *router.Router) {
    api.Get("/posts", postsIndex)
    api.Delete("/posts/:id", postsDestroy).Use(adminMw)
}).Prefix("/api").Use(authMw).Apply()
```

Group middleware runs first (outer groups before inner), then route middleware, then the handler.

### Legacy shorthand

```go
r.GroupPrefix("/api/v1", API)  // same as Group(API).Prefix("/api/v1").Apply()
```

## Web vs API layout

```go
// routes/register.go
func Register(r *router.Router) {
    Web(r)

    r.Group(func(api *router.Router) {
        API(api)
    }).Prefix("/api/v1").Name("api.").Apply()
}
```

Or use the helper:

```go
import "github.com/lsgser/gofreight/api"

api.Group(r, "v1", func(api *router.Router) {
    api.Get("/health", healthHandler)
}, auth.APITokenMiddleware(store))
```

## RESTful resources

### Web (HTML) — includes `new` and `edit`

```go
r.Resources("posts", router.ResourceHandlers{
    Index:   controller.Handler(c.Index),
    Show:    controller.Handler(c.Show),
    Create:  controller.Handler(c.Create),
    Update:  controller.Handler(c.Update),
    Destroy: controller.Handler(c.Destroy),
})
```

### API (JSON) — no `new`/`edit`

```go
api.ApiResource("posts", router.ApiResourceHandlers{
    Index:   controller.Handler(c.Index),
    Store:   controller.Handler(c.Store),
    Show:    controller.Handler(c.Show),
    Update:  controller.Handler(c.Update),
    Destroy: controller.Handler(c.Destroy),
})
```

| Method | Path | Action |
|--------|------|--------|
| GET | `/posts` | index |
| POST | `/posts` | store |
| GET | `/posts/:id` | show |
| PUT/PATCH | `/posts/:id` | update |
| DELETE | `/posts/:id` | destroy |

## Middleware order

1. Router-global middleware (`r.Use(...)`)
2. Group middleware (outer groups first, then inner)
3. Route middleware (`.Use(...)` on a single route)
4. Route handler

## List routes

```bash
gofreight route:list
```
