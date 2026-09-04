package main

import (
	"log"

	"blog/config"
	"github.com/gofreight/gofreight/application"
)

func main() {
	app := application.New()

	if err := app.ConnectDatabase(); err != nil {
		log.Printf("warning: database not connected: %v", err)
	}

	app.Draw(config.Routes)
	app.Run()
}
