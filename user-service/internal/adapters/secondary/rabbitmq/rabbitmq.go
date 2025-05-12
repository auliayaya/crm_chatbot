package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQClient(url string) *RabbitMQClient {
	// Connect with retry logic
	var conn *amqp.Connection
	var err error

	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ: %s. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to RabbitMQ after 5 attempts: %s", err))
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("Failed to open a channel: %s", err))
	}

	// Declare exchanges
	err = ch.ExchangeDeclare(
		"user_events", // name
		"topic",       // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to declare an exchange: %s", err))
	}

	log.Println("Successfully connected to RabbitMQ")

	return &RabbitMQClient{
		conn:    conn,
		channel: ch,
	}
}

func (r *RabbitMQClient) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *RabbitMQClient) PublishUserEvent(eventType string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.channel.PublishWithContext(
		ctx,
		"user_events",       // exchange
		eventType,           // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers: amqp.Table{
				"event_type": eventType,
			},
		},
	)
}

func (r *RabbitMQClient) ConsumeUserEvents(eventType string, handler func([]byte)) error {
	q, err := r.channel.QueueDeclare(
		"",    // name (empty = auto-generated)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	err = r.channel.QueueBind(
		q.Name,        // queue name
		eventType,     // routing key
		"user_events", // exchange
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := r.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			handler(d.Body)
		}
	}()

	return nil
}