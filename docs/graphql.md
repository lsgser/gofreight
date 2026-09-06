# GraphQL

Gofreight includes a modular GraphQL server inspired by [graphql-modules](https://github.com/graphql-hive/graphql-modules) — reusable schema modules, DataLoader support, GraphiQL playground, and layered security limits.

## Quick start

### 1. Create modules

Each module owns types, queries, and mutations. Build modules with `CreateModule`:

```go
// app/graphql/post_module.go
func PostModule() gfgraphql.Module {
    postType := graphql.NewObject(graphql.ObjectConfig{ ... })

    return gfgraphql.MustCreateModule(gfgraphql.ModuleConfig{
        ID:          "post",
        Description: "Post queries and mutations",
        Types:       []graphql.Type{postType},
        Query: graphql.Fields{
            "posts": &graphql.Field{ ... },
        },
        Mutation: graphql.Fields{
            "createPost": gfgraphql.Field(graphql.FieldConfig{
                Type: postType,
                Args: ...,
                Resolve: createPostResolver,
            }, gfgraphql.WithFieldRateLimit(30, time.Minute)),
        },
    })
}
```

### 2. Create the application

Merge modules into one executable schema with `CreateApplication`:

```go
import gfgraphql "github.com/lsgser/gofreight/graphql"

fieldLimits := gfgraphql.NewFieldRateLimitRegistry()
fieldLimits.Set("Mutation", "createPost", gfgraphql.RateLimitRule{
    Limit:  30,
    Window: time.Minute,
})

gqlApp, err := gfgraphql.CreateApplication(gfgraphql.ApplicationConfig{
    Modules: []gfgraphql.Module{
        UserModule(),
        PostModule(),
    },
    Playground:      true,
    RateLimit:       120, // global per IP per minute
    FieldRateLimits: fieldLimits,
    Security:        gfgraphql.DefaultSecurity(),
    OnRequest: func(r *http.Request, loaders *gfgraphql.LoaderRegistry) context.Context {
        registerLoaders(loaders)
        return gfgraphql.DefaultOnRequest(r, loaders)
    },
})
if err != nil {
    log.Fatal(err)
}

gqlApp.Mount(app.Router, "/graphql")
// Or: app.MountGraphQLApplication("/graphql", gqlApp)
```

### 3. Open the playground

- **GraphQL endpoint:** `POST /graphql`
- **Playground:** `GET /graphql/playground` (development)
- **Schema docs:** `GET /graphql/docs`

## GraphQL modules

Similar to [graphql-modules](https://github.com/graphql-hive/graphql-modules), each module is a self-contained schema fragment:

```go
userMod := gfgraphql.CreateModule(gfgraphql.ModuleConfig{
    ID:          "user",
    Description: "User types and queries",
    Types:       []graphql.Type{userType},
    Query:       userQueries,
})

postMod := gfgraphql.CreateModule(gfgraphql.ModuleConfig{
    ID:       "post",
    Types:    []graphql.Type{postType},
    Query:    postQueries,
    Mutation: postMutations,
})

gqlApp, _ := gfgraphql.CreateApplication(gfgraphql.ApplicationConfig{
    Modules: []gfgraphql.Module{userMod, postMod},
})
```

Rules when combining modules:

- Each module requires a unique `ID`
- Duplicate `Query` / `Mutation` fields across modules are rejected
- Types with the same name must be identical

See `demoapp/graphql/` for a working multi-module setup.

## Rate limiting

Rate limits work at **two levels**:

### Global (entire GraphQL endpoint)

Limits total requests per IP per minute:

```go
gfgraphql.ApplicationConfig{
    RateLimit: 120,
}
```

### Per-field (specific endpoints)

Limit individual queries or mutations:

**Option A — registry (checked before execution):**

```go
limits := gfgraphql.NewFieldRateLimitRegistry()
limits.Set("Query", "expensiveReport", gfgraphql.RateLimitRule{Limit: 10, Window: time.Minute})
limits.Set("Mutation", "createPost", gfgraphql.RateLimitRule{Limit: 30, Window: time.Minute})

cfg.FieldRateLimits = limits
```

**Option B — field helper (resolver wrapper):**

```go
"createPost": gfgraphql.Field(graphql.FieldConfig{
    Type:    postType,
    Resolve: createPost,
}, gfgraphql.WithFieldRateLimit(30, time.Minute)),
```

Or wrap an existing field:

```go
gfgraphql.RateLimitedField("Mutation", field, gfgraphql.RateLimitRule{Limit: 30, Window: time.Minute})
```

## DataLoader

Batch N+1 queries with per-request loaders:

```go
reg.Register("user", func() any {
    return gfgraphql.NewLoader[string, *User](batchLoadUsers)
})

loader, _ := gfgraphql.LoaderFromContext[string, *User](ctx, "user")
return loader.Load(ctx, userID)()
```

Loaders are created once per GraphQL request via `OnRequest`.

## Security

Configure via `SecurityConfig`:

| Option | Default (dev) | Production |
|--------|---------------|------------|
| `MaxDepth` | 10 | 8 |
| `MaxComplexity` | 200 | 100 |
| `AllowIntrospection` | true | false |
| `RateLimit` | optional | recommended |
| `FieldRateLimits` | optional | per sensitive field |

```go
gfgraphql.ProductionSecurity()
```

Queries are validated **before execution**. Introspection (`__schema`, `__type`) can be disabled in production.

## Self-documentation

- **Introspection** — standard GraphQL `__schema` when enabled
- **`/graphql/docs`** — plain-text module index
- **Playground** — GraphiQL with docs pane (development)

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

## Demo app

```bash
cd demoapp
gofreight serve
# Playground: http://localhost:5000/graphql/playground
```

## Production notes

- Disable playground in production
- Set global `RateLimit` and per-field limits on expensive mutations
- Use `auth.Guard.Middleware` if the endpoint requires authentication
- For subscriptions at scale, consider a dedicated GraphQL gateway
