package graphql

/*
|--------------------------------------------------------------------------
| Modules
|--------------------------------------------------------------------------
|
| Implements Modules as part of the graphql package in the Gofreight
| framework. Key symbols: Modules, RegisterLoaders.
| 
| Symbols defined here include: Modules (/*
| |--------------------------------------------------------------------------
| | Modules
| |--------------------------------------------------------------------------
| | | Returns demo GraphQL modules (user + post) built with CreateModule.
| | */); RegisterLoaders (/*
| |--------------------------------------------------------------------------
| | RegisterLoaders
| |--------------------------------------------------------------------------
| | | Attaches dataloaders to each GraphQL request. | */).
| 
*/

import (
	"context"
	"fmt"
	"sync"
	"time"

	gql "github.com/graphql-go/graphql"
	"github.com/graph-gophers/dataloader/v7"
	gfgraphql "github.com/lsgser/gofreight/graphql"
)

var (
	mu    sync.RWMutex
	posts = []map[string]any{
		{"id": "1", "title": "Hello GraphQL", "body": "Gofreight GraphQL is live.", "authorId": "1"},
		{"id": "2", "title": "Modular schema", "body": "Combine modules into one endpoint.", "authorId": "2"},
	}
	users = []map[string]any{
		{"id": "1", "name": "Ada", "email": "ada@example.com"},
		{"id": "2", "name": "Lin", "email": "lin@example.com"},
	}
)

/*
|--------------------------------------------------------------------------
| Modules
|--------------------------------------------------------------------------
|
| Returns demo GraphQL modules (user + post) built with CreateModule.
|
*/
func Modules() []gfgraphql.Module {
	userType := gql.NewObject(gql.ObjectConfig{
		Name: "User",
		Fields: gql.Fields{
			"id":    &gql.Field{Type: gql.NewNonNull(gql.String)},
			"name":  &gql.Field{Type: gql.NewNonNull(gql.String)},
			"email": &gql.Field{Type: gql.NewNonNull(gql.String)},
		},
	})

	postType := gql.NewObject(gql.ObjectConfig{
		Name: "Post",
		Fields: gql.Fields{
			"id":    &gql.Field{Type: gql.NewNonNull(gql.String)},
			"title": &gql.Field{Type: gql.NewNonNull(gql.String)},
			"body":  &gql.Field{Type: gql.String},
			"author": &gql.Field{
				Type: userType,
				Resolve: func(p gql.ResolveParams) (any, error) {
					authorID, _ := p.Source.(map[string]any)["authorId"].(string)
					return loadUser(p.Context, authorID)
				},
			},
		},
	})

	userModule := gfgraphql.MustCreateModule(gfgraphql.ModuleConfig{
		ID:          "user",
		Description: "User queries and types",
		Types:       []gql.Type{userType},
		Query: gql.Fields{
			"users": &gql.Field{
				Type: gql.NewList(userType),
				Resolve: func(_ gql.ResolveParams) (any, error) {
					mu.RLock()
					defer mu.RUnlock()
					return users, nil
				},
			},
			"user": &gql.Field{
				Type: userType,
				Args: gql.FieldConfigArgument{
					"id": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.String)},
				},
				Resolve: func(p gql.ResolveParams) (any, error) {
					id, _ := p.Args["id"].(string)
					return loadUser(p.Context, id)
				},
			},
		},
	})

	postModule := gfgraphql.MustCreateModule(gfgraphql.ModuleConfig{
		ID:          "post",
		Description: "Post queries, mutations, and types",
		Types:       []gql.Type{postType},
		Query: gql.Fields{
			"posts": &gql.Field{
				Type: gql.NewList(postType),
				Resolve: func(_ gql.ResolveParams) (any, error) {
					mu.RLock()
					defer mu.RUnlock()
					out := make([]map[string]any, len(posts))
					copy(out, posts)
					return out, nil
				},
			},
			"post": &gql.Field{
				Type: postType,
				Args: gql.FieldConfigArgument{
					"id": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.String)},
				},
				Resolve: func(p gql.ResolveParams) (any, error) {
					id, _ := p.Args["id"].(string)
					return loadPost(p.Context, id)
				},
			},
		},
		Mutation: gql.Fields{
			"createPost": gfgraphql.Field(gfgraphql.FieldConfig{
				Type: postType,
				Args: gql.FieldConfigArgument{
					"title":    &gql.ArgumentConfig{Type: gql.NewNonNull(gql.String)},
					"body":     &gql.ArgumentConfig{Type: gql.String},
					"authorId": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.String)},
				},
				Resolve: func(p gql.ResolveParams) (any, error) {
					title, _ := p.Args["title"].(string)
					body, _ := p.Args["body"].(string)
					authorID, _ := p.Args["authorId"].(string)
					mu.Lock()
					defer mu.Unlock()
					id := fmt.Sprintf("%d", len(posts)+1)
					post := map[string]any{
						"id": id, "title": title, "body": body, "authorId": authorID,
					}
					posts = append(posts, post)
					return post, nil
				},
			}, gfgraphql.WithFieldRateLimit(10, time.Minute)),
		},
	})

	return []gfgraphql.Module{userModule, postModule}
}

/*
|--------------------------------------------------------------------------
| RegisterLoaders
|--------------------------------------------------------------------------
|
| Attaches dataloaders to each GraphQL request.
|
*/
func RegisterLoaders(reg *gfgraphql.LoaderRegistry) {
	reg.Register("user", func() any {
		return gfgraphql.NewLoader[string, map[string]any](func(ctx context.Context, keys []string) []*dataloader.Result[map[string]any] {
			results := make([]*dataloader.Result[map[string]any], len(keys))
			for i, key := range keys {
				results[i] = &dataloader.Result[map[string]any]{Data: findUser(key)}
			}
			return results
		})
	})
	reg.Register("post", func() any {
		return gfgraphql.NewLoader[string, map[string]any](func(ctx context.Context, keys []string) []*dataloader.Result[map[string]any] {
			results := make([]*dataloader.Result[map[string]any], len(keys))
			for i, key := range keys {
				results[i] = &dataloader.Result[map[string]any]{Data: findPost(key)}
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

func loadPost(ctx context.Context, id string) (map[string]any, error) {
	if loader, ok := gfgraphql.LoaderFromContext[string, map[string]any](ctx, "post"); ok {
		return loader.Load(ctx, id)()
	}
	p := findPost(id)
	if p == nil {
		return nil, fmt.Errorf("post not found")
	}
	return p, nil
}

func findUser(id string) map[string]any {
	mu.RLock()
	defer mu.RUnlock()
	for _, u := range users {
		if u["id"] == id {
			return u
		}
	}
	return nil
}

func findPost(id string) map[string]any {
	mu.RLock()
	defer mu.RUnlock()
	for _, p := range posts {
		if p["id"] == id {
			return p
		}
	}
	return nil
}
