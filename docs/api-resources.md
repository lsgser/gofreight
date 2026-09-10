<!--
| --------------------------------------------------------------------------
| Api-resources
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
| API Resources
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# API Resources

API resources transform models into consistent JSON responses for your API.

## Basic usage

Implement the `api.Resource` interface:

```go
// app/resources/post_resource.go
package resources

import "myapp/app/models"

type PostResource struct {
    Post *models.Post
}

func (r PostResource) ToMap() map[string]any {
    return map[string]any{
        "id":         r.Post.ID,
        "title":      r.Post.Title,
        "body":       r.Post.Body,
        "created_at": r.Post.CreatedAt,
    }
}
```

Render from a controller:

```go
import "github.com/lsgser/gofreight/api"

post, _ := models.Posts.Find(ctx, id)
api.Render(w, http.StatusOK, resources.PostResource{Post: post})
```

Response shape:

```json
{
  "data": {
    "id": 1,
    "title": "Hello",
    "body": "World",
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

## Collections

Render multiple resources:

```go
posts, _ := models.Posts.Query(ctx).Get()

items := make([]api.Resource, len(posts))
for i, p := range posts {
    items[i] = resources.PostResource{Post: &p}
}
api.RenderMany(w, http.StatusOK, items)
```

Or build a collection manually:

```go
data := make([]map[string]any, len(posts))
for i, p := range posts {
    data[i] = resources.PostResource{Post: &p}.ToMap()
}

api.RenderCollection(w, http.StatusOK, api.Collection{
    Data: data,
    Meta: api.PaginatedMeta(page, perPage, total),
})
```

## Function adapter

For simple transforms, use `api.ResourceFunc`:

```go
api.Render(w, http.StatusOK, api.ResourceFunc(func() map[string]any {
    return map[string]any{"status": "ok"}
}))
```

## Pagination meta

```go
meta := api.PaginatedMeta(currentPage, perPage, totalCount)
// { "current_page": 1, "per_page": 20, "total": 100, "last_page": 5 }
```

Combine with ORM pagination:

```go
page, _ := models.Posts.Query(ctx).Paginate(1, 20)

items := make([]api.Resource, len(page.Data))
for i, p := range page.Data {
    items[i] = resources.PostResource{Post: &p}
}

api.RenderCollection(w, http.StatusOK, api.Collection{
    Data:  /* mapped items */,
    Meta:  api.PaginatedMeta(page.CurrentPage, page.PerPage, int(page.Total)),
})
```

## Scaffolding

Generate an API resource with model and routes:

```bash
gofreight make:api Post title:string body:text
```

Creates model, controller, migration, and JSON routes under `/api/v1`.

## Related

- [Controllers](controllers.md) — returning JSON from actions
- [Routing](routing.md) — `ApiResource` routes
- [ORM](orm.md) — querying models
- [Tutorial: REST API](tutorial-rest-api.md) — end-to-end API (web: `tutorial-rest-api.md`)
