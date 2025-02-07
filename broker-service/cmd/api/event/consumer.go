package event

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{conn: conn}

	if err := consumer.setup(); err != nil {
		log.Println("Failed to setup consumer", err)
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		log.Println("Failed to open a channel", err)
		return err
	}

	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) (err error) {
	var (
		channel  *amqp.Channel
		queue    amqp.Queue
		messages <-chan amqp.Delivery
		forever  = make(chan bool)
	)

	if channel, err = consumer.conn.Channel(); err != nil {
		log.Println("Failed to open a channel", err)
		return err
	}
	defer channel.Close()

	if queue, err = declareRandomQueue(channel); err != nil {
		log.Println("Failed to declare a queue", err)
		return err
	}

	for _, s := range topics {
		if err = channel.QueueBind(
			queue.Name,
			s,
			"logs_topic",
			false,
			nil,
		); err != nil {
			log.Println("Failed to bind a queue", err)
			return err
		}
	}

	if messages, err = channel.Consume(queue.Name, "", true, false, false, false, nil); err != nil {
		return err
	}

	go func() {
		for d := range messages {
			var payload Payload
			if err = json.Unmarshal(d.Body, &payload); err != nil {
				log.Println("Failed to unmarshal JSON", err)
				continue
			}

			go handlePayload(payload)
		}
	}()

	log.Printf("Waiting for message [Exchange, Queue] [logs_topic, %s]\n", queue.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		if err := logEvent(payload); err != nil {
			log.Println("Failed to log event", err)
		}

	case "auth":
		// authenticate

	// you can have as many cases as you want, as long as you write the logic

	default:
		if err := logEvent(payload); err != nil {
			log.Println("Failed to log event", err)
		}
	}
}

func logEvent(entry Payload) (err error) {
	var (
		jsonData      []byte
		logServiceURL = "http://logger-service:9092/log"
		client        = &http.Client{}
		request       *http.Request
		response      *http.Response
	)

	if jsonData, err = json.MarshalIndent(entry, "", "\t"); err != nil {
		log.Println("Failed to marshal JSON", err)
		return err
	}

	if request, err = http.NewRequest(http.MethodPost, logServiceURL, bytes.NewBuffer(jsonData)); err != nil {
		log.Println("Failed to create a new request", err)
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	if response, err = client.Do(request); err != nil {
		log.Println("Failed to send request", err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.Println("Unexpected status code", response.StatusCode)
		return errors.New("unexpected status code")
	}

	return nil
}
