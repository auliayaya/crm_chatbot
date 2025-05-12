package messaging

import (
	"errors"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQClient manages connection to RabbitMQ
type RabbitMQClient struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
}

// NewRabbitMQClient creates a new RabbitMQ client with connection retry
func NewRabbitMQClient(amqpURL string) (*RabbitMQClient, error) {
	var conn *amqp.Connection
	var err error

	// Try to connect with exponential backoff
	maxRetries := 5
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(amqpURL)
		if err == nil {
			break // Successfully connected
		}

		if i < maxRetries-1 {
			log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v. Retrying in %v...",
				i+1, maxRetries, err, retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w",
			maxRetries, err)
	}

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Create exchange
	exchangeName := "crm_events"
	err = channel.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	client := &RabbitMQClient{
		conn:         conn,
		channel:      channel,
		exchangeName: exchangeName,
	}

	// Set up connection recovery
	go client.monitorConnection(amqpURL)

	log.Println("Connected to RabbitMQ successfully")
	return client, nil
}

// monitorConnection watches for connection issues and attempts reconnection
func (c *RabbitMQClient) monitorConnection(amqpURL string) {
	connErrChan := make(chan *amqp.Error)
	c.conn.NotifyClose(connErrChan)

	for {
		err := <-connErrChan
		if err != nil {
			log.Printf("RabbitMQ connection lost: %v. Attempting to reconnect...", err)

			// Try to reconnect with backoff
			backoff := time.Second
			maxBackoff := 30 * time.Second
			for {
				time.Sleep(backoff)

				newClient, err := NewRabbitMQClient(amqpURL)
				if err == nil {
					// Successfully reconnected, update connection and channel
					c.conn = newClient.conn
					c.channel = newClient.channel
					c.exchangeName = newClient.exchangeName

					// Reset monitoring on new connection
					connErrChan = make(chan *amqp.Error)
					c.conn.NotifyClose(connErrChan)

					log.Println("Successfully reconnected to RabbitMQ")
					break
				}

				log.Printf("Failed to reconnect to RabbitMQ: %v. Retrying in %v...",
					err, backoff)

				// Increase backoff with cap
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
		}
	}
}

// PublishMessage publishes a message to the specified routing key
func (c *RabbitMQClient) PublishMessage(routingKey string, message []byte) error {
	if c.channel == nil {
		return errors.New("channel not initialized")
	}

	// Create message
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         message,
	}

	// Publish
	return c.channel.Publish(
		c.exchangeName, // exchange
		routingKey,     // routing key
		false,          // mandatory
		false,          // immediate
		msg,            // message
	)
}

// Close closes the connection
func (c *RabbitMQClient) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected checks if the client is connected
func (c *RabbitMQClient) IsConnected() bool {
	return c.conn != nil && !c.conn.IsClosed()
}
