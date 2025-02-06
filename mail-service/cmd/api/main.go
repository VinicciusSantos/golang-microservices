package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Config struct {
	Mailer Mail
}

const webPort = 9093

func main() {
	app := Config{
		Mailer: createMail(),
	}

	log.Printf("Starting mail service on port %v", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", webPort),
		Handler: app.routes(),
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

func createMail() Mail {
	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		log.Fatalf("failed to convert MAIL_PORT to int: %v", err)
	}

	return Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromName:    os.Getenv("MAIL_FROM_NAME"),
		FromAddress: os.Getenv("MAIL_FROM_ADDRESS"),
	}
}
