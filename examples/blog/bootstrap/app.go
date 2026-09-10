package bootstrap

/*
|--------------------------------------------------------------------------
| App
|--------------------------------------------------------------------------
|
| Implements App as part of the bootstrap package in the Gofreight
| framework. Key symbols: Application.
| 
| Symbols defined here include: Application (/*
| |--------------------------------------------------------------------------
| | Application
| |--------------------------------------------------------------------------
| | | Creates and configures the Gofreight application instance. | */).
| 
*/

import (
	"github.com/lsgser/gofreight/application"
	"github.com/lsgser/gofreight/config"
)

/*
|--------------------------------------------------------------------------
| Application
|--------------------------------------------------------------------------
|
| Creates and configures the Gofreight application instance.
|
*/
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
	app.UseCSRF()

	return app
}
