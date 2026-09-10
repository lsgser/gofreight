<!--
| --------------------------------------------------------------------------
| Tutorial-graphql
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
| Tutorial-graphql
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

<p align="center">
  <img src="assets/gofreight-icon.png" alt="Gofreight" width="48">
</p>

# GraphQL Tutorial

This tutorial builds a posts API with Gofreight GraphQL — schema defined in SDL strings (like `gql` in JavaScript/TypeScript), Go resolvers, DataLoader, and the GraphiQL playground.

## What you'll build

- `User` and `Post` types defined in GraphQL SDL
- Query resolvers: `users`, `user`, `posts`, `post`
- Mutation resolver: `createPost`
- Field resolver: `Post.author` (nested relation via `p.Source`)
- Named and inline **fragments** for reusable field selections
- DataLoader batching for author lookups
- Playground at `/graphql/playground`

## Prerequisites

A running Gofreight app (`gofreight new myapp && gofreight serve`). Add the GraphQL module files below under `app/graphql/`.

## Step 1 — Define schema in SDL

Create `app/graphql/schema.go` with your type definitions as GraphQL strings:

```go
package graphql

import gfgraphql "github.com/lsgser/gofreight/graphql"

var userSDL = gfgraphql.MustGQL(`
    type User {
        id: ID!
        name: String!
        email: String!
    }

    extend type Query {
        users: [User!]!
        user(id: ID!): User
    }
`)

var postSDL = gfgraphql.MustGQL(`
    type Post {
        id: ID!
        title: String!
        body: String
        authorId: ID!
        author: User!
    }

    extend type Query {
        posts: [Post!]!
        post(id: ID!): Post
    }

    extend type Mutation {
        createPost(title: String!, body: String, authorId: ID!): Post!
    }
`)
```

`MustGQL` validates the SDL at startup — invalid schema strings fail immediately, like the `gql` tag in TypeScript.

## Step 2 — In-memory data store

For this tutorial, use a simple in-memory store (swap for the ORM in production):

```go
// app/graphql/store.go
package graphql

import "sync"

var (
    storeMu sync.RWMutex
    users   = []map[string]any{
        {"id": "1", "name": "Ada", "email": "ada@example.com"},
        {"id": "2", "name": "Lin", "email": "lin@example.com"},
    }
    posts = []map[string]any{
        {"id": "1", "title": "Hello GraphQL", "body": "First post.", "authorId": "1"},
        {"id": "2", "title": "SDL schemas", "body": "Define types as strings.", "authorId": "2"},
    }
)

func findUser(id string) map[string]any {
    storeMu.RLock()
    defer storeMu.RUnlock()
    for _, u := range users {
        if u["id"] == id {
            return u
        }
    }
    return nil
}

func findPost(id string) map[string]any {
    storeMu.RLock()
    defer storeMu.RUnlock()
    for _, p := range posts {
        if p["id"] == id {
            return p
        }
    }
    return nil
}
```

## Step 3 — Understand resolver execution

When a client sends this query:

```graphql
query {
  posts {
    title
    author { name }
  }
}
```

Gofreight calls resolvers in order:

1. **`Query.posts`** — root resolver, returns a list of Post objects
2. **`Post.title`** — skipped (value already on each post map)
3. **`Post.author`** — field resolver, receives each post as `p.Source`

```
Query.posts          →  [{ id, title, authorId }, ...]
Post.author (×N)     →  { name, email }  (one per post, batched via DataLoader)
```

Every resolver is a Go function:

```go
func myResolver(p gql.ResolveParams) (any, error)
```

| `ResolveParams` | Purpose |
|-----------------|---------|
| `p.Context` | Request context — auth, DataLoaders, ORM |
| `p.Args` | Arguments (`id`, `title`, …) |
| `p.Source` | Parent object (field resolvers only) |

---

## Step 4 — Write resolver functions

Create `app/graphql/resolvers.go` with **named functions** — one per field that needs custom logic:

```go
// app/graphql/resolvers.go
package graphql

import (
    "context"
    "fmt"

    gql "github.com/graphql-go/graphql"
)

// --- Query resolvers (root — no parent) ---

func listUsers(_ gql.ResolveParams) (any, error) {
    storeMu.RLock()
    defer storeMu.RUnlock()
    out := make([]map[string]any, len(users))
    copy(out, users)
    return out, nil
}

func getUser(p gql.ResolveParams) (any, error) {
    id, _ := p.Args["id"].(string)
    u := findUser(id)
    if u == nil {
        return nil, fmt.Errorf("user %s not found", id)
    }
    return u, nil
}

func listPosts(_ gql.ResolveParams) (any, error) {
    storeMu.RLock()
    defer storeMu.RUnlock()
    out := make([]map[string]any, len(posts))
    copy(out, posts)
    return out, nil
}

func getPost(p gql.ResolveParams) (any, error) {
    id, _ := p.Args["id"].(string)
    post := findPost(id)
    if post == nil {
        return nil, fmt.Errorf("post %s not found", id)
    }
    return post, nil
}

// --- Mutation resolver ---

func createPost(p gql.ResolveParams) (any, error) {
    title, _ := p.Args["title"].(string)
    body, _ := p.Args["body"].(string)
    authorID, _ := p.Args["authorId"].(string)

    if authorID == "" {
        return nil, fmt.Errorf("authorId is required")
    }
    if findUser(authorID) == nil {
        return nil, fmt.Errorf("author %s not found", authorID)
    }

    storeMu.Lock()
    defer storeMu.Unlock()
    id := fmt.Sprintf("%d", len(posts)+1)
    post := map[string]any{
        "id": id, "title": title, "body": body, "authorId": authorID,
    }
    posts = append(posts, post)
    return post, nil
}

// --- Field resolver (nested) ---

func resolvePostAuthor(p gql.ResolveParams) (any, error) {
    // p.Source is the parent Post returned by Query.posts or Query.post
    post, ok := p.Source.(map[string]any)
    if !ok {
        return nil, fmt.Errorf("invalid post source")
    }
    authorID, _ := post["authorId"].(string)
    return loadUser(p.Context, authorID)
}
```

### Bind resolvers to SDL fields

Create `app/graphql/modules.go` to wire functions to schema fields:

```go
// app/graphql/modules.go
package graphql

import (
    gql "github.com/graphql-go/graphql"
    gfgraphql "github.com/lsgser/gofreight/graphql"
)

func userModule() gfgraphql.Module {
    mod, err := gfgraphql.ModuleFromSDL(gfgraphql.SDLModuleConfig{
        ID:  "user",
        SDL: userSDL,
        Resolvers: gfgraphql.SDLResolvers{
            Query: map[string]gql.FieldResolveFn{
                "users": listUsers,
                "user":  getUser,
            },
        },
    })
    if err != nil {
        panic(err)
    }
    return mod
}

func postModule() gfgraphql.Module {
    mod, err := gfgraphql.ModuleFromSDL(gfgraphql.SDLModuleConfig{
        ID:  "post",
        SDL: postSDL,
        Resolvers: gfgraphql.SDLResolvers{
            Query: map[string]gql.FieldResolveFn{
                "posts": listPosts,
                "post":  getPost,
            },
            Mutation: map[string]gql.FieldResolveFn{
                "createPost": createPost,
            },
            TypeFields: map[string]gql.FieldResolveFn{
                "Post.author": resolvePostAuthor,
            },
        },
    })
    if err != nil {
        panic(err)
    }
    return mod
}
```

### Resolver binding reference

| SDL field | Map key | Resolver function |
|-----------|---------|-------------------|
| `extend type Query { users }` | `Query["users"]` | `listUsers` |
| `extend type Query { post(id) }` | `Query["post"]` | `getPost` |
| `extend type Mutation { createPost }` | `Mutation["createPost"]` | `createPost` |
| `type Post { author }` | `TypeFields["Post.author"]` | `resolvePostAuthor` |

Fields like `Post.title` and `Post.id` need **no resolver** when the parent map already contains those keys.

---

## Step 5 — DataLoader for Post.author

When a client requests many posts with `author { name }`, loading each author individually causes N+1 queries. Register a DataLoader:

```go
// app/graphql/loaders.go
package graphql

import (
    "context"

    "github.com/graph-gophers/dataloader/v7"
    gfgraphql "github.com/lsgser/gofreight/graphql"
)

func registerLoaders(reg *gfgraphql.LoaderRegistry) {
    reg.Register("user", func() any {
        return gfgraphql.NewLoader[string, map[string]any](func(ctx context.Context, keys []string) []*dataloader.Result[map[string]any] {
            results := make([]*dataloader.Result[map[string]any], len(keys))
            for i, key := range keys {
                results[i] = &dataloader.Result[map[string]any]{Data: findUser(key)}
            }
            return results
        })
    })
}

func loadUser(ctx context.Context, id string) (map[string]any, error) {
    if loader, ok := gfgraphql.LoaderFromContext[string, map[string]any](ctx, "user"); ok {
        return loader.Load(ctx, id)()
    }
    u := findUser(id)
    if u == nil {
        return nil, fmt.Errorf("user not found")
    }
    return u, nil
}
```

The field resolver `resolvePostAuthor` calls `loadUser`, which uses the loader when available:

```go
func loadUser(ctx context.Context, id string) (map[string]any, error) {
    if loader, ok := gfgraphql.LoaderFromContext[string, map[string]any](ctx, "user"); ok {
        return loader.Load(ctx, id)()
    }
    u := findUser(id)
    if u == nil {
        return nil, fmt.Errorf("user %s not found", id)
    }
    return u, nil
}
```

Without DataLoader, requesting 10 posts with `author { name }` triggers 10 separate user lookups. With DataLoader, they batch into one call.

---

## Step 6 — Mount on the application

Wire everything in `app/graphql/register.go`:

```go
package graphql

import (
    "context"
    "log"
    "net/http"
    "time"

    "myapp/bootstrap" // your app package
    "github.com/lsgser/gofreight/application"
    gfgraphql "github.com/lsgser/gofreight/graphql"
)

func Mount(app *application.Application) {
    fieldLimits := gfgraphql.NewFieldRateLimitRegistry()
    fieldLimits.Set("Mutation", "createPost", gfgraphql.RateLimitRule{
        Limit: 30, Window: time.Minute,
    })

    gqlApp, err := gfgraphql.CreateApplication(gfgraphql.ApplicationConfig{
        Modules: []gfgraphql.Module{
            userModule(),
            postModule(),
        },
        Playground:      true,
        Path:            "/graphql",
        RateLimit:       120,
        FieldRateLimits: fieldLimits,
        Security:        gfgraphql.DefaultSecurity(),
        OnRequest: func(r *http.Request, loaders *gfgraphql.LoaderRegistry) context.Context {
            registerLoaders(loaders)
            return gfgraphql.DefaultOnRequest(r, loaders)
        },
    })
    if err != nil {
        log.Printf("graphql: %v", err)
        return
    }

    app.MountGraphQLApplication("/graphql", gqlApp)
}
```

Call `graphql.Mount(app)` from `bootstrap/app.go`.

---

## Step 7 — Try it in the playground

```bash
gofreight serve
```

Open **http://localhost:5000/graphql/playground**

### List posts with authors

```graphql
query {
  posts {
    id
    title
    author {
      name
      email
    }
  }
}
```

### Create a post

```graphql
mutation {
  createPost(title: "My first mutation", body: "Hello!", authorId: "1") {
    id
    title
  }
}
```

### Test error handling

Request a missing post — the resolver returns an error:

```graphql
query {
  post(id: "999") {
    title
  }
}
```

Response:

```json
{
  "errors": [{ "message": "post 999 not found" }],
  "data": { "post": null }
}
```

### Use fragments to reuse field selections

Fragments let you define a shared set of fields once and spread them into any query — the same pattern as `gql` on the client:

```graphql
fragment PostFields on Post {
  id
  title
  body
}

fragment AuthorFields on User {
  id
  name
  email
}

query PostsWithAuthors {
  posts {
    ...PostFields
    author {
      ...AuthorFields
    }
  }
}
```

Run the same query in the playground. Gofreight expands `...PostFields` and `...AuthorFields` before calling resolvers — `Post.author` still runs `resolvePostAuthor` for each post.

**Inline fragment** — select fields on a specific type without naming the fragment:

```graphql
query {
  posts {
    ... on Post {
      title
      author { name }
    }
  }
}
```

**With variables:**

```graphql
query PostDetail($id: ID!) {
  post(id: $id) {
    ...PostFields
  }
}
```

Query variables (playground **Variables** panel):

```json
{ "id": "1" }
```

Depth and complexity limits account for fields inside fragments — see **[GraphQL: Fragments](graphql.md#fragments)**.

---

## Step 8 — Connect to the ORM (optional)

Replace in-memory helpers with Gofreight models. Each resolver keeps the same signature — only the data layer changes:

```go
// app/graphql/resolvers.go
import (
    "strconv"
    "myapp/app/models"
)

func listPosts(p gql.ResolveParams) (any, error) {
    return models.Posts.Query(p.Context).OrderDesc("created_at").Get()
}

func getPost(p gql.ResolveParams) (any, error) {
    id, _ := p.Args["id"].(string)
    postID, _ := strconv.ParseInt(id, 10, 64)
    return models.Posts.Find(p.Context, postID)
}

func createPost(p gql.ResolveParams) (any, error) {
    post := &models.Post{
        Title:  p.Args["title"].(string),
        Body:   p.Args["body"].(string),
        UserID: parseAuthorID(p.Args["authorId"]),
    }
    if err := post.Save(p.Context); err != nil {
        return nil, err // validation errors propagate to GraphQL errors
    }
    return post, nil
}

func resolvePostAuthor(p gql.ResolveParams) (any, error) {
    post, _ := p.Source.(*models.Post)
    return models.Users.Find(p.Context, post.UserID)
}
```

Use `Query().With("Author")` on list queries for eager loading, or keep DataLoader on `Post.author` for batching.

See **[ORM](orm.md)** for models, queries, and associations.

---

## Step 9 — Protect resolvers with auth (optional)

When GraphQL routes use JWT or session middleware, read the user inside mutations:

```go
import "github.com/lsgser/gofreight/auth"

func createPost(p gql.ResolveParams) (any, error) {
    userID, ok := auth.UserIDFromContext(p.Context)
    if !ok {
        return nil, fmt.Errorf("authentication required")
    }
    // set author from authenticated user instead of authorId arg
    ...
}
```

See **[Authentication](authentication.md)** for login and middleware setup.

---

## Supported types recap

| SDL | Example |
|-----|---------|
| Scalars | `String`, `Int`, `Float`, `Boolean`, `ID` |
| Objects | `type Post { ... }` |
| Enums | `enum Status { DRAFT PUBLISHED }` |
| Inputs | `input CreatePostInput { title: String! }` |
| Lists | `[Post!]!` |
| Root ops | `extend type Query`, `extend type Mutation` |

Interfaces and unions are supported programmatically — see **[GraphQL](graphql.md#supported-graphql-types)**.

## Production checklist

- Set `Playground: false` when `GOFREIGHT_ENV=production`
- Use `gfgraphql.ProductionSecurity()` to disable introspection
- Add `auth.Guard.Middleware` if the endpoint requires login
- Set global and per-mutation rate limits

## Next steps

- **[GraphQL reference](graphql.md)** — resolvers, types, security, DataLoader
- **[Authentication](authentication.md)** — protect your GraphQL endpoint
- **[Real-time WebSockets](realtime.md)** — live updates alongside GraphQL
