package publisher

import (
	"encoding/json"
	"log"
	"os"

	"eiffel-bridge-svc/internal/eiffel"

	"github.com/streadway/amqp"
)

type RabbitPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewRabbitPublisher() (*RabbitPublisher, error) {
	url := os.Getenv("RABBIT_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		"eiffel.events", // queue name
		true,            // durable
		false,           // auto-delete
		false,           // exclusive
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		return nil, err
	}

	return &RabbitPublisher{conn: conn, channel: ch, queue: q}, nil
}

func (p *RabbitPublisher) Publish(event eiffel.Event) error {
	body, _ := json.Marshal(event)
	err := p.channel.Publish(
		"",           // exchange (empty → default direct)
		p.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("failed to publish: %v", err)
	}
	return err
}

func (p *RabbitPublisher) Close() {
	p.channel.Close()
	p.conn.Close()
}
