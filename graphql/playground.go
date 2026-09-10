package graphql

/*
|--------------------------------------------------------------------------
| Playground
|--------------------------------------------------------------------------
|
| Implements Playground as part of the graphql package in the Gofreight
| framework.
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
*/

import (
	_ "embed"
	"net/http"
)

//go:embed playground.html
var playgroundHTML []byte

func (s *Server) playgroundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(playgroundHTML)
	}
}
