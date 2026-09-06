package graphql

import (
	"context"
	"fmt"
	"net/http"

	gql "github.com/graphql-go/graphql"
	"github.com/lsgser/gofreight/router"
)

// ModuleConfig defines a reusable GraphQL module (graphql-modules style).
// Each module owns a slice of the schema: types, queries, and mutations.
type ModuleConfig struct {
	ID           string
	Description  string
	Types        []gql.Type
	Query        gql.Fields
	Mutation     gql.Fields
	Subscription gql.Fields
}

// Module is a composable GraphQL schema module.
// Use CreateModule to build modules and CreateApplication to merge them.
type Module struct {
	ID           string
	Description  string
	Types        []gql.Type
	Query        gql.Fields
	Mutation     gql.Fields
	Subscription gql.Fields
}

// CreateModule builds a GraphQL module from config, similar to graphql-modules.
func CreateModule(cfg ModuleConfig) (Module, error) {
	if cfg.ID == "" {
		return Module{}, fmt.Errorf("graphql: module ID is required")
	}
	return Module{
		ID:           cfg.ID,
		Description:  cfg.Description,
		Types:        cfg.Types,
		Query:        cfg.Query,
		Mutation:     cfg.Mutation,
		Subscription: cfg.Subscription,
	}, nil
}

// MustCreateModule calls CreateModule and panics on error (for init/bootstrap code).
func MustCreateModule(cfg ModuleConfig) Module {
	mod, err := CreateModule(cfg)
	if err != nil {
		panic(err)
	}
	return mod
}

// ApplicationConfig configures a combined GraphQL application from modules.
type ApplicationConfig struct {
	Modules          []Module
	Security         SecurityConfig
	Playground       bool
	Path             string
	RateLimit        int // global requests per IP per minute (0 = disabled)
	FieldRateLimits  *FieldRateLimitRegistry
	OnRequest        func(r *http.Request, loaders *LoaderRegistry) context.Context
}

// Application is a merged GraphQL application built from modules.
type Application struct {
	server *Server
}

// CreateApplication merges modules into one executable GraphQL application.
func CreateApplication(cfg ApplicationConfig) (*Application, error) {
	srv, err := NewServer(Config{
		Modules:         cfg.Modules,
		Security:        cfg.Security,
		Playground:      cfg.Playground,
		Path:            cfg.Path,
		RateLimit:       cfg.RateLimit,
		FieldRateLimits: cfg.FieldRateLimits,
		OnRequest:       cfg.OnRequest,
	})
	if err != nil {
		return nil, err
	}
	return &Application{server: srv}, nil
}

// Server returns the underlying HTTP server.
func (a *Application) Server() *Server {
	return a.server
}

// Handler returns the GraphQL HTTP handler.
func (a *Application) Handler() http.Handler {
	return a.server.Handler()
}

// Mount registers routes on the Gofreight router.
func (a *Application) Mount(r *router.Router, path string) {
	a.server.Mount(r, path)
}

// Schema returns the compiled GraphQL schema.
func (a *Application) Schema() gql.Schema {
	return a.server.Schema()
}

// PrintSchema returns human-readable module documentation.
func (a *Application) PrintSchema() string {
	return a.server.PrintSchema()
}

// Combine merges modules into a single executable schema.
func Combine(modules ...Module) (gql.Schema, error) {
	if len(modules) == 0 {
		return gql.Schema{}, fmt.Errorf("graphql: at least one module is required")
	}

	typeMap := map[string]gql.Type{}
	queryFields := gql.Fields{}
	mutationFields := gql.Fields{}
	subscriptionFields := gql.Fields{}

	for _, mod := range modules {
		id := moduleID(mod)
		for _, t := range mod.Types {
			if t == nil {
				continue
			}
			name := typeName(t)
			if name == "" {
				continue
			}
			if existing, ok := typeMap[name]; ok && existing != t {
				return gql.Schema{}, fmt.Errorf("graphql: module %q redefines type %s", id, name)
			}
			typeMap[name] = t
		}

		for field, cfg := range mod.Query {
			if _, exists := queryFields[field]; exists {
				return gql.Schema{}, fmt.Errorf("graphql: duplicate Query.%s", field)
			}
			queryFields[field] = cfg
		}
		for field, cfg := range mod.Mutation {
			if _, exists := mutationFields[field]; exists {
				return gql.Schema{}, fmt.Errorf("graphql: duplicate Mutation.%s", field)
			}
			mutationFields[field] = cfg
		}
		for field, cfg := range mod.Subscription {
			if _, exists := subscriptionFields[field]; exists {
				return gql.Schema{}, fmt.Errorf("graphql: duplicate Subscription.%s", field)
			}
			subscriptionFields[field] = cfg
		}
	}

	types := make([]gql.Type, 0, len(typeMap))
	for _, t := range typeMap {
		types = append(types, t)
	}

	var queryType, mutationType, subscriptionType gql.Type

	if len(queryFields) > 0 {
		queryType = gql.NewObject(gql.ObjectConfig{
			Name:   "Query",
			Fields: queryFields,
		})
		types = append(types, queryType)
	}

	if len(mutationFields) > 0 {
		mutationType = gql.NewObject(gql.ObjectConfig{
			Name:   "Mutation",
			Fields: mutationFields,
		})
		types = append(types, mutationType)
	}

	if len(subscriptionFields) > 0 {
		subscriptionType = gql.NewObject(gql.ObjectConfig{
			Name:   "Subscription",
			Fields: subscriptionFields,
		})
		types = append(types, subscriptionType)
	}

	if queryType == nil {
		return gql.Schema{}, fmt.Errorf("graphql: combined schema requires at least one Query field")
	}

	return gql.NewSchema(gql.SchemaConfig{
		Query:        queryType.(*gql.Object),
		Mutation:     objectOrNil(mutationType),
		Subscription: objectOrNil(subscriptionType),
		Types:        types,
	})
}

func objectOrNil(t gql.Type) *gql.Object {
	if t == nil {
		return nil
	}
	obj, _ := t.(*gql.Object)
	return obj
}

func typeName(t gql.Type) string {
	switch v := t.(type) {
	case *gql.Object:
		return v.Name()
	case *gql.InputObject:
		return v.Name()
	case *gql.Enum:
		return v.Name()
	case *gql.Scalar:
		return v.Name()
	case *gql.Union:
		return v.Name()
	case *gql.Interface:
		return v.Name()
	default:
		if n, ok := t.(interface{ Name() string }); ok {
			return n.Name()
		}
		return ""
	}
}

func moduleID(mod Module) string {
	if mod.ID != "" {
		return mod.ID
	}
	return "module"
}

// ModuleNames returns registered module IDs.
func ModuleNames(modules []Module) []string {
	out := make([]string, len(modules))
	for i, m := range modules {
		out[i] = moduleID(m)
	}
	return out
}
