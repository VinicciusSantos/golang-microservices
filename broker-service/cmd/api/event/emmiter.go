package event

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	connection *amqp.Connection
}

func (e *Emitter) setup() error {
	ch, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return declareExchange(ch)
}

func (e *Emitter) Push(event, severity string) error {
	ch, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	log.Printf("Publishing event %s with severity %s\n", event, severity)

	return ch.Publish(
		"log_topic", // exchange
		severity,    // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		},
	)
}

func NewEventEmitter(conn *amqp.Connection) (emitter Emitter, err error) {
	emitter = Emitter{connection: conn}

	if err := emitter.setup(); err != nil {
		log.Println("Failed to setup emitter")
		return Emitter{}, err
	}

	return emitter, nil
}
