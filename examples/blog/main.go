package main

/*
|--------------------------------------------------------------------------
| Application Entry Point
|--------------------------------------------------------------------------
|
| Bootstraps the Gofreight application and starts the HTTP server.
| Wiring: bootstrap/app.go · Routes: routes/
|
*/

import (
	"log"

	"blog/bootstrap"
	"blog/routes"
)

func main() {
	app := bootstrap.Application()

	if err := app.ConnectDatabase(); err != nil {
		log.Printf("warning: database not connected: %v", err)
	}

	app.Draw(routes.Register)
	app.Run()
}
