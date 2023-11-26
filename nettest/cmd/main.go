package main

import (
	"log"
	"nettest/internal/pkg/app"
)

func main() {
	//log.Println("Start")
	//api.StartServer()
	//log.Println("terminated")

	application := app.New()
	log.Println("app created")
	log.Println("run server")
	application.Run()
	log.Println("server terminated")
}
