package main

import (
	"log"

	"main.go/config"
	"main.go/httpserver"
)

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	httpserver.Start(conf.Port)
}
