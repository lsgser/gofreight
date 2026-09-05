package main

/*
|--------------------------------------------------------------------------
| Application Entry Point
|--------------------------------------------------------------------------
|
| This file bootstraps the Gofreight application and starts the HTTP server.
| Application wiring lives in bootstrap/app.go. Routes are registered from
| the routes/ directory via routes.Register.
|
| Run locally:
|   gofreight serve
|   gofreight dev
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
