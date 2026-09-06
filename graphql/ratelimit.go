package graphql

import (
	"fmt"
	"sync"
	"time"

	"github.com/vektah/gqlparser/v2/ast"
)

// RateLimitRule limits how often a key may be used within a time window.
type RateLimitRule struct {
	Limit  int
	Window time.Duration
}

// FieldRateLimitRegistry applies rate limits to specific GraphQL fields
// (e.g. Query.posts, Mutation.createPost) in addition to any global limit.
type FieldRateLimitRegistry struct {
	mu       sync.Mutex
	rules    map[string]RateLimitRule
	limiters map[string]*rateLimiter
}

// NewFieldRateLimitRegistry creates an empty per-field rate limit registry.
func NewFieldRateLimitRegistry() *FieldRateLimitRegistry {
	return &FieldRateLimitRegistry{
		rules:    make(map[string]RateLimitRule),
		limiters: make(map[string]*rateLimiter),
	}
}

// Set registers a rate limit for a field. parentType is Query, Mutation, or Subscription.
func (r *FieldRateLimitRegistry) Set(parentType, field string, rule RateLimitRule) {
	if rule.Window == 0 {
		rule.Window = time.Minute
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules[fieldKey(parentType, field)] = rule
}

// Check validates requested root fields against registered limits.
func (r *FieldRateLimitRegistry) Check(doc *ast.QueryDocument, ip string) error {
	if r == nil || len(r.rules) == 0 {
		return nil
	}

	for _, ref := range rootFieldRefs(doc) {
		key := fieldKey(ref.Parent, ref.Name)
		r.mu.Lock()
		rule, ok := r.rules[key]
		limiter := r.limiters[key]
		if ok && limiter == nil {
			limiter = newRateLimiter(rule.Limit, rule.Window)
			r.limiters[key] = limiter
		}
		r.mu.Unlock()

		if !ok {
			continue
		}
		if !limiter.allow(ip + ":" + key) {
			return fmt.Errorf("graphql: rate limit exceeded for %s.%s", ref.Parent, ref.Name)
		}
	}
	return nil
}

type fieldRef struct {
	Parent string
	Name   string
}

func rootFieldRefs(doc *ast.QueryDocument) []fieldRef {
	var out []fieldRef
	for _, op := range doc.Operations {
		parent := "Query"
		switch op.Operation {
		case ast.Mutation:
			parent = "Mutation"
		case ast.Subscription:
			parent = "Subscription"
		}
		for _, sel := range op.SelectionSet {
			if field, ok := sel.(*ast.Field); ok {
				out = append(out, fieldRef{Parent: parent, Name: field.Name})
			}
		}
	}
	return out
}

func fieldKey(parentType, field string) string {
	return parentType + "." + field
}

// rateLimiter is a lightweight per-key rate limiter.
type rateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	requests map[string][]time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	if window == 0 {
		window = time.Minute
	}
	return &rateLimiter{
		limit:    limit,
		window:   window,
		requests: make(map[string][]time.Time),
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	var recent []time.Time
	for _, t := range rl.requests[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= rl.limit {
		rl.requests[key] = recent
		return false
	}
	rl.requests[key] = append(recent, now)
	return true
}
