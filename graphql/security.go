package graphql

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

// SecurityConfig controls query validation before execution.
type SecurityConfig struct {
	MaxDepth           int // 0 = unlimited
	MaxComplexity      int // 0 = unlimited
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

// ParseQuery parses a GraphQL operation document including fragments.
func ParseQuery(query string) (*ast.QueryDocument, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("graphql: empty query")
	}
	return parser.ParseQuery(&ast.Source{Input: query})
}

// ValidateQuery checks depth, complexity, and introspection rules.
// Named fragments, inline fragments, and fragment spreads are fully expanded.
func ValidateQuery(query string, cfg SecurityConfig) error {
	doc, err := ParseQuery(query)
	if err != nil {
		return fmt.Errorf("graphql: invalid query: %w", err)
	}
	return ValidateQueryDocument(doc, cfg)
}

// ValidateQueryDocument validates a parsed query document.
func ValidateQueryDocument(doc *ast.QueryDocument, cfg SecurityConfig) error {
	if doc == nil {
		return fmt.Errorf("graphql: empty query document")
	}

	if !cfg.AllowIntrospection && documentHasIntrospection(doc) {
		return fmt.Errorf("graphql: introspection is disabled")
	}

	for _, op := range doc.Operations {
		if cfg.MaxDepth > 0 {
			if depth := selectionDepth(doc.Fragments, op.SelectionSet, 1, nil); depth > cfg.MaxDepth {
				return fmt.Errorf("graphql: query depth %d exceeds limit %d", depth, cfg.MaxDepth)
			}
		}
		if cfg.MaxComplexity > 0 {
			if score := selectionComplexity(doc.Fragments, op.SelectionSet, 1, nil); score > cfg.MaxComplexity {
				return fmt.Errorf("graphql: query complexity %d exceeds limit %d", score, cfg.MaxComplexity)
			}
		}
	}

	return nil
}

func documentHasIntrospection(doc *ast.QueryDocument) bool {
	for _, op := range doc.Operations {
		if selectionHasIntrospection(doc.Fragments, op.SelectionSet, nil) {
			return true
		}
	}
	for _, frag := range doc.Fragments {
		if selectionHasIntrospection(doc.Fragments, frag.SelectionSet, nil) {
			return true
		}
	}
	return false
}

func selectionHasIntrospection(fragments ast.FragmentDefinitionList, set ast.SelectionSet, spreadStack []string) bool {
	for _, sel := range set {
		switch node := sel.(type) {
		case *ast.Field:
			if node.Name == "__schema" || node.Name == "__type" {
				return true
			}
			if selectionHasIntrospection(fragments, node.SelectionSet, spreadStack) {
				return true
			}
		case *ast.InlineFragment:
			if selectionHasIntrospection(fragments, node.SelectionSet, spreadStack) {
				return true
			}
		case *ast.FragmentSpread:
			if fragmentSpreadHasIntrospection(fragments, node, spreadStack) {
				return true
			}
		}
	}
	return false
}

func fragmentSpreadHasIntrospection(fragments ast.FragmentDefinitionList, spread *ast.FragmentSpread, spreadStack []string) bool {
	if spread == nil {
		return false
	}
	if fragmentOnStack(spreadStack, spread.Name) {
		return false
	}
	frag := fragmentForSpread(fragments, spread)
	if frag == nil {
		return false
	}
	return selectionHasIntrospection(fragments, frag.SelectionSet, append(spreadStack, spread.Name))
}

func selectionDepth(fragments ast.FragmentDefinitionList, set ast.SelectionSet, current int, spreadStack []string) int {
	if len(set) == 0 {
		return current - 1
	}
	max := current
	for _, sel := range set {
		switch node := sel.(type) {
		case *ast.Field:
			depth := current
			if len(node.SelectionSet) > 0 {
				depth = selectionDepth(fragments, node.SelectionSet, current+1, spreadStack)
			}
			if depth > max {
				max = depth
			}
		case *ast.InlineFragment:
			if d := selectionDepth(fragments, node.SelectionSet, current, spreadStack); d > max {
				max = d
			}
		case *ast.FragmentSpread:
			if d := fragmentSpreadDepth(fragments, node, current, spreadStack); d > max {
				max = d
			}
		}
	}
	return max
}

func fragmentSpreadDepth(fragments ast.FragmentDefinitionList, spread *ast.FragmentSpread, current int, spreadStack []string) int {
	if spread == nil || fragmentOnStack(spreadStack, spread.Name) {
		return current
	}
	frag := fragmentForSpread(fragments, spread)
	if frag == nil {
		return current
	}
	return selectionDepth(fragments, frag.SelectionSet, current, append(spreadStack, spread.Name))
}

func selectionComplexity(fragments ast.FragmentDefinitionList, set ast.SelectionSet, multiplier int, spreadStack []string) int {
	total := 0
	for _, sel := range set {
		switch node := sel.(type) {
		case *ast.Field:
			total += multiplier
			if len(node.SelectionSet) > 0 {
				total += selectionComplexity(fragments, node.SelectionSet, multiplier+1, spreadStack)
			}
		case *ast.InlineFragment:
			total += selectionComplexity(fragments, node.SelectionSet, multiplier, spreadStack)
		case *ast.FragmentSpread:
			total += fragmentSpreadComplexity(fragments, node, multiplier, spreadStack)
		}
	}
	return total
}

func fragmentSpreadComplexity(fragments ast.FragmentDefinitionList, spread *ast.FragmentSpread, multiplier int, spreadStack []string) int {
	if spread == nil || fragmentOnStack(spreadStack, spread.Name) {
		return 0
	}
	frag := fragmentForSpread(fragments, spread)
	if frag == nil {
		return 0
	}
	return selectionComplexity(fragments, frag.SelectionSet, multiplier, append(spreadStack, spread.Name))
}

func fragmentForSpread(fragments ast.FragmentDefinitionList, spread *ast.FragmentSpread) *ast.FragmentDefinition {
	if spread.Definition != nil {
		return spread.Definition
	}
	return fragments.ForName(spread.Name)
}

func fragmentOnStack(stack []string, name string) bool {
	for _, n := range stack {
		if n == name {
			return true
		}
	}
	return false
}
