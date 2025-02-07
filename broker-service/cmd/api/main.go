package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "9090"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	rabbitConn, err := connectRabbitMQ()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	app := Config{
		Rabbit: rabbitConn,
	}

	log.Printf("Starting broker service on port %s\n", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	if err = srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

func connectRabbitMQ() (conn *amqp.Connection, err error) {
	var (
		counts  int64
		backOff = 1 * time.Second
		url     = "amqp://rabbitmq:password@rabbitmq"
	)

	for {
		if conn, err = amqp.Dial(url); err != nil {
			log.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			log.Println("Connected to RabbitMQ!")
			break
		}

		if counts > 5 {
			log.Println("Failed to connect to RabbitMQ after 5 attempts.")
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off for", backOff)
		time.Sleep(backOff)
		continue
	}

	return conn, nil
}
