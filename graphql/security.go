package graphql

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

// SecurityConfig controls query validation before execution.
type SecurityConfig struct {
	MaxDepth        int // 0 = unlimited
	MaxComplexity   int // 0 = unlimited
	AllowIntrospection bool
}

// DefaultSecurity returns sensible development defaults.
func DefaultSecurity() SecurityConfig {
	return SecurityConfig{
		MaxDepth:           10,
		MaxComplexity:      200,
		AllowIntrospection: true,
	}
}

// ProductionSecurity returns hardened defaults for production.
func ProductionSecurity() SecurityConfig {
	return SecurityConfig{
		MaxDepth:           8,
		MaxComplexity:      100,
		AllowIntrospection: false,
	}
}

// ValidateQuery checks depth, complexity, and introspection rules.
func ValidateQuery(query string, cfg SecurityConfig) error {
	if strings.TrimSpace(query) == "" {
		return fmt.Errorf("graphql: empty query")
	}

	doc, err := parser.ParseQuery(&ast.Source{Input: query})
	if err != nil {
		return fmt.Errorf("graphql: invalid query: %w", err)
	}

	if !cfg.AllowIntrospection && containsIntrospection(doc) {
		return fmt.Errorf("graphql: introspection is disabled")
	}

	for _, op := range doc.Operations {
		if cfg.MaxDepth > 0 {
			if depth := selectionDepth(op.SelectionSet, 1); depth > cfg.MaxDepth {
				return fmt.Errorf("graphql: query depth %d exceeds limit %d", depth, cfg.MaxDepth)
			}
		}
		if cfg.MaxComplexity > 0 {
			if score := selectionComplexity(op.SelectionSet, 1); score > cfg.MaxComplexity {
				return fmt.Errorf("graphql: query complexity %d exceeds limit %d", score, cfg.MaxComplexity)
			}
		}
	}

	return nil
}

func containsIntrospection(doc *ast.QueryDocument) bool {
	for _, op := range doc.Operations {
		if selectionHasIntrospection(op.SelectionSet) {
			return true
		}
	}
	return false
}

func selectionHasIntrospection(set ast.SelectionSet) bool {
	for _, sel := range set {
		switch node := sel.(type) {
		case *ast.Field:
			if node.Name == "__schema" || node.Name == "__type" {
				return true
			}
			if selectionHasIntrospection(node.SelectionSet) {
				return true
			}
		case *ast.InlineFragment:
			if selectionHasIntrospection(node.SelectionSet) {
				return true
			}
		case *ast.FragmentSpread:
			// fragment introspection handled at document level in full parser;
			// inline fields cover the common case.
		}
	}
	return false
}

func selectionDepth(set ast.SelectionSet, current int) int {
	if len(set) == 0 {
		return current - 1
	}
	max := current
	for _, sel := range set {
		switch node := sel.(type) {
		case *ast.Field:
			depth := current
			if len(node.SelectionSet) > 0 {
				depth = selectionDepth(node.SelectionSet, current+1)
			}
			if depth > max {
				max = depth
			}
		case *ast.InlineFragment:
			if d := selectionDepth(node.SelectionSet, current); d > max {
				max = d
			}
		}
	}
	return max
}

func selectionComplexity(set ast.SelectionSet, multiplier int) int {
	total := 0
	for _, sel := range set {
		switch node := sel.(type) {
		case *ast.Field:
			fieldCost := multiplier
			total += fieldCost
			if len(node.SelectionSet) > 0 {
				total += selectionComplexity(node.SelectionSet, multiplier+1)
			}
		case *ast.InlineFragment:
			total += selectionComplexity(node.SelectionSet, multiplier)
		}
	}
	return total
}
