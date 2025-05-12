// internal/core/ports/primary.go
package ports

import "chat-service/internal/core/domain"


type ChatService interface {
	SaveMessage(message *domain.Message) error
	GetChatHistory(customerID string) ([]domain.Message, error)
	GetConversation(conversationID string) (*domain.Conversation, error)
	CreateConversation(customerID string) (*domain.Conversation, error)
	CloseConversation(conversationID string) error
	SubscribeToMessages(handler func(*domain.Message)) error
}
