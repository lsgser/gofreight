package graphql_test

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
