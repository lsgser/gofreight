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
	"encoding/json"
	"os"
	"sync"

	"demoapp/app/services"
	appgraphql "demoapp/graphql"
	"github.com/lsgser/gofreight/application"
	"github.com/lsgser/gofreight/channels"
)

var chatMu sync.Mutex
var chatMessages []map[string]string

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

	if os.Getenv("SESSION_DRIVER") == "redis" {
		_ = app.UseRedisSessions(os.Getenv("REDIS_URL"))
	}

	if os.Getenv("QUEUE_DRIVER") == "redis" {
		_ = app.UseRedisQueue(os.Getenv("REDIS_URL"))
	}

	_ = app.LoadLocales("config/locales")
	app.UseLocale()
	app.UseCSRF()

	app.Singleton("example", func() any {
		return services.NewExampleService()
	})

	wireRealtime(app)
	appgraphql.Mount(app)

	return app
}

func wireRealtime(app *application.Application) {
	app.Channels.OnConnect(func(c *channels.Connection) {
		c.Join("chat:lobby")
	})

	app.Channels.On("chat:history", func(c *channels.Connection, _ json.RawMessage) {
		chatMu.Lock()
		history := append([]map[string]string(nil), chatMessages...)
		chatMu.Unlock()
		app.Channels.To("chat:lobby").Emit("chat:history", history)
	})

	app.Channels.On("chat:message", func(c *channels.Connection, raw json.RawMessage) {
		var payload struct {
			Text string `json:"text"`
			User string `json:"user"`
		}
		if json.Unmarshal(raw, &payload) != nil || payload.Text == "" {
			return
		}
		if payload.User == "" {
			payload.User = "guest"
		}

		msg := map[string]string{
			"text": payload.Text,
			"user": payload.User,
			"id":   c.ID,
		}

		chatMu.Lock()
		chatMessages = append(chatMessages, msg)
		if len(chatMessages) > 50 {
			chatMessages = chatMessages[len(chatMessages)-50:]
		}
		chatMu.Unlock()

		app.Channels.To("chat:lobby").Emit("chat:message", msg)
	})

	app.MountSocket("/socket")
}
