package bootstrap

/*
|--------------------------------------------------------------------------
| Application Bootstrap
|--------------------------------------------------------------------------
|
| Configure the Gofreight application here: middleware, sessions, queues,
| locale, and service container bindings. This file is loaded on every
| request via main.go.
|
| Register services:
|   app.Singleton("payment", func() any { return services.NewPaymentService() })
|
| Resolve in controllers:
|   svc := app.Make("payment").(*services.PaymentService)
|
*/

import (
	"os"

	"demoapp/app/services"
	"github.com/lsgser/gofreight/application"
)

// Application creates and configures the Gofreight application instance.
func Application() *application.Application {
	app := application.New()

	// Session driver: memory (default) or redis (set SESSION_DRIVER=redis).
	if os.Getenv("SESSION_DRIVER") == "redis" {
		_ = app.UseRedisSessions(os.Getenv("REDIS_URL"))
	}

	// Background jobs: memory (default) or redis (set QUEUE_DRIVER=redis).
	if os.Getenv("QUEUE_DRIVER") == "redis" {
		_ = app.UseRedisQueue(os.Getenv("REDIS_URL"))
	}

	_ = app.LoadLocales("config/locales")
	app.UseLocale()

	// Service container — register application services here.
	app.Singleton("example", func() any {
		return services.NewExampleService()
	})

	return app
}
