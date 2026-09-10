<!--
| --------------------------------------------------------------------------
| Routing
| --------------------------------------------------------------------------
|
| Official Gofreight documentation consumed by the docs site and linked
| from the CLI welcome page.
|
| Keep examples aligned with the current CLI and version; sync copies to
| gofreight-web when publishing.
|
| Covers concepts, tutorials, and reference material for this topic.
|

-->

<!--
| --------------------------------------------------------------------------
| Routing
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Routing

Gofreight provides a **full-featured router** with route groups, named routes, redirects, constraints, model binding, domain routing, signed URLs, file upload/download helpers, and HTTP status helpers. Routes live in `routes/web.go` and `routes/api.go`, wired from `routes/register.go`.

## On this page

- [Basic routes & groups](#basic-routes)
- [REST resources](#restful-resources)
- [Redirects & named URLs](#redirect-routes)
- [File downloads & uploads](#file-downloads--uploads)
- [HTTP status codes](#http-status-codes)
- [Route constraints](#route-constraints)
- [Wildcard & optional params](#wildcard-and-optional-parameters)
- [Domain & subdomain routing](#domain-and-subdomain-routing)
- [Route model binding](#route-model-binding)
- [Signed URLs](#signed-urls)
- [Feature overview](#feature-overview)

---

Gofreight uses **route groups** to share prefixes, middleware, and name prefixes across related routes. Define routes in a callback, then chain `.Prefix()`, `.Use()`, `.Name()`, `.Domain()`, and `.Apply()`.

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
| `Domain(host)` | Restrict group to a host (`api.example.com`, `{tenant}.example.com`) |
| `Subdomain(name)` | Prefix default domain (`api` → `api.example.com`) |
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

---

## Redirect routes

Register URL redirects directly on the router:

```go
r.PermanentRedirect("/old-blog", "/blog")           // 301
r.Redirect("/legacy", "/home", http.StatusFound)    // 302 (default)
```

In controllers you can also redirect imperatively:

```go
base.Redirect("/posts", http.StatusSeeOther)
base.RedirectBack("/")                              // Referer, or fallback
base.RedirectRoute("posts.show", map[string]string{"id": id}, http.StatusSeeOther)
```

`RedirectRoute` and `controller.RouteURL()` use **named routes** — see below.

---

## Named routes & URL generation

Pass a route name as the last argument when registering:

```go
r.Get("/posts/:id", controller.Handler(c.Show), "posts.show")
```

Build URLs from route names:

```go
url, _ := app.Router.URL("posts.show", map[string]string{"id": "42"})
// /posts/42

url, _ := app.Router.URLParams("posts.show", "id", "42")

path, _ := controller.RouteURL("posts.show", map[string]string{"id": "42"})
```

`Resources` and `ApiResource` auto-name routes (`posts.index`, `posts.show`, etc.).

---

## File downloads & uploads

### Download a file

Send files as attachments or inline in the browser:

```go
r.Get("/reports/:id/download", controller.Handler(func(base controller.Base) error {
    return base.Download("storage/reports/report.pdf", "monthly-report.pdf")
}))

r.Get("/images/:file", controller.Handler(func(base controller.Base) error {
    return base.File(filepath.Join("public/images", base.Param("file")))
}))
```

For generated content:

```go
base.DownloadBytes([]byte(sql), "export.sql", "application/sql")
base.StreamDownload(reader, "archive.zip", "application/zip", size)
```

### Upload a file

Handle multipart file uploads:

```go
r.Post("/uploads", controller.Handler(func(base controller.Base) error {
    path, err := base.StoreUpload("file", "storage/uploads", 10) // 10 MB max
    if err != nil {
        return err
    }
    base.RenderJSON(map[string]string{"path": path})
    return nil
}))
```

Lower-level access:

```go
file, header, err := base.UploadedFile("file", 10)
defer file.Close()
path, err := upload.SaveFile(file, header, "storage/uploads")
```

See [Security](security.md) for size limits and filename sanitization.

---

## Additional HTTP verbs

```go
r.Head("/health", headHandler)
r.Options("/api", corsHandler)
r.Match([]string{"GET", "POST"}, "/webhook", webhookHandler)
r.Any("/callback", callbackHandler)   // GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
```

---

## HTTP status codes

Set status codes in **controllers** for dynamic responses, or on **routes** when a path always returns the same code.

### In controllers

Set `base.Status` before rendering, or use semantic helpers:

```go
base.Created(post)
base.NoContent()
base.Head(http.StatusOK)
base.Abort(http.StatusNotFound, "Post not found")
base.AbortNamed("forbidden", "You cannot edit this post")

// Shorthand error responses
base.NotFound("Post not found")       // 404 JSON
base.Unauthorized("Login required")   // 401 JSON
base.Forbidden("Access denied")       // 403
base.Unprocessable(errs)              // 422 validation
base.Conflict("Email already taken")  // 409
base.TooManyRequests("Slow down")     // 429
```

**Symbolic status names** work anywhere you need a named code:

```go
code, ok := controller.StatusFromName("created")     // 201
code, ok := controller.StatusFromName("no-content")  // 204

base.RedirectNamed("/posts", "see_other")            // 303 redirect
```

| Helper | Code | Typical use |
|--------|------|-------------|
| `OK(data)` | 200 | JSON response body |
| `Created(data)` | 201 | Resource created |
| `Accepted(data)` | 202 | Async acceptance |
| `NoContent()` | 204 | Empty success body |
| `Head(code)` | any | Status only, no body |
| `Abort(code, msg)` | any | Error response |
| `Redirect(url, code)` | 3xx | HTTP redirect |

For HTML views, set status then render:

```go
base.SetStatus(http.StatusNotFound)
return base.RenderView("errors/404", nil)
```

See **[Controllers](controllers.md)** for the full response API.

### On routes

Force a fixed status for a route:

```go
r.Get("/deprecated", deprecatedHandler).Status(http.StatusGone)
r.Get("/health/ping", pingHandler).StatusName("no_content")  // 204
```

Route status overrides whatever status the handler writes — useful for static "gone" or empty health-check endpoints.

### Redirect status codes

```go
r.PermanentRedirect("/old", "/new")                    // 301
r.Redirect("/temp", "/new", http.StatusFound)          // 302
base.Redirect("/posts", http.StatusSeeOther)           // 303 (after POST)
base.RedirectNamed("/home", "moved_permanently")       // 301
```

---

## Fallback route

Handle unmatched URLs with a fallback handler:

```go
r.Fallback(controller.Handler(func(base controller.Base) error {
    base.Status = http.StatusNotFound
    return base.RenderView("errors/404", nil)
}))
```

---

## Route constraints

Restrict parameter shapes with regex:

```go
r.Get("/posts/:id", controller.Handler(c.Show), "posts.show").
    Where(map[string]string{"id": "[0-9]+"})

r.Get("/users/:username", handler).
    WhereParam("username", "[A-Za-z0-9_]+")
```

Non-matching paths return 404 — `/posts/abc` won't match when `id` must be numeric.

---

## Wildcard and optional parameters

Gofreight accepts braced `{param}` syntax and colon `:param` syntax:

| Pattern | Meaning | Example match |
|---------|---------|---------------|
| `:id` / `{id}` | Single segment | `/posts/42` |
| `:id?` / `{id?}` | Optional segment | `/posts` or `/posts/42` |
| `:path*` / `{path*}` | Catch-all (rest of path) | `/files/a/b/c` → `path=a/b/c` |

```go
r.Get("/files/{path*}", controller.Handler(func(base controller.Base) error {
    return base.File(filepath.Join("storage", base.Param("path")))
}))

r.Get("/posts/{id?}", controller.Handler(c.IndexOrShow))
```

Build URLs with wildcards:

```go
app.Router.URL("files.show", map[string]string{"path": "docs/guide.md"})
```

---

## Domain and subdomain routing

Restrict routes to a host pattern:

```go
r.SetDefaultDomain("example.com") // from APP_URL in bootstrap

r.Group(func(api *router.Router) {
    api.Get("/status", apiStatus)
}).Subdomain("api").Apply()  // api.example.com

r.Group(func(tenant *router.Router) {
    tenant.Get("/dashboard", tenantDashboard)
}).Domain("{tenant}.example.com").Apply()
```

Parameterized domains expose `host_<name>` path values (e.g. `host_tenant` for `{tenant}`).

---

## Route model binding

Resolve route params to models automatically:

```go
import "myapp/app/models"

router.BindModel(
    r.Get("/posts/:id", controller.Handler(c.Show), "posts.show"),
    &models.Posts,
    "id",
)
```

By slug:

```go
router.BindModelBy(
    r.Get("/posts/:slug", controller.Handler(c.Show), "posts.show"),
    &models.Posts,
    "slug",
    "slug",
)
```

Custom binding:

```go
r.Get("/users/:id", handler).Bind("id", func(req *http.Request) (any, error) {
    return usersService.Find(req.Context(), req.PathValue("id"))
})
```

In controllers:

```go
post, ok := controller.BoundAs[models.Post](base, "id")
if !ok {
    base.NotFound("Post not found")
    return nil
}
```

---

## Signed URLs

Generate tamper-proof, expiring links:

```go
signer := app.URLSigner()

signed, _ := signer.TemporarySignedRoute(app.Router, "invites.accept", time.Hour, map[string]string{
    "token": token,
})

// Or in controllers:
link := base.SignedURL("/invites/accept?token="+token, time.Hour)
```

Protect a route:

```go
r.Get("/invites/accept", controller.Handler(c.Accept)).
    Signed(app.URLSigner())
```

Verify in a handler:

```go
if !base.HasValidSignature() {
    base.Forbidden("Invalid or expired link")
    return nil
}
```

Signed URLs append `?expires=...&signature=...` query params signed with `APP_KEY`.

---

## Feature overview

| Feature | Gofreight |
|---------|-----------|
| Route groups, middleware, REST resources | Yes |
| Declarative redirect routes | Yes — `Redirect`, `PermanentRedirect` |
| Named route URL generation | Yes — `Router.URL`, `RouteURL`, `RedirectRoute` |
| Controller redirects | Yes — `Redirect`, `RedirectBack`, `RedirectRoute` |
| File download / inline file | Yes — `Download`, `File`, `DownloadBytes`, `StreamDownload` |
| File upload | Yes — `UploadedFile`, `StoreUpload` |
| `Any`, `Match`, `Head`, `Options`, `Fallback` | Yes |
| HTTP status helpers (`Created`, `NoContent`, `Abort`, etc.) | Yes |
| Route-level status (`.Status()`, `.StatusName()`) | Yes |
| Symbolic status names (`created`, `no_content`) | Yes |
| Route constraints / `where()` | Yes — `Where()`, `WhereParam()` |
| Route model binding | Yes — `BindModel`, `BindModelBy`, `Bind()` |
| Domain / subdomain routing | Yes — `Domain()`, `Subdomain()` |
| Signed URLs | Yes — `URLSigner`, `.Signed()`, `TemporarySignedRoute` |
| Catch-all / optional params | Yes — `{path*}`, `{id?}`, `:path*` |
