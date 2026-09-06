# Authorization

Authentication confirms who the user is; authorization decides what they can do. Gofreight provides policies (named rules) and role-based middleware.

## Policies

Policies are named authorization rules:

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

Check a policy in a controller:

```go
if !policy.Allows("edit-post", base.Request) {
    base.Unauthorized("Forbidden")
    return nil
}
```

Or enforce via middleware:

```go
r.Put("/posts/{id}", controller.Handler(c.Update)).
    Use(policy.RequirePolicy("edit-post"))
```

## Role middleware

Restrict routes to users with a specific role:

```go
r.Group(func(admin *router.Router) {
    admin.Get("/dashboard", dashboardHandler)
}).Use(auth.RequireRole("admin", "current_user_id")).Apply()
```

Roles are read from JWT claims or session. Users with role `"admin"` pass any role check.

## Application policies

Store policies in `app/policies/`:

```go
// app/policies/post_policy.go
package policies

import (
    "net/http"

    "github.com/lsgser/gofreight/auth"
)

func Register(p *auth.Policy) {
    p.Define("edit-post", func(r *http.Request) bool {
        userID, ok := auth.UserIDFromRequest(r, "current_user_id")
        if !ok {
            return false
        }
        // ownership check...
        return true
    })
}
```

Wire in bootstrap:

```go
policy := auth.NewPolicy()
policies.Register(policy)
app.Router.Use(policy.RequirePolicy("edit-post")) // or per-route
```

## Auth context helpers

| Function | Returns |
|----------|---------|
| `auth.UserIDFromRequest(r, sessionKey)` | User ID from JWT, API token, or session |
| `auth.UserIDFromContext(ctx)` | User ID from request context (API token/JWT) |
| `auth.RoleFromRequest(r)` | Role from JWT claims |

These work regardless of which authentication method the client used.

## Controller-level checks

For resource-specific authorization, check in the action:

```go
func (c PostController) Update(base controller.Base) error {
    userID, ok := auth.UserIDFromRequest(base.Request, "current_user_id")
    if !ok {
        base.Unauthorized("Login required")
        return nil
    }

    post, err := models.Posts.Find(base.Request.Context(), parseID(base.Param("id")))
    if err != nil || post.UserID != userID {
        base.Unauthorized("Forbidden")
        return nil
    }
    // update...
}
```

## Related

- [Authentication](authentication.md) — login, JWT, API tokens
- [Middleware](middleware.md) — applying policy middleware
- [Security](security.md) — production auth hardening
