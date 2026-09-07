package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GraphQL installs GraphQL scaffolding in the application.
func GraphQL(appPath string) error {
	module := moduleName(appPath)
	data := struct{ Module string }{Module: module}

	dir := filepath.Join(appPath, "graphql")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	files := map[string]string{
		"register.go": graphqlRegisterTmpl,
		"modules.go":  graphqlModulesTmpl,
		"loaders.go":  graphqlLoadersDocTmpl,
	}
	for name, tmpl := range files {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := writeTemplate(path, tmpl, data); err != nil {
			return err
		}
	}

	if err := patchGraphQLBootstrap(appPath, module); err != nil {
		return err
	}
	return ensureGraphQLGoMod(appPath)
}

// GraphQLModule generates a GraphQL module with queries, mutations, and loaders.
func GraphQLModule(appPath, name string, fields map[string]string) error {
	if _, err := os.Stat(filepath.Join(appPath, "graphql", "modules.go")); os.IsNotExist(err) {
		if err := GraphQL(appPath); err != nil {
			return err
		}
	}

	data := buildGraphQLModuleData(appPath, name, fields)
	path := filepath.Join(appPath, "graphql", data.FileName)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("graphql module already exists: %s", data.FileName)
	}
	if err := writeTemplate(path, graphqlModuleTmpl, data); err != nil {
		return err
	}
	if err := appendGraphQLModuleRegistration(appPath, data.Name); err != nil {
		return err
	}
	return ensureGraphQLGoMod(appPath)
}

type GraphQLModuleData struct {
	Module, Name, Plural, Singular, Title, ModuleID, FileName string
	ObjectFields                                              string
	CreateArgs                                                string
	CreateBody                                                string
}

func buildGraphQLModuleData(appPath, name string, fields map[string]string) GraphQLModuleData {
	plural := pluralize(strings.ToLower(name))
	singular := strings.ToLower(name)
	titleName := title(name)
	moduleID := singular

	var objectLines []string
	var createArgLines []string
	var createBodyLines []string

	objectLines = append(objectLines, "\t\t\"id\": &gql.Field{Type: gql.NewNonNull(gql.String)},")
	for fname, ftype := range fields {
		pf := ParseField(fname, ftype)
		gqlField := graphqlFieldKey(pf)
		gqlType := graphqlGQLExpr(pf, true)
		objectLines = append(objectLines, fmt.Sprintf("\t\t\"%s\": &gql.Field{Type: %s},", gqlField, gqlType))

		if pf.ReferenceTable != "" {
			continue
		}
		argType := graphqlGQLExpr(pf, graphqlFieldRequired(pf))
		createArgLines = append(createArgLines, fmt.Sprintf("\t\t\t\t\"%s\": &gql.ArgumentConfig{Type: %s},", gqlField, argType))
		createBodyLines = append(createBodyLines, fmt.Sprintf("\t\t\t\t\t\"%s\": %s,", gqlField, graphqlArgExtract(gqlField, pf)))
	}
	for fname, ftype := range fields {
		pf := ParseField(fname, ftype)
		if pf.ReferenceTable == "" {
			continue
		}
		gqlField := graphqlFieldKey(pf)
		argType := graphqlGQLExpr(pf, true)
		createArgLines = append(createArgLines, fmt.Sprintf("\t\t\t\t\"%s\": &gql.ArgumentConfig{Type: %s},", gqlField, argType))
		createBodyLines = append(createBodyLines, fmt.Sprintf("\t\t\t\t\t\"%s\": %s,", gqlField, graphqlArgExtract(gqlField, pf)))
	}

	return GraphQLModuleData{
		Module:       moduleName(appPath),
		Name:         titleName,
		Plural:       plural,
		Singular:     singular,
		Title:        title(plural),
		ModuleID:     moduleID,
		FileName:     singular + "_module.go",
		ObjectFields: strings.Join(objectLines, "\n"),
		CreateArgs:   strings.Join(createArgLines, "\n"),
		CreateBody:   strings.Join(createBodyLines, "\n"),
	}
}

func graphqlFieldKey(pf ParsedField) string {
	if pf.ReferenceTable != "" {
		return graphqlForeignKeyName(pf.DBTag)
	}
	return graphqlCamelCase(pf.DBTag)
}

func graphqlForeignKeyName(dbTag string) string {
	key := graphqlCamelCase(dbTag)
	if strings.HasSuffix(strings.ToLower(key), "id") {
		return key
	}
	return key + "Id"
}

func graphqlCamelCase(dbTag string) string {
	parts := strings.Split(dbTag, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 {
			parts[i] = strings.ToLower(p)
			continue
		}
		parts[i] = title(p)
	}
	return strings.Join(parts, "")
}

func graphqlFieldRequired(pf ParsedField) bool {
	switch pf.FormType {
	case "textarea", "checkbox":
		return false
	default:
		return pf.ReferenceTable != "" || pf.FormType != "textarea"
	}
}

func graphqlGQLExpr(pf ParsedField, required bool) string {
	base, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(pf.RawType)), ":")
	var scalar string
	switch base {
	case "int", "integer", "bigint", "references", "reference", "belongs_to":
		scalar = "gql.Int"
	case "float", "decimal", "double":
		scalar = "gql.Float"
	case "bool", "boolean":
		scalar = "gql.Boolean"
	default:
		scalar = "gql.String"
	}
	if required {
		return "gql.NewNonNull(" + scalar + ")"
	}
	return scalar
}

func graphqlArgExtract(gqlField string, pf ParsedField) string {
	base, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(pf.RawType)), ":")
	switch base {
	case "int", "integer", "bigint":
		return fmt.Sprintf(`p.Args["%s"].(int)`, gqlField)
	case "float", "decimal", "double":
		return fmt.Sprintf(`p.Args["%s"].(float64)`, gqlField)
	case "bool", "boolean":
		return fmt.Sprintf(`p.Args["%s"].(bool)`, gqlField)
	default:
		return fmt.Sprintf(`p.Args["%s"].(string)`, gqlField)
	}
}

func patchGraphQLBootstrap(appPath, module string) error {
	path := filepath.Join(appPath, "bootstrap", "app.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	if strings.Contains(content, "graphql.Mount(app)") {
		return nil
	}

	importLine := fmt.Sprintf("\tappgraphql %q\n", module+"/graphql")
	if !strings.Contains(content, module+"/graphql") {
		needle := "\t\"github.com/lsgser/gofreight/application\"\n"
		if strings.Contains(content, needle) {
			content = strings.Replace(content, needle, importLine+needle, 1)
		}
	}

	mountNeedle := "\n\treturn app\n"
	mountInsert := "\n\tappgraphql.Mount(app)\n\n\treturn app\n"
	if strings.Contains(content, mountNeedle) {
		content = strings.Replace(content, mountNeedle, mountInsert, 1)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func appendGraphQLModuleRegistration(appPath, name string) error {
	path := filepath.Join(appPath, "graphql", "modules.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	moduleLine := fmt.Sprintf("\tmodules = append(modules, %sModule())\n", name)
	if strings.Contains(content, moduleLine) {
		return nil
	}
	moduleMarker := "\t// gofreight:graphql-modules"
	if strings.Contains(content, moduleMarker) {
		content = strings.Replace(content, moduleMarker, moduleLine+"\t"+moduleMarker, 1)
	}

	loaderLine := fmt.Sprintf("\tregister%sLoaders(reg)\n", name)
	loaderMarker := "\t// gofreight:graphql-loaders"
	if strings.Contains(content, loaderMarker) && !strings.Contains(content, loaderLine) {
		content = strings.Replace(content, loaderMarker, loaderLine+"\t"+loaderMarker, 1)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func ensureGraphQLGoMod(appPath string) error {
	path := filepath.Join(appPath, "go.mod")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)
	var missing []string
	for _, dep := range []string{
		"github.com/graph-gophers/dataloader/v7 v7.2.0",
		"github.com/graphql-go/graphql v0.8.1",
	} {
		pkg := strings.Split(dep, " ")[0]
		if strings.Contains(content, pkg) {
			continue
		}
		missing = append(missing, dep)
	}
	if len(missing) == 0 {
		return nil
	}

	if strings.Contains(content, "require (") {
		needle := "require (\n"
		var addition strings.Builder
		for _, dep := range missing {
			addition.WriteString("\t" + dep + "\n")
		}
		content = strings.Replace(content, needle, needle+addition.String(), 1)
	} else {
		for _, dep := range missing {
			content = strings.TrimRight(content, "\n") + "\nrequire " + dep + "\n"
		}
	}
	return os.WriteFile(path, []byte(content), 0644)
}

const graphqlRegisterTmpl = `package graphql

import (
	"context"
	"log"
	"net/http"

	"github.com/lsgser/gofreight/application"
	gfgraphql "github.com/lsgser/gofreight/graphql"
)

/*
|--------------------------------------------------------------------------
| Mount
|--------------------------------------------------------------------------
|
| Registers combined GraphQL modules on the application.
| Playground: GET /graphql/playground
|
*/
func Mount(app *application.Application) {
	gqlApp, err := gfgraphql.CreateApplication(gfgraphql.ApplicationConfig{
		Modules:    Modules(),
		Playground: true,
		Path:       "/graphql",
		RateLimit:  120,
		Security:   gfgraphql.DefaultSecurity(),
		OnRequest: func(r *http.Request, loaders *gfgraphql.LoaderRegistry) context.Context {
			RegisterLoaders(loaders)
			return gfgraphql.DefaultOnRequest(r, loaders)
		},
	})
	if err != nil {
		log.Printf("graphql: %v", err)
		return
	}

	app.MountGraphQLApplication("/graphql", gqlApp)
}
`

const graphqlModulesTmpl = `package graphql

import (
	gfgraphql "github.com/lsgser/gofreight/graphql"
)

/*
|--------------------------------------------------------------------------
| Modules
|--------------------------------------------------------------------------
|
| Returns GraphQL modules composed for the application endpoint.
| Add modules with: gofreight make:graphql-module Post title:string body:text
|
*/
func Modules() []gfgraphql.Module {
	modules := []gfgraphql.Module{}
	// gofreight:graphql-modules
	return modules
}

/*
|--------------------------------------------------------------------------
| RegisterLoaders
|--------------------------------------------------------------------------
|
| Attaches dataloaders to each GraphQL request.
|
*/
func RegisterLoaders(reg *gfgraphql.LoaderRegistry) {
	// gofreight:graphql-loaders
}
`

const graphqlLoadersDocTmpl = `package graphql

/*
|--------------------------------------------------------------------------
| Loaders
|--------------------------------------------------------------------------
|
| Shared loader helpers live next to each module file generated by
| gofreight make:graphql-module. Each module registers its loader in
| RegisterLoaders inside modules.go.
|
*/
`

const graphqlModuleTmpl = `package graphql

import (
	"context"
	"fmt"
	"sync"
	"time"

	gql "github.com/graphql-go/graphql"
	"github.com/graph-gophers/dataloader/v7"
	gfgraphql "github.com/lsgser/gofreight/graphql"
)

var (
	{{.Plural}}Mu    sync.RWMutex
	{{.Plural}}Store = []map[string]any{}
)

/*
|--------------------------------------------------------------------------
| {{.Name}} Module
|--------------------------------------------------------------------------
|
| Generated GraphQL module for {{.Title}}.
| Wire to models by replacing the in-memory store helpers below.
|
*/
func {{.Name}}Module() gfgraphql.Module {
	{{.Singular}}Type := gql.NewObject(gql.ObjectConfig{
		Name: "{{.Name}}",
		Fields: gql.Fields{
{{.ObjectFields}}
		},
	})

	return gfgraphql.MustCreateModule(gfgraphql.ModuleConfig{
		ID:          "{{.ModuleID}}",
		Description: "{{.Title}} queries and mutations",
		Types:       []gql.Type{ {{.Singular}}Type },
		Query: gql.Fields{
			"{{.Plural}}": &gql.Field{
				Type: gql.NewList({{.Singular}}Type),
				Resolve: func(_ gql.ResolveParams) (any, error) {
					{{.Plural}}Mu.RLock()
					defer {{.Plural}}Mu.RUnlock()
					out := make([]map[string]any, len({{.Plural}}Store))
					copy(out, {{.Plural}}Store)
					return out, nil
				},
			},
			"{{.Singular}}": &gql.Field{
				Type: {{.Singular}}Type,
				Args: gql.FieldConfigArgument{
					"id": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.String)},
				},
				Resolve: func(p gql.ResolveParams) (any, error) {
					id, _ := p.Args["id"].(string)
					return load{{.Name}}(p.Context, id)
				},
			},
		},
		Mutation: gql.Fields{
			"create{{.Name}}": gfgraphql.Field(gfgraphql.FieldConfig{
				Type: {{.Singular}}Type,
				Args: gql.FieldConfigArgument{
{{.CreateArgs}}
				},
				Resolve: func(p gql.ResolveParams) (any, error) {
					{{.Plural}}Mu.Lock()
					defer {{.Plural}}Mu.Unlock()
					id := fmt.Sprintf("%d", len({{.Plural}}Store)+1)
					item := map[string]any{
						"id": id,
{{.CreateBody}}
					}
					{{.Plural}}Store = append({{.Plural}}Store, item)
					return item, nil
				},
			}, gfgraphql.WithFieldRateLimit(30, time.Minute)),
		},
	})
}

func register{{.Name}}Loaders(reg *gfgraphql.LoaderRegistry) {
	reg.Register("{{.ModuleID}}", func() any {
		return gfgraphql.NewLoader[string, map[string]any](func(ctx context.Context, keys []string) []*dataloader.Result[map[string]any] {
			results := make([]*dataloader.Result[map[string]any], len(keys))
			for i, key := range keys {
				results[i] = &dataloader.Result[map[string]any]{Data: find{{.Name}}(key)}
			}
			return results
		})
	})
}

func load{{.Name}}(ctx context.Context, id string) (map[string]any, error) {
	if loader, ok := gfgraphql.LoaderFromContext[string, map[string]any](ctx, "{{.ModuleID}}"); ok {
		return loader.Load(ctx, id)()
	}
	item := find{{.Name}}(id)
	if item == nil {
		return nil, fmt.Errorf("{{.Singular}} not found")
	}
	return item, nil
}

func find{{.Name}}(id string) map[string]any {
	{{.Plural}}Mu.RLock()
	defer {{.Plural}}Mu.RUnlock()
	for _, item := range {{.Plural}}Store {
		if item["id"] == id {
			return item
		}
	}
	return nil
}
`
