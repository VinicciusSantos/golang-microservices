package main

import (
	"log"
	"math"
	"os"
	"time"

	"github.com/VinicciusSantos/golang-microservices/listener-service/event"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	var (
		rabbitConn *amqp.Connection
		consumer   event.Consumer
		err        error
	)

	if rabbitConn, err = connect(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	log.Println("Listening for and consuming RabbitMQ messages...")

	if consumer, err = event.NewConsumer(rabbitConn); err != nil {
		panic(err)
	}

	if err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"}); err != nil {
		log.Println("Failed to listen for messages", err)
	}
}

func connect() (conn *amqp.Connection, err error) {
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
