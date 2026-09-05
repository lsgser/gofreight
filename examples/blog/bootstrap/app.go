package bootstrap

/*
|--------------------------------------------------------------------------
| Application Bootstrap
|--------------------------------------------------------------------------
|
| Configure middleware, sessions, queues, locale, and service bindings.
| Loaded on every request via main.go.
|
*/

import (
	"github.com/lsgser/gofreight/application"
	"github.com/lsgser/gofreight/config"
)

// Application creates and configures the Gofreight application instance.
func Application() *application.Application {
	app := application.New()

	if config.ResolveSessionDriver() == "redis" {
		_ = app.UseRedisSessions(config.ResolveRedisURL())
	}
	if config.ResolveQueueConnection() == "redis" {
		_ = app.UseRedisQueue(config.ResolveRedisURL())
	}

	_ = app.LoadLocales("config/locales")
	app.UseLocale()

	return app
}
