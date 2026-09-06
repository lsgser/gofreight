package graphql

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/lsgser/gofreight/application"
	gfgraphql "github.com/lsgser/gofreight/graphql"
)

/*
|--------------------------------------------------------------------------
| Mount
|--------------------------------------------------------------------------
|
| Registers combined GraphQL modules on the application.
|
*/
func Mount(app *application.Application) {
	fieldLimits := gfgraphql.NewFieldRateLimitRegistry()
	fieldLimits.Set("Mutation", "createPost", gfgraphql.RateLimitRule{Limit: 30, Window: time.Minute})

	gqlApp, err := gfgraphql.CreateApplication(gfgraphql.ApplicationConfig{
		Modules:    Modules(),
		Playground: true,
		Path:       "/graphql",
		RateLimit:  120, /* global per IP per minute */
		FieldRateLimits: fieldLimits,
		Security:        gfgraphql.DefaultSecurity(),
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
