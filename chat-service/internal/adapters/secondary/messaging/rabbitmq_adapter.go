// internal/adapters/secondary/messaging/rabbitmq_adapter.go
package messaging

import (
	"chat-service/internal/core/domain"
	"chat-service/internal/core/ports"
)

type RabbitMQAdapter struct {
	client *RabbitMQClient
}

// Ensure it implements the interface
var _ ports.MessagePublisher = (*RabbitMQAdapter)(nil)

func NewRabbitMQAdapter(client *RabbitMQClient) *RabbitMQAdapter {
	return &RabbitMQAdapter{client: client}
}

func (a *RabbitMQAdapter) PublishChatMessage(message *domain.Message) error {
	return a.client.PublishChatMessage(message)
}

func (a *RabbitMQAdapter) SubscribeToMessages(handler func(*domain.Message)) error {
	return a.client.SubscribeToMessages(handler)
}

// Close closes the underlying RabbitMQ client connection.
func (a *RabbitMQAdapter) Close() {
	a.client.Close()
}
