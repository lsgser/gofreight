package graphql_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gql "github.com/graphql-go/graphql"
	"github.com/lsgser/gofreight/graphql"
)

const postSDL = `
	type User {
		id: ID!
		name: String!
		email: String!
	}

	type Post {
		id: ID!
		title: String!
		body: String
	}

	extend type Query {
		hello: String!
		posts: [Post!]!
	}
`

func TestGQLValidatesSDL(t *testing.T) {
	_, err := graphql.GQL(`type Foo { bar: String }`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = graphql.GQL(`not valid graphql {{{`)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestModuleFromSDL(t *testing.T) {
	mod, err := graphql.ModuleFromSDL(graphql.SDLModuleConfig{
		ID:  "posts",
		SDL: postSDL,
		Resolvers: graphql.SDLResolvers{
			Query: map[string]gql.FieldResolveFn{
				"hello": func(_ gql.ResolveParams) (any, error) {
					return "Gofreight", nil
				},
				"posts": func(_ gql.ResolveParams) (any, error) {
					return []map[string]any{
						{"id": "1", "title": "First", "body": "Hello"},
					}, nil
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	app, err := graphql.CreateApplication(graphql.ApplicationConfig{
		Modules:    []graphql.Module{mod},
		Security:   graphql.DefaultSecurity(),
		Playground: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	body := `{"query":"{ hello posts { id title } }"}`
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data struct {
			Hello string `json:"hello"`
			Posts []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"posts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Hello != "Gofreight" {
		t.Fatalf("hello: got %q", resp.Data.Hello)
	}
	if len(resp.Data.Posts) != 1 || resp.Data.Posts[0].Title != "First" {
		t.Fatalf("posts: %+v", resp.Data.Posts)
	}
}

func TestModuleFromSDLRequiresQueryResolver(t *testing.T) {
	_, err := graphql.ModuleFromSDL(graphql.SDLModuleConfig{
		ID:  "bad",
		SDL: `extend type Query { ping: String! }`,
		Resolvers: graphql.SDLResolvers{
			Query: map[string]gql.FieldResolveFn{},
		},
	})
	if err == nil {
		t.Fatal("expected missing resolver error")
	}
}
