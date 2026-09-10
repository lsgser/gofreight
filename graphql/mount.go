package graphql

/*
|--------------------------------------------------------------------------
| Mount
|--------------------------------------------------------------------------
|
| Implements Mount as part of the graphql package in the Gofreight
| framework. Key symbols: Mount, MountOn.
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
| Symbols defined here include: Mount (Mount registers GraphQL routes on
| the application router.); MountOn (Mountable matches router and channels
| mount patterns.).
| 
*/

import (
	"net/http"
	"strings"

	"github.com/lsgser/gofreight/router"
)

// Mount registers GraphQL routes on the application router.
func (s *Server) Mount(r *router.Router, path string) {
	if path == "" {
		path = s.config.Path
	}
	if path == "" {
		path = "/graphql"
	}
	path = strings.TrimSuffix(path, "/")

	handler := s.Handler()
	r.Get(path, handler.ServeHTTP, "graphql")
	r.Post(path, handler.ServeHTTP)
	r.Get(path+"/docs", s.schemaDocsHandler().ServeHTTP, "graphql.docs")

	if s.config.Playground {
		r.Get(path+"/playground", s.playgroundHandler().ServeHTTP, "graphql.playground")
	}
}

// Mountable matches router and channels mount patterns.
func (s *Server) MountOn(m interface {
	Get(path string, handler http.HandlerFunc, name ...string)
	Post(path string, handler http.HandlerFunc, name ...string)
}, path string) {
	if path == "" {
		path = s.config.Path
	}
	if path == "" {
		path = "/graphql"
	}
	path = strings.TrimSuffix(path, "/")

	handler := s.Handler()
	m.Get(path, handler.ServeHTTP, "graphql")
	m.Post(path, handler.ServeHTTP)
}
