package graphql

/*
|--------------------------------------------------------------------------
| Field
|--------------------------------------------------------------------------
|
| Helpers to define GraphQL fields with optional per-resolver rate
| limiting.
| 
| Composes with graphql-go FieldConfig and security middleware documented
| in the GraphQL guide.
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
| Symbols defined here include: FieldConfig (exported type); FieldOption
| (exported type); WithFieldRateLimit (WithFieldRateLimit applies a
| per-field rate limit to this resolver endpoint.); Field (Field creates a
| GraphQL field with optional per-endpoint rate limiting.);
| RateLimitedField (RateLimitedField wraps an existing field resolver with
| a per-endpoint rate limit.).
| 
*/

import (
	"fmt"
	"time"

	gql "github.com/graphql-go/graphql"
)

// FieldConfig defines a GraphQL field for the Field helper.
type FieldConfig struct {
	Type              gql.Output
	Args              gql.FieldConfigArgument
	Resolve           gql.FieldResolveFn
	DeprecationReason string
	Description       string
}

// FieldOption configures optional field behaviour.
type FieldOption func(*fieldBuild)

type fieldBuild struct {
	rateLimit *RateLimitRule
}

// WithFieldRateLimit applies a per-field rate limit to this resolver endpoint.
func WithFieldRateLimit(limit int, window time.Duration) FieldOption {
	if window == 0 {
		window = time.Minute
	}
	return func(b *fieldBuild) {
		b.rateLimit = &RateLimitRule{Limit: limit, Window: window}
	}
}

// Field creates a GraphQL field with optional per-endpoint rate limiting.
func Field(cfg FieldConfig, opts ...FieldOption) *gql.Field {
	build := &fieldBuild{}
	for _, opt := range opts {
		opt(build)
	}

	resolve := cfg.Resolve
	if build.rateLimit != nil && resolve != nil {
		original := resolve
		limiter := newRateLimiter(build.rateLimit.Limit, build.rateLimit.Window)
		resolve = func(p gql.ResolveParams) (any, error) {
			ip := ClientIPFromContext(p.Context)
			if ip == "" {
				ip = "unknown"
			}
			key := ip + ":" + p.Info.ParentType.Name() + "." + p.Info.FieldName
			if !limiter.allow(key) {
				return nil, fmt.Errorf("rate limit exceeded for %s.%s", p.Info.ParentType.Name(), p.Info.FieldName)
			}
			return original(p)
		}
	}

	return &gql.Field{
		Type:              cfg.Type,
		Args:              cfg.Args,
		Resolve:           resolve,
		DeprecationReason: cfg.DeprecationReason,
		Description:       cfg.Description,
	}
}

// RateLimitedField wraps an existing field resolver with a per-endpoint rate limit.
func RateLimitedField(parentType string, field *gql.Field, rule RateLimitRule) *gql.Field {
	if field == nil {
		return nil
	}
	if rule.Window == 0 {
		rule.Window = time.Minute
	}
	if field.Resolve == nil {
		return field
	}

	original := field.Resolve
	limiter := newRateLimiter(rule.Limit, rule.Window)
	field.Resolve = func(p gql.ResolveParams) (any, error) {
		ip := ClientIPFromContext(p.Context)
		if ip == "" {
			ip = "unknown"
		}
		key := ip + ":" + parentType + "." + p.Info.FieldName
		if !limiter.allow(key) {
			return nil, fmt.Errorf("rate limit exceeded for %s.%s", parentType, p.Info.FieldName)
		}
		return original(p)
	}
	return field
}
