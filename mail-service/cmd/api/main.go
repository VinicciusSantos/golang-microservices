package main

import (
	"fmt"
	"log"
	"net/http"
)

type Config struct{}

const webPort = 9093

func main() {
	app := Config{}

	log.Printf("Starting mail service on port %v", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", webPort),
		Handler: app.routes(),
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
