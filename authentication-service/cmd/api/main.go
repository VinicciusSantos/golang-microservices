package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/VinicciusSantos/golang-microservices/authentication/data"
)

const webPort = 7070

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting authentication service")

	// TODO: Connect to the database

	app := Config{}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", webPort),
		Handler: app.routes(),
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Panic("server failed to start:", err)
	}
}
