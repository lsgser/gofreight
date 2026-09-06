package graphql

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
