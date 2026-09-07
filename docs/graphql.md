# GraphQL

Gofreight includes a modular GraphQL server with reusable schema modules, SDL string definitions (like `gql` in JavaScript/TypeScript), DataLoader support, GraphiQL playground, and layered security limits.

## Quick start

### Option A — CLI generators (fastest)

```bash
gofreight make:graphql
gofreight make:graphql-module Post title:string body:text
gofreight make:graphql-module User name:string email:email
go mod tidy
gofreight serve
# → http://localhost:5000/graphql/playground
```

This scaffolds `graphql/register.go`, `graphql/modules.go`, and per-resource module files with list/show queries, create mutations, and dataloaders. See [Generators](generators.md#graphql).

### Option B — SDL strings (manual)

Define your schema in GraphQL SDL and bind Go resolvers:

```go
import (
    gql "github.com/graphql-go/graphql"
    gfgraphql "github.com/lsgser/gofreight/graphql"
)

const postSchema = `
    type Post {
        id: ID!
        title: String!
        body: String
    }

    extend type Query {
        posts: [Post!]!
        post(id: ID!): Post
    }

    extend type Mutation {
        createPost(title: String!, body: String): Post!
    }
`

func PostModule() gfgraphql.Module {
    sdl, _ := gfgraphql.GQL(postSchema)

    mod, err := gfgraphql.ModuleFromSDL(gfgraphql.SDLModuleConfig{
        ID:          "post",
        Description: "Post queries and mutations",
        SDL:         sdl,
        Resolvers: gfgraphql.SDLResolvers{
            Query: map[string]gql.FieldResolveFn{
                "posts": listPosts,
                "post":  getPost,
            },
            Mutation: map[string]gql.FieldResolveFn{
                "createPost": createPost,
            },
            TypeFields: map[string]gql.FieldResolveFn{
                // "Post.author": resolvePostAuthor,
            },
        },
    })
    if err != nil {
        panic(err)
    }
    return mod
}
```

`GQL()` validates SDL at init time — the same workflow as the `gql` template tag in Apollo or graphql-tag.

### Option B — programmatic types

Build types with `graphql-go` directly when you need interfaces, unions, or dynamic schemas:

```go
postType := gql.NewObject(gql.ObjectConfig{
    Name: "Post",
    Fields: gql.Fields{
        "id":    &gql.Field{Type: gql.NewNonNull(gql.ID)},
        "title": &gql.Field{Type: gql.NewNonNull(gql.String)},
    },
})

return gfgraphql.MustCreateModule(gfgraphql.ModuleConfig{
    ID:    "post",
    Types: []gql.Type{postType},
    Query: gql.Fields{
        "posts": &gql.Field{Type: gql.NewList(postType), Resolve: listPosts},
    },
})
```

See `demoapp/graphql/modules.go` for a full programmatic example.

### Mount the application

```go
gqlApp, err := gfgraphql.CreateApplication(gfgraphql.ApplicationConfig{
    Modules: []gfgraphql.Module{
        PostModule(),
        UserModule(),
    },
    Playground: true,
    RateLimit:  120,
    Security:   gfgraphql.DefaultSecurity(),
    OnRequest: func(r *http.Request, loaders *gfgraphql.LoaderRegistry) context.Context {
        registerLoaders(loaders)
        return gfgraphql.DefaultOnRequest(r, loaders)
    },
})

gqlApp.Mount(app.Router, "/graphql")
// Or: app.MountGraphQLApplication("/graphql", gqlApp)
```

### Endpoints

| URL | Purpose |
|-----|---------|
| `POST /graphql` | Execute queries and mutations |
| `GET /graphql/playground` | GraphiQL playground (development) |
| `GET /graphql/docs` | Plain-text module index |

---

## SDL schema definitions

Write schema types as GraphQL SDL strings. Gofreight parses them with the same spec parser used across the GraphQL ecosystem.

### The GQL helper

```go
// Validates at package init — panics if SDL is invalid
var schema = gfgraphql.MustGQL(`
    type User {
        id: ID!
        name: String!
    }
`)

// Or return an error
sdl, err := gfgraphql.GQL(`
    extend type Query {
        users: [User!]!
    }
`)
```

### Resolver bindings

SDL defines **what** the API looks like; Go functions define **how** fields resolve:

| Map key | Binds to |
|---------|----------|
| `Resolvers.Query["posts"]` | `extend type Query { posts: ... }` |
| `Resolvers.Mutation["createPost"]` | `extend type Mutation { createPost: ... }` |
| `Resolvers.Subscription["onMessage"]` | `extend type Subscription { ... }` |
| `Resolvers.TypeFields["Post.author"]` | `type Post { author: User }` field resolver |

Every root field declared in SDL must have a matching resolver function.

---

## Resolvers

A **resolver** is a Go function that fetches the data for a single GraphQL field. Gofreight uses [graphql-go](https://github.com/graphql-go/graphql) resolver signatures throughout.

### Resolver function signature

```go
import gql "github.com/graphql-go/graphql"

// FieldResolveFn — return (value, error)
func listPosts(p gql.ResolveParams) (any, error) {
    return posts, nil
}
```

Bind resolvers when building modules:

| Approach | How |
|----------|-----|
| **SDL module** | `SDLResolvers.Query`, `.Mutation`, `.TypeFields` maps |
| **Programmatic module** | `Resolve:` on each `gql.Field` |

### ResolveParams

Every resolver receives `gql.ResolveParams`:

| Field | Type | Used for |
|-------|------|----------|
| `p.Context` | `context.Context` | Request context, auth, DataLoaders, deadlines |
| `p.Args` | `map[string]any` | Field arguments (`id`, `title`, input objects) |
| `p.Source` | `any` | Parent object — **field resolvers only** |
| `p.Info` | `graphql.ResolveInfo` | Field name, parent type, return type |

```go
func getPost(p gql.ResolveParams) (any, error) {
    // Argument from: post(id: "1")
    id, _ := p.Args["id"].(string)

    post := findPost(id)
    if post == nil {
        return nil, fmt.Errorf("post not found")
    }
    return post, nil
}
```

### Root resolvers (Query / Mutation)

Root resolvers start a query tree. They receive **no parent** — `p.Source` is nil.

```go
// SDL binding
Resolvers: gfgraphql.SDLResolvers{
    Query: map[string]gql.FieldResolveFn{
        "posts": listPosts,
        "post":  getPost,
    },
    Mutation: map[string]gql.FieldResolveFn{
        "createPost": createPost,
    },
}

// Programmatic binding
Query: gql.Fields{
    "posts": &gql.Field{
        Type:    gql.NewList(postType),
        Resolve: listPosts,
    },
}
```

**Query resolver** — return a single object, a list, or a scalar:

```go
func listPosts(_ gql.ResolveParams) (any, error) {
    return []map[string]any{
        {"id": "1", "title": "Hello"},
    }, nil
}
```

**Mutation resolver** — read args, persist, return the changed object:

```go
func createPost(p gql.ResolveParams) (any, error) {
    title, _ := p.Args["title"].(string)
    body, _ := p.Args["body"].(string)
    authorID, _ := p.Args["authorId"].(string)

    post := savePost(title, body, authorID)
    return post, nil
}
```

### Field resolvers (nested objects)

When a client requests nested data, GraphQL calls a **field resolver** for each nested field that needs custom logic:

```graphql
query {
  posts {
    title
    author { name }   # triggers Post.author resolver
  }
}
```

The parent `Post` is passed as `p.Source`:

```go
func resolvePostAuthor(p gql.ResolveParams) (any, error) {
    post, _ := p.Source.(map[string]any)
    authorID, _ := post["authorId"].(string)
    return loadUser(p.Context, authorID)
}

// SDL: TypeFields["Post.author"]
// Programmatic: Resolve on the "author" gql.Field
```

**When you need a field resolver:**

| Situation | Resolver needed? |
|-----------|------------------|
| Field value is already on the parent object (same key name) | No — graphql-go reads it from the map/struct |
| Field is a relation (author, comments, tags) | Yes |
| Field needs computation (fullName, isPublished) | Yes |
| Field uses DataLoader batching | Yes |

For `map[string]any` parents, keys like `"title"` and `"id"` resolve automatically. Relations like `"author"` need an explicit resolver unless the map already contains a nested `"author"` value.

### Returning data

Resolvers can return:

```go
// map — flexible, common in tutorials
return map[string]any{"id": "1", "title": "Hello"}, nil

// struct — typed models from the ORM
return &models.Post{Title: "Hello"}, nil

// slice
return []map[string]any{...}, nil

// scalar
return "ok", nil

// nullable field — return nil, nil
return nil, nil
```

GraphQL serializes struct fields using their Go names (export `json` tags if you customize serialization via maps).

### Reading arguments

Arguments from the GraphQL query arrive in `p.Args`:

```graphql
query {
  post(id: "42") { title }
}

mutation {
  createPost(title: "Hi", body: "Text", authorId: "1") { id }
}
```

```go
id, _ := p.Args["id"].(string)
title, _ := p.Args["title"].(string)

// Optional arg — may be nil if omitted
if body, ok := p.Args["body"].(string); ok {
    post.Body = body
}
```

For input object arguments:

```graphql
mutation {
  createPost(input: { title: "Hi", authorId: "1" }) { id }
}
```

```go
input, _ := p.Args["input"].(map[string]any)
title, _ := input["title"].(string)
```

### Errors

Return an error to surface a GraphQL error response:

```go
func getPost(p gql.ResolveParams) (any, error) {
    id, _ := p.Args["id"].(string)
    post := findPost(id)
    if post == nil {
        return nil, fmt.Errorf("post %s not found", id)
    }
    return post, nil
}
```

The client receives:

```json
{
  "errors": [{ "message": "post 99 not found" }],
  "data": { "post": null }
}
```

Use errors for exceptional cases. For expected "not found" on nullable fields, `return nil, nil` is also valid.

### Context, auth, and DataLoaders

`p.Context` carries per-request data. Register loaders and session in `OnRequest`:

```go
OnRequest: func(r *http.Request, loaders *gfgraphql.LoaderRegistry) context.Context {
    registerLoaders(loaders)
    return gfgraphql.DefaultOnRequest(r, loaders) // loaders + session + client IP
},
```

**Auth in resolvers** (when the route uses `auth.Guard.Middleware` or JWT middleware):

```go
import "github.com/lsgser/gofreight/auth"

func createPost(p gql.ResolveParams) (any, error) {
    userID, ok := auth.UserIDFromContext(p.Context)
    if !ok {
        return nil, fmt.Errorf("authentication required")
    }
    // use userID as authorId...
}
```

**DataLoader in field resolvers:**

```go
func resolvePostAuthor(p gql.ResolveParams) (any, error) {
    post, _ := p.Source.(map[string]any)
    authorID, _ := post["authorId"].(string)

    loader, ok := gfgraphql.LoaderFromContext[string, map[string]any](p.Context, "user")
    if !ok {
        return findUser(authorID), nil
    }
    return loader.Load(p.Context, authorID)()
}
```

### Resolvers with the ORM

Load data in resolvers using Gofreight models:

```go
import (
    "context"
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
        Title: p.Args["title"].(string),
        Body:  p.Args["body"].(string),
    }
    if err := models.Posts.Create(p.Context, post); err != nil {
        return nil, err
    }
    return post, nil
}

func resolvePostAuthor(p gql.ResolveParams) (any, error) {
    post, _ := p.Source.(*models.Post)
    return models.Users.Find(p.Context, post.UserID)
}
```

Use `Query().With("Comments")` for eager loading at the list level, or DataLoaders for nested fields — see **[ORM](orm.md)**.

### Organizing resolver code

Keep schema and resolver wiring separate:

```
app/graphql/
├── schema.go      # SDL strings (MustGQL)
├── resolvers.go   # resolver functions
├── loaders.go     # DataLoader batch functions
├── modules.go     # ModuleFromSDL + bind resolvers
└── register.go    # CreateApplication + Mount
```

Named functions stay testable and reusable:

```go
// resolvers.go
func listPosts(p gql.ResolveParams) (any, error) { ... }
func getPost(p gql.ResolveParams) (any, error) { ... }
func createPost(p gql.ResolveParams) (any, error) { ... }
func resolvePostAuthor(p gql.ResolveParams) (any, error) { ... }

// modules.go
Resolvers: gfgraphql.SDLResolvers{
    Query:      map[string]gql.FieldResolveFn{"posts": listPosts, "post": getPost},
    Mutation:   map[string]gql.FieldResolveFn{"createPost": createPost},
    TypeFields: map[string]gql.FieldResolveFn{"Post.author": resolvePostAuthor},
},
```

### Rate-limited resolvers

Wrap sensitive mutations with per-field rate limits:

```go
"createPost": gfgraphql.Field(gfgraphql.FieldConfig{
    Type:    postType,
    Args:    createPostArgs,
    Resolve: createPost,
}, gfgraphql.WithFieldRateLimit(30, time.Minute)),
```

### Multi-module SDL

Split schema across modules — each module owns its types and root fields:

```go
userMod, _ := gfgraphql.ModuleFromSDL(gfgraphql.SDLModuleConfig{
    ID:  "user",
    SDL: userSDL,
    Resolvers: gfgraphql.SDLResolvers{
        Query: map[string]gql.FieldResolveFn{"users": listUsers},
    },
})

postMod, _ := gfgraphql.ModuleFromSDL(gfgraphql.SDLModuleConfig{
    ID:  "post",
    SDL: postSDL,
    Resolvers: gfgraphql.SDLResolvers{ ... },
})
```

Rules when combining modules:

- Each module requires a unique `ID`
- Duplicate `Query` / `Mutation` fields across modules are rejected
- Types with the same name must be identical

---

## Supported GraphQL types

Split schema across modules — each module owns its types and root fields:

```go
// user_module.go
userMod, _ := gfgraphql.ModuleFromSDL(gfgraphql.SDLModuleConfig{
    ID:  "user",
    SDL: userSDL,
    Resolvers: gfgraphql.SDLResolvers{
        Query: map[string]gql.FieldResolveFn{
            "users": listUsers,
        },
    },
})

// post_module.go
postMod, _ := gfgraphql.ModuleFromSDL(gfgraphql.SDLModuleConfig{
    ID:  "post",
    SDL: postSDL,
    Resolvers: gfgraphql.SDLResolvers{ ... },
})

gqlApp, _ := gfgraphql.CreateApplication(gfgraphql.ApplicationConfig{
    Modules: []gfgraphql.Module{userMod, postMod},
})
```

Rules when combining modules:

- Each module requires a unique `ID`
- Duplicate `Query` / `Mutation` fields across modules are rejected
- Types with the same name must be identical

---

## Supported GraphQL types

Gofreight builds on [graphql-go](https://github.com/graphql-go/graphql). Types available from **SDL strings**:

| GraphQL type | SDL example | Go value in resolvers |
|--------------|-------------|----------------------|
| `String` | `name: String` | `string` |
| `Int` | `count: Int` | `int`, `int32`, `int64` |
| `Float` | `price: Float` | `float32`, `float64` |
| `Boolean` | `published: Boolean` | `bool` |
| `ID` | `id: ID!` | `string` |
| `Enum` | `status: Status!` | `string` (enum value name) |
| `Object` | `author: User!` | `map[string]any`, struct, or pointer |
| `InputObject` | `input: CreatePostInput!` | `map[string]any` (from args) |
| `List` | `tags: [String!]!` | `[]any`, typed slice |
| `NonNull` | `title: String!` | Must not return `nil` |
| Custom scalar | `date: DateTime` | Any serializable value (configure via `Scalars`) |

**Programmatic-only** (use `gql.NewInterface`, `gql.NewUnion`):

| Type | Notes |
|------|-------|
| `Interface` | Requires `ResolveType` function |
| `Union` | Requires `ResolveType` function |
| `Subscription` | Root fields via SDL or programmatic `ModuleConfig` |

### Custom scalars in SDL

```go
gfgraphql.ModuleFromSDL(gfgraphql.SDLModuleConfig{
    SDL: `
        scalar DateTime
        type Event { startsAt: DateTime! }
        extend type Query { events: [Event!]! }
    `,
    Scalars: map[string]gql.ScalarConfig{
        "DateTime": {
            Name:        "DateTime",
            Description: "ISO8601 datetime",
            Serialize:   func(value any) any { return value },
            ParseValue:  func(value any) any { return value },
        },
    },
    Resolvers: ...,
})
```

### Type modifiers

```graphql
title: String          # nullable String
title: String!         # non-null String
posts: [Post]          # nullable list of nullable Post
posts: [Post!]!        # non-null list of non-null Post
```

---

## GraphQL modules

Each module is a self-contained schema fragment. Create with `CreateModule` (programmatic) or `ModuleFromSDL` (SDL string):

```go
mod := gfgraphql.MustCreateModule(gfgraphql.ModuleConfig{
    ID:          "user",
    Description: "User types and queries",
    Types:       []gql.Type{userType},
    Query:       userQueries,
})
```

Merge modules with `CreateApplication`.

---

## Rate limiting

### Global (entire endpoint)

```go
gfgraphql.ApplicationConfig{RateLimit: 120} // per IP per minute
```

### Per-field

```go
limits := gfgraphql.NewFieldRateLimitRegistry()
limits.Set("Mutation", "createPost", gfgraphql.RateLimitRule{Limit: 30, Window: time.Minute})

// Or on a single field:
gfgraphql.Field(gfgraphql.FieldConfig{ ... }, gfgraphql.WithFieldRateLimit(30, time.Minute))
```

---

## DataLoader

Batch N+1 queries with per-request loaders:

```go
reg.Register("user", func() any {
    return gfgraphql.NewLoader[string, *User](batchLoadUsers)
})

loader, _ := gfgraphql.LoaderFromContext[string, *User](ctx, "user")
return loader.Load(ctx, userID)()
```

Register loaders in `OnRequest` — they are created once per GraphQL request.

---

## Fragments

GraphQL **fragments** let clients reuse field selections across queries. Gofreight supports the full GraphQL fragment spec — named fragments, inline fragments, and fragment spreads — at execution time and in security validation.

No server-side setup is required. Define reusable selections in your **client queries** (or in the playground).

### Named fragments

Define a reusable block with `fragment Name on Type`:

```graphql
fragment PostFields on Post {
  id
  title
  body
}

query ListPosts {
  posts {
    ...PostFields
    author {
      name
    }
  }
}

query GetPost($id: ID!) {
  post(id: $id) {
    ...PostFields
  }
}
```

Spread a fragment with `...PostFields`. Resolvers run the same way — each field inside the fragment triggers the same resolver as if it were written inline.

### Inline fragments

Select fields conditionally on a type without defining a named fragment:

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

Useful for union or interface fields when you add those types programmatically.

### Fragments with variables

Operation variables work inside fragment spreads:

```graphql
query PostDetail($id: ID!) {
  post(id: $id) {
    ...PostFields
  }
}
```

Send variables in the POST body:

```json
{
  "query": "query PostDetail($id: ID!) { post(id: $id) { ...PostFields } } fragment PostFields on Post { id title }",
  "variables": { "id": "1" }
}
```

### Client-side with `gql`

In JavaScript or TypeScript, combine SDL schema strings (server) with fragment queries (client):

```typescript
import { gql, request } from 'graphql-request'

const PostFields = gql`
  fragment PostFields on Post {
    id
    title
    body
  }
`

const LIST_POSTS = gql`
  ${PostFields}
  query ListPosts {
    posts {
      ...PostFields
      author { name }
    }
  }
`

const data = await request('http://localhost:5000/graphql', LIST_POSTS)
```

On the server, only resolvers matter — fragments are expanded by the GraphQL engine before field execution.

### Security and fragments

Depth and complexity limits **expand fragment spreads** before checking limits, so clients cannot bypass `MaxDepth` or `MaxComplexity` by hiding fields inside fragments.

```go
gfgraphql.ValidateQuery(query, gfgraphql.ProductionSecurity())
gfgraphql.ParseQuery(query) // parse operations + fragment definitions
```

Recursive fragment spreads (`...A` → `...A`) are detected and do not cause infinite loops during validation.

### Playground example

Paste into **http://localhost:5000/graphql/playground**:

```graphql
fragment PostCard on Post {
  id
  title
  author {
    name
  }
}

query {
  posts {
    ...PostCard
  }
}
```

---

## Security

| Option | Default (dev) | Production |
|--------|---------------|------------|
| `MaxDepth` | 10 | 8 |
| `MaxComplexity` | 200 | 100 |
| `AllowIntrospection` | true | false |
| `RateLimit` | optional | recommended |

```go
gfgraphql.ProductionSecurity()
```

Queries are validated **before execution**. Use `auth.Guard.Middleware` on the route group if the endpoint requires authentication.

---

## Example queries

```graphql
query {
  posts {
    id
    title
    author { name email }
  }
}

mutation {
  createPost(title: "New post", body: "Hello", authorId: "1") {
    id
    title
  }
}
```

---

## Demo app

```bash
cd demoapp
gofreight serve
# Playground: http://localhost:5000/graphql/playground
```

---

## Related

- **[Tutorial: GraphQL](tutorial-graphql.md)** — build a posts API with SDL and step-by-step resolvers
- [Authentication](authentication.md) — protect GraphQL routes
- [ORM](orm.md) — load data in resolvers
