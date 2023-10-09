package main

import (
	"log"

	"nettest/internal/api"
)

func main() {
	log.Println("Start")
	api.StartServer()
	log.Println("terminated")
}
