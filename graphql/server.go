package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	gql "github.com/graphql-go/graphql"
	"github.com/lsgser/gofreight/middleware"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

// Config configures the GraphQL server.
type Config struct {
	Modules         []Module
	Security        SecurityConfig
	Playground      bool
	Path            string
	RateLimit       int // global requests per IP per minute (0 = disabled)
	FieldRateLimits *FieldRateLimitRegistry
	OnRequest       func(r *http.Request, loaders *LoaderRegistry) context.Context
}

// Server serves GraphQL over HTTP with playground and security plugins.
type Server struct {
	schema    gql.Schema
	config    Config
	limiter   *rateLimiter
	parsedDoc func(string) (*ast.QueryDocument, error)
}

// NewServer builds a server from config, combining all modules.
func NewServer(cfg Config) (*Server, error) {
	schema, err := Combine(cfg.Modules...)
	if err != nil {
		return nil, err
	}
	if cfg.Path == "" {
		cfg.Path = "/graphql"
	}
	s := &Server{
		schema: schema,
		config: cfg,
		parsedDoc: func(query string) (*ast.QueryDocument, error) {
			return parser.ParseQuery(&ast.Source{Input: query})
		},
	}
	if cfg.RateLimit > 0 {
		s.limiter = newRateLimiter(cfg.RateLimit, time.Minute)
	}
	return s, nil
}

// Schema returns the compiled GraphQL schema.
func (s *Server) Schema() gql.Schema {
	return s.schema
}

// Handler returns the HTTP handler for GraphQL POST/GET.
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(s.serveHTTP)
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if s.limiter != nil && !s.limiter.allow(RequestClientIP(r)) {
		writeGQLError(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.serveGET(w, r)
	case http.MethodPost:
		s.servePOST(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) serveGET(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		if s.config.Playground {
			s.playgroundHandler().ServeHTTP(w, r)
			return
		}
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}
	variables := parseVariables(r.URL.Query().Get("variables"))
	s.execute(w, r, query, variables, r.URL.Query().Get("operationName"))
}

func (s *Server) servePOST(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Query         string         `json:"query"`
		Variables     map[string]any `json:"variables"`
		OperationName string         `json:"operationName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeGQLError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	s.execute(w, r, body.Query, body.Variables, body.OperationName)
}

func (s *Server) execute(w http.ResponseWriter, r *http.Request, query string, variables map[string]any, operationName string) {
	if err := ValidateQuery(query, s.config.Security); err != nil {
		writeGQLError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if s.config.FieldRateLimits != nil {
		doc, err := s.parsedDoc(query)
		if err != nil {
			writeGQLError(w, "Invalid query", http.StatusBadRequest)
			return
		}
		if err := s.config.FieldRateLimits.Check(doc, RequestClientIP(r)); err != nil {
			writeGQLError(w, err.Error(), http.StatusTooManyRequests)
			return
		}
	}

	loaders := NewLoaderRegistry()
	ctx := r.Context()
	if s.config.OnRequest != nil {
		ctx = s.config.OnRequest(r, loaders)
	} else {
		ctx = WithLoaders(ctx, loaders)
	}
	ctx = WithClientIP(ctx, RequestClientIP(r))

	params := gql.Params{
		Schema:         s.schema,
		RequestString:  query,
		VariableValues: variables,
		OperationName:  operationName,
		Context:        ctx,
	}

	result := gql.Do(params)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// PrintSchema returns a human-readable schema description for self-documentation.
func (s *Server) PrintSchema() string {
	var b strings.Builder
	b.WriteString("# Gofreight GraphQL Schema\n\n")
	b.WriteString("## Modules\n\n")
	for _, m := range s.config.Modules {
		b.WriteString("- **")
		b.WriteString(moduleID(m))
		b.WriteString("**")
		if m.Description != "" {
			b.WriteString(": ")
			b.WriteString(m.Description)
		}
		b.WriteString("\n")
	}
	b.WriteString("\nUse introspection (`__schema`) or the playground when enabled.\n")
	return b.String()
}

func (s *Server) schemaDocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(s.PrintSchema()))
	}
}

func writeGQLError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"errors": []map[string]string{{"message": msg}},
	})
}

func parseVariables(raw string) map[string]any {
	if raw == "" {
		return nil
	}
	var vars map[string]any
	_ = json.Unmarshal([]byte(raw), &vars)
	return vars
}

// DefaultOnRequest attaches loaders, session, and client IP to GraphQL context.
func DefaultOnRequest(r *http.Request, loaders *LoaderRegistry) context.Context {
	ctx := WithLoaders(r.Context(), loaders)
	ctx = WithClientIP(ctx, RequestClientIP(r))
	if session := middleware.SessionFromContext(r.Context()); session != nil {
		ctx = context.WithValue(ctx, sessionContextKey{}, session)
	}
	return ctx
}

type sessionContextKey struct{}
