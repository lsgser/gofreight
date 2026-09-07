# Authorization

Authentication confirms **who** the user is; authorization decides **what** they can do. Gofreight provides three complementary tools:

| Tool | Best for |
|------|----------|
| **`auth.Policy`** | Named request-level rules (`func(r *http.Request) bool`) |
| **`auth.Gate`** | Model-aware rules with user ID + action + model |
| **`auth.RequireRole`** | Simple role-based route protection |

---

## Policies

Policies are named boolean rules evaluated against the current request.

### Define policies

```go
policy := auth.NewPolicy()

policy.Define("edit-post", func(r *http.Request) bool {
    userID, ok := auth.UserIDFromRequest(r, "current_user_id")
    if !ok {
        return false
    }
    postID := r.PathValue("id")
    return ownsPost(userID, postID)
})

policy.Define("admin-only", func(r *http.Request) bool {
    role, ok := auth.RoleFromRequest(r)
    return ok && role == "admin"
})
```

### Check in controllers

```go
if !policy.Allows("edit-post", base.Request) {
    base.Unauthorized("Forbidden")
    return nil
}
```

### Middleware

```go
r.Put("/posts/{id}", controller.Handler(c.Update)).
    Use(policy.RequirePolicy("edit-post"))
```

---

## Gates (model-aware authorization)

Gates are model-aware authorization rules that receive the **user ID**, **action name**, and **model instance**.

### Define gates

```go
gate := auth.NewGate()

gate.Define("update-post", func(ctx context.Context, userID int64, action string, model any) bool {
    post, ok := model.(*models.Post)
    if !ok {
        return false
    }
    return post.UserID == userID
})

gate.Define("delete-post", func(ctx context.Context, userID int64, action string, model any) bool {
    post, ok := model.(*models.Post)
    if !ok {
        return false
    }
    role, _ := auth.RoleFromRequest(/* need request in closure */)
    return post.UserID == userID || role == "admin"
})
```

### Check in controllers

```go
post, _ := models.Posts.Find(ctx, id)
if !gate.Allows(base.Request, "current_user_id", "update-post", post) {
    base.Unauthorized("Forbidden")
    return nil
}
```

### Gate middleware

```go
post := &models.Post{ID: 1, UserID: 42}
r.Put("/posts/{id}", controller.Handler(c.Update)).
    Use(gate.RequireGate("update-post", "current_user_id", post))
```

---

## When to use Policy vs Gate

| Scenario | Use |
|----------|-----|
| Route doesn't need a loaded model | **Policy** |
| Authorization depends on model fields | **Gate** |
| Simple role check | **`RequireRole`** |
| API token scope checks | **Policy** or controller check |

---

## Role middleware

Restrict routes to users with a specific role:

```go
r.Group(func(admin *router.Router) {
    admin.Get("/dashboard", controller.Handler(c.Dashboard))
    admin.Get("/users", controller.Handler(c.Users))
}).Use(auth.RequireRole("admin", "current_user_id")).Apply()
```

Rules:

- Role read from JWT claims or session
- Users with role `"admin"` pass **any** role check
- Unauthenticated users receive **401 Unauthorized**
- Authenticated but wrong role receives **403 Forbidden**

---

## Application structure

Store authorization logic in `app/policies/`:

```go
// app/policies/post_policy.go
package policies

import (
    "context"
    "net/http"
    "strconv"

    "myapp/app/models"
    "github.com/lsgser/gofreight/auth"
)

func RegisterPolicy(p *auth.Policy) {
    p.Define("edit-post", func(r *http.Request) bool {
        userID, ok := auth.UserIDFromRequest(r, "current_user_id")
        if !ok {
            return false
        }
        id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
        post, err := models.Posts.Find(r.Context(), id)
        return err == nil && post.UserID == userID
    })
}

func RegisterGate(g *auth.Gate) {
    g.Define("update-post", func(ctx context.Context, userID int64, action string, model any) bool {
        post, ok := model.(*models.Post)
        return ok && post.UserID == userID
    })
}
```

Wire in bootstrap:

```go
policy := auth.NewPolicy()
gate := auth.NewGate()
policies.RegisterPolicy(policy)
policies.RegisterGate(gate)
```

---

## Auth context helpers

| Function | Returns |
|----------|---------|
| `auth.UserIDFromRequest(r, sessionKey)` | User ID from JWT, API token, or session |
| `auth.UserIDFromContext(ctx)` | User ID from request context |
| `auth.RoleFromRequest(r)` | Role from JWT claims |
| `auth.CurrentUserID(r, sessionKey)` | Session-only user ID |

These work regardless of which authentication method the client used.

---

## Controller-level resource checks

For fine-grained control, check in the action:

```go
func (c PostController) Update(base controller.Base) error {
    userID, ok := auth.UserIDFromRequest(base.Request, "current_user_id")
    if !ok {
        base.Unauthorized("Login required")
        return nil
    }

    post, err := models.Posts.Find(base.Request.Context(), parseID(base.Param("id")))
    if err != nil {
        base.NotFound("Post not found")
        return nil
    }
    if post.UserID != userID {
        base.Unauthorized("Forbidden")
        return nil
    }

    // apply updates...
    return base.JSON(post)
}
```

---

## Generating policies

```bash
gofreight make:policy Post
```

Creates `app/policies/post_policy.go` — add your rules and register in bootstrap.

---

## Testing authorization

```go
func TestEditPostForbidden(t *testing.T) {
    app := gftest.NewApp(t, gftest.WithMigrateDir("db/migrate"))
    app.Draw(routes.Register)

    app.PutJSON("/posts/1", map[string]string{"title": "Hacked"}).
        AssertStatus(401)
}
```

Test with authenticated sessions or JWT headers for allowed cases.

---

## Related

- [Authentication](authentication.md) — login, JWT, API tokens
- [Middleware](middleware.md) — applying middleware
- [Security](security.md) — production hardening
- [Generators](generators.md) — `make:policy`
