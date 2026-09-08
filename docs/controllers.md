# Controllers

Controllers handle HTTP requests and return responses — HTML views, JSON, redirects, or errors. Each resource gets a struct with action methods bound via `controller.Handler`.

## Anatomy of a controller

Controllers live in `app/controllers/`. Each action receives a `controller.Base` with the current request and response:

```go
// app/controllers/post_controller.go
package controllers

import (
    "context"
    "net/http"
    "strconv"

    "myapp/app/models"
    "github.com/lsgser/gofreight/controller"
)

type PostController struct{}

func (c PostController) Index(base controller.Base) error {
    posts, err := models.Posts.Query(base.Request.Context()).Get()
    if err != nil {
        return err
    }

    if base.WantsJSON() {
        base.RenderJSON(posts)
        return nil
    }

    return base.RenderView("posts/index", base.ViewData(map[string]any{
        "Posts": posts,
    }))
}

func (c PostController) Show(base controller.Base) error {
    id, _ := strconv.ParseInt(base.Param("id"), 10, 64)
    post, err := models.Posts.Find(base.Request.Context(), id)
    if err != nil {
        base.NotFound("Post not found")
        return nil
    }
    return base.RenderView("posts/show", base.ViewData(map[string]any{"Post": post}))
}
```

Register actions in `routes/web.go`:

```go
c := controllers.PostController{}
r.Resources("posts", router.ResourceHandlers{
    Index:   controller.Handler(c.Index),
    Show:    controller.Handler(c.Show),
    Create:  controller.Handler(c.Create),
    Update:  controller.Handler(c.Update),
    Destroy: controller.Handler(c.Destroy),
})
```

See **[Routing](routing.md)** for `Resources`, `ApiResource`, and route groups.

## controller.Base

| Method | Purpose |
|--------|---------|
| `Param(key)` | Route parameter (`:id` from `/posts/:id`) |
| `Query(key)` | Query string value |
| `RenderView(name, data)` | Render a GFT template |
| `RenderPartial(name, data)` | Render without layout |
| `RenderJSON(data)` | JSON response |
| `RenderText(text)` | Plain text response |
| `Redirect(url, code)` | HTTP redirect |
| `RedirectBack(fallback, code)` | Redirect to Referer or fallback |
| `RedirectRoute(name, params, code)` | Redirect to a named route |
| `Download(path, name)` | Send file as attachment |
| `File(path)` | Send file inline |
| `DownloadBytes(data, name, type)` | Download in-memory content |
| `StreamDownload(r, name, type, size)` | Stream a download |
| `UploadedFile(field, maxMB)` | Parse multipart upload field |
| `StoreUpload(field, destDir, maxMB)` | Save uploaded file to disk |
| `SetStatus(code)` | Set status before `RenderView` / `RenderJSON` |
| `OK(data)` | 200 JSON response |
| `Created(data)` | 201 JSON (or empty 201) |
| `Accepted(data)` | 202 JSON |
| `NoContent()` | 204 empty response |
| `Head(code)` | Status only, no body |
| `Abort(code, msg)` | Stop with error (JSON or HTML) |
| `AbortNamed(name, msg)` | Abort with symbolic status name |
| `Forbidden(msg)` | 403 |
| `Conflict(msg)` | 409 |
| `TooManyRequests(msg)` | 429 |
| `ViewData(extra)` | Merge CSRF, flash errors, old input into view data |
| `WantsJSON()` | True when client expects JSON |
| `ValidateUsing(schema)` | Validate with Vine; auto-handles HTML vs JSON |
| `NotFound(msg)` | 404 JSON response |
| `Unauthorized(msg)` | 401 JSON response |
| `Unprocessable(errors)` | 422 JSON validation errors |
| `Bound(name)` | Route-model binding from context |
| `BoundAs[T](base, name)` | Typed route-model binding |
| `RouteURL(name, params)` | Build URL for a named route |
| `SignedURL(path, ttl)` | Temporary signed relative URL |
| `TemporarySignedRoute(name, ttl, params)` | Signed named route URL |
| `HasValidSignature()` | Verify current request signature |

Set `base.Status` before rendering, or prefer semantic helpers like `Created()`, `NoContent()`, and `Abort()`.

### Status code examples

```go
func (c PostController) Store(base controller.Base) error {
    // ... validate and save ...
    base.Created(post)          // 201 + JSON
    return nil
}

func (c PostController) Destroy(base controller.Base) error {
    // ... delete ...
    base.NoContent()            // 204
    return nil
}

func (c PostController) Show(base controller.Base) error {
    post, err := models.Posts.Find(ctx, id)
    base.AbortIf(err != nil, http.StatusNotFound, "Post not found")
    return base.RenderView("posts/show", base.ViewData(map[string]any{"Post": post}))
}
```

Use `controller.StatusFromName("created")` for symbolic status codes. See **[Routing — HTTP status codes](routing.md#http-status-codes)**.

## Route model binding

When a route resolves a parameter to a database record, use **implicit binding** instead of manual lookups:

```go
// routes/web.go
router.BindModel(
    r.Get("/posts/:id", controller.Handler(c.Show), "posts.show"),
    &models.Posts,
    "id",
)

router.BindModelBy(
    r.Get("/posts/:slug", controller.Handler(c.ShowBySlug), "posts.by_slug"),
    &models.Posts,
    "slug",
    "slug",
)
```

In the controller:

```go
func (c PostController) Show(base controller.Base) error {
    post, ok := controller.BoundAs[models.Post](&base, "id")
    if !ok {
        base.NotFound("Post not found")
        return nil
    }
    return base.RenderView("posts/show", base.ViewData(map[string]any{"Post": post}))
}
```

Custom bindings with `.Bind("param", func(req *http.Request) (any, error) { ... })`. See **[Routing — Route model binding](routing.md#route-model-binding)**.

## Named routes & signed URLs

Build URLs from route names:

```go
url, _ := controller.RouteURL("posts.show", map[string]string{"id": "42"})
base.RedirectRoute("posts.show", map[string]string{"id": id}, http.StatusSeeOther)
```

Temporary signed links for password resets, email verification, or secure downloads:

```go
link := base.SignedURL("/invites/accept?token="+token, time.Hour)
signed, _ := base.TemporarySignedRoute("invites.accept", time.Hour, map[string]string{"token": token})

if !base.HasValidSignature() {
    base.Forbidden("Invalid or expired link")
    return nil
}
```

Protect routes with `.Signed(app.URLSigner())`. See **[Routing — Signed URLs](routing.md#signed-urls)** and **[Security](security.md)**.

## File responses

Send downloads and handle uploads from controllers:

```go
return base.Download("storage/reports/monthly.pdf", "report.pdf")
return base.File("public/images/logo.png")

path, err := base.StoreUpload("avatar", "storage/uploads", 10) // 10 MB max
```

See **[Routing — File downloads & uploads](routing.md#file-downloads--uploads)**.

## Handler wrapper

`controller.Handler` converts an action method into `http.HandlerFunc`:

```go
r.Get("/health", controller.Handler(func(base controller.Base) error {
    base.RenderJSON(map[string]string{"status": "ok"})
    return nil
}))
```

Return a non-nil `error` for unexpected failures (500). Use response helpers (`NotFound`, `Unprocessable`) for expected client errors.

## HTML vs JSON

Many apps serve both browser and API clients from the same controller:

```go
func (c PostController) Index(base controller.Base) error {
    posts, err := models.Posts.Query(base.Request.Context()).Get()
    if err != nil {
        return err
    }
    if base.WantsJSON() {
        base.RenderJSON(posts)
        return nil
    }
    return base.RenderView("posts/index", base.ViewData(map[string]any{"Posts": posts}))
}
```

`WantsJSON()` checks the `Accept` header and `Content-Type`. HTML forms send `Accept: text/html`; API clients send `application/json`.

## Validation in controllers

Use Vine schemas for request validation:

```go
var postSchema = vine.Object(map[string]vine.Rule{
    "title": vine.String().Required().MinLength(3).MaxLength(255),
    "body":  vine.String().Required(),
})

func (c PostController) Store(base controller.Base) error {
    data, err := base.ValidateUsing(postSchema)
    if err != nil {
        return nil // ValidateUsing already responded
    }
    // create from data...
    base.Redirect("/posts", http.StatusSeeOther)
    return nil
}
```

On failure, HTML requests redirect back with flashed errors; JSON requests get 422. See **[Forms & Validation](forms-validation.md)**.

## View data and flash

`ViewData` injects request-scoped template helpers:

```go
base.RenderView("posts/edit", base.ViewData(map[string]any{
    "Post": post,
}))
```

Templates receive CSRF tokens, validation errors (`{{ error "title" }}`), and old input automatically.

## REST resource actions

| Action | Method | Path | Typical use |
|--------|--------|------|-------------|
| `Index` | GET | `/posts` | List all |
| `Show` | GET | `/posts/:id` | Single record |
| `Create` | POST | `/posts` | Store new record |
| `Update` | PUT/PATCH | `/posts/:id` | Update record |
| `Destroy` | DELETE | `/posts/:id` | Delete record |

HTML resources also register `New` and `Edit` form pages. API resources (`ApiResource`) omit those.

## Scaffolding

Generate a full controller with views, model, and migration:

```bash
gofreight make:scaffold Post title:string body:text
gofreight make:controller Comment
gofreight make:api Post title:string body:text
```

## Keep controllers thin

Move business logic to `app/services/`:

```go
// app/services/post_service.go
func PublishPost(ctx context.Context, id int64) error { ... }

// controller
if err := services.PublishPost(base.Request.Context(), id); err != nil { ... }
```

See **[Services & Container](services.md)**.

## API JSON resources

Transform models into consistent JSON with `app/resources/` and the `api` package. See **[API Resources](api-resources.md)**.

## Reference app

The [blog example on GitHub](https://github.com/lsgser/gofreight/tree/main/examples/blog/app/controllers) shows dual HTML/JSON controllers with associations and validation.

## Related

- [Routing](routing.md) — route registration and middleware
- [Middleware](middleware.md) — CSRF, sessions, auth
- [ORM](orm.md) — models and queries
- [Templating](templating.md) — GFT views
- [Tutorial: HTML CRUD](tutorial-html-crud.md) — step-by-step CRUD with GFT
