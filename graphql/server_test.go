package graphql_test

/*
|--------------------------------------------------------------------------
| Server
|--------------------------------------------------------------------------
|
| Test suite for Server in the graphql package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| The graphql package integrates graphql-go with Gofreight: schema
| registration, HTTP mount, playground, DataLoader batching, and
| field-level rate limits.
| 
| Generate modules with gofreight make:graphql-module; SDL and resolvers
| live under app/graphql in your project.
| 
| See docs/graphql.md and the tutorial for N+1 avoidance and security
| middleware.
| 
| Run with go test ./graphql/... or go test for this package from the
| framework root.
| 
*/

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gql "github.com/graphql-go/graphql"
	"github.com/lsgser/gofreight/graphql"
)

func testModules() []graphql.Module {
	return []graphql.Module{
		graphql.MustCreateModule(graphql.ModuleConfig{
			ID: "hello",
			Query: gql.Fields{
				"hello": &gql.Field{
					Type: gql.String,
					Resolve: func(_ gql.ResolveParams) (any, error) {
						return "world", nil
					},
				},
			},
		}),
	}
}

func TestCreateModuleRequiresID(t *testing.T) {
	_, err := graphql.CreateModule(graphql.ModuleConfig{})
	if err == nil {
		t.Fatal("expected error for missing module ID")
	}
}

func TestCreateApplication(t *testing.T) {
	app, err := graphql.CreateApplication(graphql.ApplicationConfig{
		Modules:    testModules(),
		Security:   graphql.DefaultSecurity(),
		Playground: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if app == nil {
		t.Fatal("expected application")
	}
}

func TestGraphQLQuery(t *testing.T) {
	app, err := graphql.CreateApplication(graphql.ApplicationConfig{
		Modules:    testModules(),
		Security:   graphql.DefaultSecurity(),
		Playground: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	body := `{"query":"{ hello }"}`
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data["hello"] != "world" {
		t.Fatalf("expected world, got %q", resp.Data["hello"])
	}
}

func TestFieldRateLimitRegistry(t *testing.T) {
	limits := graphql.NewFieldRateLimitRegistry()
	limits.Set("Query", "hello", graphql.RateLimitRule{Limit: 1, Window: time.Minute})

	app, err := graphql.CreateApplication(graphql.ApplicationConfig{
		Modules:         testModules(),
		Security:        graphql.DefaultSecurity(),
		FieldRateLimits: limits,
	})
	if err != nil {
		t.Fatal(err)
	}

	body := `{"query":"{ hello }"}`
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first request failed: %d %s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req2)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on second request, got %d", rec.Code)
	}
}

func TestDepthLimit(t *testing.T) {
	security := graphql.DefaultSecurity()
	security.MaxDepth = 2

	query := `{ a { b { c { id } } } }`
	err := graphql.ValidateQuery(query, security)
	if err == nil {
		t.Fatal("expected depth limit error")
	}
}

func TestIntrospectionBlocked(t *testing.T) {
	security := graphql.ProductionSecurity()
	err := graphql.ValidateQuery("{ __schema { types { name } } }", security)
	if err == nil {
		t.Fatal("expected introspection to be blocked")
	}
}

func postFragmentModules() []graphql.Module {
	postType := gql.NewObject(gql.ObjectConfig{
		Name: "Post",
		Fields: gql.Fields{
			"id":    &gql.Field{Type: gql.NewNonNull(gql.ID)},
			"title": &gql.Field{Type: gql.NewNonNull(gql.String)},
			"body":  &gql.Field{Type: gql.String},
		},
	})

	return []graphql.Module{
		graphql.MustCreateModule(graphql.ModuleConfig{
			ID:    "post",
			Types: []gql.Type{postType},
			Query: gql.Fields{
				"posts": &gql.Field{
					Type: gql.NewList(postType),
					Resolve: func(_ gql.ResolveParams) (any, error) {
						return []map[string]any{
							{"id": "1", "title": "Hello", "body": "World"},
						}, nil
					},
				},
			},
		}),
	}
}

func TestGraphQLNamedFragmentQuery(t *testing.T) {
	app, err := graphql.CreateApplication(graphql.ApplicationConfig{
		Modules:    postFragmentModules(),
		Security:   graphql.DefaultSecurity(),
		Playground: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	query := `
		fragment PostFields on Post {
			id
			title
		}
		query {
			posts {
				...PostFields
				body
			}
		}
	`
	bodyBytes, _ := json.Marshal(map[string]string{"query": query})
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data struct {
			Posts []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
				Body  string `json:"body"`
			} `json:"posts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Data.Posts) != 1 || resp.Data.Posts[0].Title != "Hello" {
		t.Fatalf("unexpected data: %+v", resp.Data.Posts)
	}
}

func TestGraphQLInlineFragmentQuery(t *testing.T) {
	app, err := graphql.CreateApplication(graphql.ApplicationConfig{
		Modules:    postFragmentModules(),
		Security:   graphql.DefaultSecurity(),
		Playground: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	query := `{ posts { ... on Post { id title } } }`
	bodyBytes, _ := json.Marshal(map[string]string{"query": query})
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDepthLimitWithNamedFragment(t *testing.T) {
	security := graphql.DefaultSecurity()
	security.MaxDepth = 2

	query := `
		fragment Deep on Post {
			id
			body
		}
		query {
			posts {
				...Deep
			}
		}
	`
	if err := graphql.ValidateQuery(query, security); err != nil {
		t.Fatalf("expected fragment query within depth limit, got %v", err)
	}

	deepQuery := `
		fragment F1 on Post {
			wrapper {
				...F2
			}
		}
		fragment F2 on Wrapper {
			inner {
				id
			}
		}
		query {
			posts {
				...F1
			}
		}
	`
	if err := graphql.ValidateQuery(deepQuery, security); err == nil {
		t.Fatal("expected depth limit error for nested fields inside fragments")
	}
}

func TestComplexityWithFragmentSpread(t *testing.T) {
	security := graphql.DefaultSecurity()
	security.MaxComplexity = 3

	query := `
		fragment PostFields on Post {
			id
			title
			body
		}
		query {
			posts {
				...PostFields
			}
		}
	`
	if err := graphql.ValidateQuery(query, security); err == nil {
		t.Fatal("expected complexity limit error")
	}
}
