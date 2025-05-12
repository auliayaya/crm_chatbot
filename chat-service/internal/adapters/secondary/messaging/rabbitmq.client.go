// internal/adapters/secondary/messaging/rabbitmq_client.go
package messaging

import (
	"chat-service/internal/core/domain"
	"chat-service/internal/core/ports"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn    AMQPConnection // Interface instead of concrete *amqp.Connection
	channel AMQPChannel    // Interface instead of concrete *amqp.Channel
}
var _ ports.MessagePublisher = (*RabbitMQClient)(nil)


// NewRabbitMQClient creates a new client
func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Set up exchange
	err = ch.ExchangeDeclare(
		"chat_events", // name
		"topic",       // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	log.Println("Successfully connected to RabbitMQ")

	return &RabbitMQClient{
		conn:    conn, // amqp.Connection satisfies AMQPConnection
		channel: ch,   // amqp.Channel satisfies AMQPChannel
	}, nil
}

// NewRabbitMQClientWithDependencies creates a client with provided dependencies (for testing)
func NewRabbitMQClientWithDependencies(conn AMQPConnection, ch AMQPChannel) *RabbitMQClient {
	return &RabbitMQClient{
		conn:    conn,
		channel: ch,
	}
}

func (r *RabbitMQClient) PublishChatMessage(message *domain.Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return r.channel.Publish(
		"chat_events",   // exchange
		"chat.messages", // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}

func (r *RabbitMQClient) SubscribeToMessages(handler func(*domain.Message)) error {
	q, err := r.channel.QueueDeclare(
		"chat_service_queue", // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		return err
	}

	err = r.channel.QueueBind(
		q.Name,          // queue name
		"bot.responses", // routing key
		"chat_events",   // exchange
		false,           // no-wait
		nil,             // arguments
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
			var message domain.Message
			if err := json.Unmarshal(d.Body, &message); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			handler(&message)
		}
	}()

	return nil
}

func (r *RabbitMQClient) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

// Ensure RabbitMQClient implements MessagePublisher interface
var _ ports.MessagePublisher = (*RabbitMQClient)(nil)
