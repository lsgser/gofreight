<!--
| --------------------------------------------------------------------------
| Tutorial-rest-api
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
| Tutorial-rest-api
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

<p align="center">
  <img src="assets/gofreight-icon.png" alt="Gofreight" width="48">
</p>

# Build a REST API

This tutorial shows how to build a JSON REST API in Gofreight using route groups, controllers, and `ApiResource`.

## What you'll build

A versioned API at `/api/v1` with a health check, a posts resource, and group-level middleware.

## Prerequisites

Complete **[Your First App](tutorial-first-app.md)** or start from a fresh scaffold:

```bash
gofreight new blog-api
cd blog-api
go mod tidy
gofreight key:generate
gofreight db:create && gofreight migrate
```

## Step 1 — Understand route registration

API routes live in `routes/api.go`. They are registered inside a group in `routes/register.go`:

```go
func Register(r *router.Router) {
    Web(r)

    r.Group(func(api *router.Router) {
        API(api)
    }).Prefix("/api/v1").Name("api.").Apply()
}
```

Every route inside `API()` is prefixed with `/api/v1` and named with the `api.` prefix.

## Step 2 — Add a health check

In `routes/api.go`:

```go
func API(r *router.Router) {
    r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })
}
```

Test it:

```bash
curl http://localhost:5000/api/v1/health
# {"status":"ok"}
```

## Step 3 — Scaffold a model and controller

```bash
gofreight make:resource Post title:string body:text published:boolean
gofreight migrate
```

This generates a model, migration, controller, and route stubs.

## Step 4 — Register an API resource

Gofreight provides `ApiResource` for JSON CRUD without HTML `new`/`edit` routes:

```go
r.ApiResource("posts", router.ResourceHandlers{
    Index:   controller.Handler(postsController.Index),
    Show:    controller.Handler(postsController.Show),
    Create:  controller.Handler(postsController.Store),
    Update:  controller.Handler(postsController.Update),
    Destroy: controller.Handler(postsController.Destroy),
})
```

This registers:

| Method | Path | Action |
|--------|------|--------|
| GET | `/api/v1/posts` | List posts |
| POST | `/api/v1/posts` | Create post |
| GET | `/api/v1/posts/:id` | Show post |
| PUT/PATCH | `/api/v1/posts/:id` | Update post |
| DELETE | `/api/v1/posts/:id` | Delete post |

## Step 5 — Return JSON from controllers

Use `controller.Handler` and semantic status helpers:

```go
func (c PostController) Index(base controller.Base) error {
    posts, err := models.Posts.Query(base.Request.Context()).Get()
    if err != nil {
        return err
    }
    base.OK(posts)
    return nil
}

func (c PostController) Store(base controller.Base) error {
    // ... validate and save ...
    base.Created(post)
    return nil
}

func (c PostController) Destroy(base controller.Base) error {
    // ... delete ...
    base.NoContent()
    return nil
}
```

Path parameters are available via `base.Param("id")`. Add numeric constraints on routes:

```go
router.BindModel(
    r.Get("/posts/:id", controller.Handler(c.Show), "posts.show").
        WhereParam("id", "[0-9]+"),
    &models.Posts,
    "id",
)

post, _ := controller.BoundAs[models.Post](&base, "id")
```

## Step 6 — Add group middleware

Protect an entire API group with middleware:

```go
r.Group(func(api *router.Router) {
    API(api)
}).Prefix("/api/v1").Use(authMw).Apply()
```

Or attach middleware to a single route:

```go
api.Delete("/posts/:id", postsDestroy).Use(adminMw)
```

Group middleware runs first (outer groups before inner), then route middleware, then the handler.

## Step 7 — Nested version groups

For multiple API versions, nest groups:

```go
r.Group(func(api *router.Router) {
    api.Group(func(v1 *router.Router) {
        v1.Get("/users", usersIndex) // GET /api/v1/users
    }).Prefix("/v1").Apply()

    api.Group(func(v2 *router.Router) {
        v2.Get("/users", usersV2Index) // GET /api/v2/users
    }).Prefix("/v2").Apply()
}).Prefix("/api").Apply()
```

## Testing with curl

```bash
# Create
curl -X POST http://localhost:5000/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{"title":"Hello API","body":"First post","published":true}'

# List
curl http://localhost:5000/api/v1/posts

# Show
curl http://localhost:5000/api/v1/posts/1

# Update
curl -X PUT http://localhost:5000/api/v1/posts/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated","body":"Revised","published":true}'

# Delete
curl -X DELETE http://localhost:5000/api/v1/posts/1
```

## Next steps

- **[JWT Authentication](tutorial-auth-jwt.md)** — require Bearer tokens on protected routes
- **[Testing](../docs/testing.md)** — write API tests with `gftest`
- **[Routing](../docs/routing.md)** — constraints, model binding, signed URLs, status codes
