package graphql

import (
	"fmt"
	"strings"

	gql "github.com/graphql-go/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

// GQL parses and validates a GraphQL SDL string (like the gql template tag in JavaScript/TypeScript).
// Returns the trimmed SDL on success. Use with ModuleFromSDL to build a schema module.
func GQL(sdl string) (string, error) {
	sdl = strings.TrimSpace(sdl)
	if sdl == "" {
		return "", fmt.Errorf("graphql: empty SDL")
	}
	if _, err := parser.ParseSchema(&ast.Source{Input: sdl}); err != nil {
		return "", fmt.Errorf("graphql: invalid SDL: %w", err)
	}
	return sdl, nil
}

// MustGQL calls GQL and panics on error (for package-level schema constants).
func MustGQL(sdl string) string {
	out, err := GQL(sdl)
	if err != nil {
		panic(err)
	}
	return out
}

// ParseSDL parses a GraphQL Schema Definition Language document.
func ParseSDL(sdl string) (*ast.SchemaDocument, error) {
	sdl = strings.TrimSpace(sdl)
	if sdl == "" {
		return nil, fmt.Errorf("graphql: empty SDL")
	}
	return parser.ParseSchema(&ast.Source{Input: sdl})
}

// SDLResolvers binds Go resolver functions to SDL-defined fields.
type SDLResolvers struct {
	Query        map[string]gql.FieldResolveFn
	Mutation     map[string]gql.FieldResolveFn
	Subscription map[string]gql.FieldResolveFn
	// TypeFields maps "TypeName.fieldName" to resolver functions for object fields.
	TypeFields map[string]gql.FieldResolveFn
}

// SDLModuleConfig defines a GraphQL module from SDL plus resolver bindings.
type SDLModuleConfig struct {
	ID          string
	Description string
	SDL         string
	Resolvers   SDLResolvers
	// Scalars maps custom scalar names to graphql-go scalar configs.
	Scalars map[string]gql.ScalarConfig
}

// ModuleFromSDL builds a Module from an SDL string and resolver bindings.
func ModuleFromSDL(cfg SDLModuleConfig) (Module, error) {
	if cfg.ID == "" {
		return Module{}, fmt.Errorf("graphql: module ID is required")
	}
	doc, err := ParseSDL(cfg.SDL)
	if err != nil {
		return Module{}, err
	}

	typeMap, err := buildTypesFromSDL(doc, cfg.Resolvers, cfg.Scalars)
	if err != nil {
		return Module{}, err
	}

	types := make([]gql.Type, 0, len(typeMap))
	for name, t := range typeMap {
		if isRootType(name) {
			continue
		}
		types = append(types, t)
	}

	queryFields, mutationFields, subscriptionFields, err := rootFieldsFromSDL(doc, typeMap, cfg.Resolvers)
	if err != nil {
		return Module{}, err
	}

	return CreateModule(ModuleConfig{
		ID:           cfg.ID,
		Description:  cfg.Description,
		Types:        types,
		Query:        queryFields,
		Mutation:     mutationFields,
		Subscription: subscriptionFields,
	})
}

func isRootType(name string) bool {
	return name == "Query" || name == "Mutation" || name == "Subscription"
}

func buildTypesFromSDL(doc *ast.SchemaDocument, resolvers SDLResolvers, scalars map[string]gql.ScalarConfig) (map[string]gql.Type, error) {
	defs := append([]*ast.Definition{}, doc.Definitions...)
	defs = append(defs, doc.Extensions...)

	names := map[string]*ast.Definition{}
	for _, def := range defs {
		if def == nil || isRootType(def.Name) {
			continue
		}
		if _, exists := names[def.Name]; exists {
			return nil, fmt.Errorf("graphql: duplicate type %s in SDL", def.Name)
		}
		names[def.Name] = def
	}

	out := map[string]gql.Type{
		"Query":        nil,
		"Mutation":       nil,
		"Subscription":   nil,
		"String":         gql.String,
		"Int":            gql.Int,
		"Float":          gql.Float,
		"Boolean":        gql.Boolean,
		"ID":             gql.ID,
	}

	for name, def := range names {
		switch def.Kind {
		case ast.Scalar:
			if isBuiltInScalar(name) {
				continue
			}
			if cfg, ok := scalars[name]; ok {
				out[name] = gql.NewScalar(cfg)
			} else {
				out[name] = gql.NewScalar(gql.ScalarConfig{
					Name:        name,
					Description: def.Description,
					Serialize:   serializePassThrough,
				})
			}
		case ast.Enum:
			values := gql.EnumValueConfigMap{}
			for _, v := range def.EnumValues {
				values[v.Name] = &gql.EnumValueConfig{
					Value:       v.Name,
					Description: v.Description,
				}
			}
			out[name] = gql.NewEnum(gql.EnumConfig{
				Name:        name,
				Description: def.Description,
				Values:      values,
			})
		case ast.InputObject:
			fields := gql.InputObjectConfigFieldMap{}
			for _, f := range def.Fields {
				t, err := resolveOutputType(f.Type, out, names)
				if err != nil {
					return nil, fmt.Errorf("graphql: input %s.%s: %w", name, f.Name, err)
				}
				in, ok := t.(gql.Input)
				if !ok {
					return nil, fmt.Errorf("graphql: input %s.%s: type is not an input type", name, f.Name)
				}
				fields[f.Name] = &gql.InputObjectFieldConfig{
					Type:        in,
					Description: f.Description,
				}
			}
			out[name] = gql.NewInputObject(gql.InputObjectConfig{
				Name:        name,
				Description: def.Description,
				Fields:      fields,
			})
		case ast.Object:
			// Built lazily below after all names are registered.
		case ast.Interface, ast.Union:
			return nil, fmt.Errorf("graphql: SDL type %s (%s) — define interfaces and unions programmatically for now", name, def.Kind)
		default:
			return nil, fmt.Errorf("graphql: unsupported SDL kind %s for %s", def.Kind, name)
		}
	}

	for name, def := range names {
		if def.Kind != ast.Object {
			continue
		}
		obj, err := buildObjectType(def, out, names, resolvers.TypeFields)
		if err != nil {
			return nil, err
		}
		out[name] = obj
	}

	return out, nil
}

func buildObjectType(def *ast.Definition, built map[string]gql.Type, pending map[string]*ast.Definition, typeFields map[string]gql.FieldResolveFn) (gql.Type, error) {
	defCopy := def
	fieldResolvers := typeFields
	typeMap := built
	pendingMap := pending

	return gql.NewObject(gql.ObjectConfig{
		Name:        defCopy.Name,
		Description: defCopy.Description,
		Fields: gql.FieldsThunk(func() gql.Fields {
			fields := gql.Fields{}
			for _, f := range defCopy.Fields {
				output, err := resolveOutputType(f.Type, typeMap, pendingMap)
				if err != nil {
					panic(fmt.Sprintf("graphql: %s.%s: %v", defCopy.Name, f.Name, err))
				}
				key := defCopy.Name + "." + f.Name
				resolve := fieldResolvers[key]
				args := argumentConfigs(f.Arguments, typeMap, pendingMap)
				fields[f.Name] = &gql.Field{
					Type:        output,
					Args:        args,
					Resolve:     resolve,
					Description: f.Description,
				}
			}
			return fields
		}),
	}), nil
}

func rootFieldsFromSDL(doc *ast.SchemaDocument, typeMap map[string]gql.Type, resolvers SDLResolvers) (query, mutation, subscription gql.Fields, err error) {
	query = gql.Fields{}
	mutation = gql.Fields{}
	subscription = gql.Fields{}

	addRoot := func(rootName string, target gql.Fields, resolverMap map[string]gql.FieldResolveFn) error {
		for _, ext := range doc.Extensions {
			if ext.Name != rootName {
				continue
			}
			for _, f := range ext.Fields {
				output, err := resolveOutputType(f.Type, typeMap, nil)
				if err != nil {
					return fmt.Errorf("graphql: %s.%s: %w", rootName, f.Name, err)
				}
				resolve := resolverMap[f.Name]
				if resolve == nil {
					return fmt.Errorf("graphql: missing resolver for %s.%s", rootName, f.Name)
				}
				target[f.Name] = &gql.Field{
					Type:        output,
					Args:        argumentConfigs(f.Arguments, typeMap, nil),
					Resolve:     resolve,
					Description: f.Description,
				}
			}
		}
		return nil
	}

	if err := addRoot("Query", query, resolvers.Query); err != nil {
		return nil, nil, nil, err
	}
	if err := addRoot("Mutation", mutation, resolvers.Mutation); err != nil {
		return nil, nil, nil, err
	}
	if err := addRoot("Subscription", subscription, resolvers.Subscription); err != nil {
		return nil, nil, nil, err
	}

	if len(query) == 0 {
		return nil, nil, nil, fmt.Errorf("graphql: SDL module requires extend type Query with at least one field")
	}
	return query, mutation, subscription, nil
}

func argumentConfigs(args ast.ArgumentDefinitionList, built map[string]gql.Type, pending map[string]*ast.Definition) gql.FieldConfigArgument {
	if len(args) == 0 {
		return nil
	}
	out := gql.FieldConfigArgument{}
	for _, arg := range args {
		t, err := resolveOutputType(arg.Type, built, pending)
		if err != nil {
			continue
		}
		in, ok := t.(gql.Input)
		if !ok {
			continue
		}
		out[arg.Name] = &gql.ArgumentConfig{
			Type:         in,
			Description:  arg.Description,
			DefaultValue: defaultValue(arg.DefaultValue),
		}
	}
	return out
}

func defaultValue(v *ast.Value) any {
	if v == nil {
		return nil
	}
	if v.Raw != "" {
		return v.Raw
	}
	return v.String()
}

func resolveOutputType(t *ast.Type, built map[string]gql.Type, pending map[string]*ast.Definition) (gql.Output, error) {
	if t == nil {
		return nil, fmt.Errorf("nil type")
	}

	var inner gql.Output
	if t.Elem != nil {
		elem, err := resolveOutputType(t.Elem, built, pending)
		if err != nil {
			return nil, err
		}
		inner = gql.NewList(elem)
	} else {
		name := t.NamedType
		if isBuiltInScalar(name) {
			inner = builtInScalar(name)
		} else if typ, ok := built[name]; ok && typ != nil {
			out, ok := typ.(gql.Output)
			if !ok {
				return nil, fmt.Errorf("type %s is not an output type", name)
			}
			inner = out
		} else if def, ok := pending[name]; ok {
			switch def.Kind {
			case ast.Object, ast.Interface, ast.Union, ast.Enum, ast.Scalar:
				// Forward reference — type will be built; use lazy object if needed.
				if typ, ok := built[name]; ok && typ != nil {
					inner, _ = typ.(gql.Output)
				} else {
					return nil, fmt.Errorf("forward reference to %s not yet built", name)
				}
			case ast.InputObject:
				return nil, fmt.Errorf("input type %s used as output", name)
			default:
				return nil, fmt.Errorf("unknown type %s", name)
			}
		} else {
			return nil, fmt.Errorf("unknown type %s", name)
		}
	}

	if t.NonNull {
		return gql.NewNonNull(inner), nil
	}
	return inner, nil
}

func isBuiltInScalar(name string) bool {
	switch name {
	case "String", "Int", "Float", "Boolean", "ID":
		return true
	default:
		return false
	}
}

func builtInScalar(name string) gql.Output {
	switch name {
	case "String":
		return gql.String
	case "Int":
		return gql.Int
	case "Float":
		return gql.Float
	case "Boolean":
		return gql.Boolean
	case "ID":
		return gql.ID
	default:
		return gql.String
	}
}

func serializePassThrough(value any) any {
	return value
}

// SupportedSDLKinds documents GraphQL type kinds Gofreight can build from SDL strings.
var SupportedSDLKinds = []string{
	"Scalar (built-in: String, Int, Float, Boolean, ID; custom scalars via Scalars config)",
	"Object",
	"Enum",
	"InputObject",
	"List ([Type])",
	"NonNull (Type!)",
	"extend type Query / Mutation / Subscription",
}
