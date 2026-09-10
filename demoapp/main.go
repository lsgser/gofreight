package main

/*
|--------------------------------------------------------------------------
| Application Entry Point
|--------------------------------------------------------------------------
|
| Application main: loads bootstrap, registers routes, and starts the HTTP
| server.
| 
| Use gofreight serve for development; production uses a binary from
| gofreight build.
| 
*/

import (
	"log"

	"demoapp/bootstrap"
	"demoapp/routes"
)

func main() {
	app := bootstrap.Application()

	if err := app.ConnectDatabase(); err != nil {
		log.Printf("warning: database not connected: %v", err)
	}

	app.Draw(routes.Register)
	app.Run()
}
