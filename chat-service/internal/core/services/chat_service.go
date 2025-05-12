// internal/core/services/chat_service.go
package services

import (
	"chat-service/internal/core/domain"
	"chat-service/internal/core/ports"
	"context"
	"time"

	"github.com/google/uuid"
)

type ChatServiceImpl struct {
	messageRepo      ports.MessageRepository
	conversationRepo ports.ConversationRepository
	messagePublisher ports.MessagePublisher
}

func NewChatService(
	messageRepo ports.MessageRepository,
	conversationRepo ports.ConversationRepository,
	messagePublisher ports.MessagePublisher,
) ports.ChatService {
	return &ChatServiceImpl{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
		messagePublisher: messagePublisher,
	}
}

func (s *ChatServiceImpl) SaveMessage(message *domain.Message) error {
	// Generate ID if not provided
	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	ctx := context.Background()

	// Save to repository
	err := s.messageRepo.SaveMessage(ctx, message)
	if err != nil {
		return err
	}

	// Publish to message broker
	return s.messagePublisher.PublishChatMessage(message)
}

func (s *ChatServiceImpl) GetChatHistory(customerID string) ([]domain.Message, error) {
	ctx := context.Background()
	return s.messageRepo.GetMessagesByCustomer(ctx, customerID)
}

func (s *ChatServiceImpl) GetConversation(conversationID string) (*domain.Conversation, error) {
	ctx := context.Background()
	return s.conversationRepo.GetConversation(ctx, conversationID)
}

func (s *ChatServiceImpl) CreateConversation(customerID string) (*domain.Conversation, error) {
	ctx := context.Background()
	existing, err := s.conversationRepo.GetActiveConversationByCustomer(ctx, customerID)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Create new conversation
	conversation := &domain.Conversation{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		StartedAt:  time.Now(),
		Status:     "active",
	}

	err = s.conversationRepo.CreateConversation(ctx, conversation)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func (s *ChatServiceImpl) CloseConversation(conversationID string) error {
	ctx := context.Background()
	conversation, err := s.conversationRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return err
	}

	conversation.Status = "closed"
	conversation.EndedAt = time.Now()

	return s.conversationRepo.UpdateConversation(ctx, conversation)
}

// Add this method to your ChatService struct
func (s *ChatServiceImpl) SubscribeToMessages(handler func(*domain.Message)) error {
	// Pass through to the message publisher
	// Assuming messagePublisher has a SubscribeToMessages method
	return s.messagePublisher.SubscribeToMessages(handler)
}
