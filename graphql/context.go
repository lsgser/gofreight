package graphql

/*
|--------------------------------------------------------------------------
| Context
|--------------------------------------------------------------------------
|
| Implements Context as part of the graphql package in the Gofreight
| framework. Key symbols: WithClientIP, ClientIPFromContext,
| RequestClientIP.
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
| Symbols defined here include: WithClientIP (WithClientIP stores the
| request client IP in context (used by field rate limits).);
| ClientIPFromContext (ClientIPFromContext returns the client IP attached
| to a GraphQL request.); RequestClientIP (RequestClientIP extracts the
| client IP from an HTTP request.).
| 
*/

import (
	"context"
	"net/http"
)

type clientIPKey struct{}

// WithClientIP stores the request client IP in context (used by field rate limits).
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

// ClientIPFromContext returns the client IP attached to a GraphQL request.
func ClientIPFromContext(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey{}).(string)
	return ip
}

// RequestClientIP extracts the client IP from an HTTP request.
func RequestClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		for i := 0; i < len(fwd); i++ {
			if fwd[i] == ',' {
				return fwd[:i]
			}
		}
		return fwd
	}
	return r.RemoteAddr
}
