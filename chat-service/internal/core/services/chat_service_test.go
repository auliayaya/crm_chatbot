// internal/core/services/chat_service_test.go
package services_test

import (
	"chat-service/internal/core/domain"
	"chat-service/internal/core/services"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories and publisher
type MockMessageRepo struct {
	mock.Mock
}

func (m *MockMessageRepo) SaveMessage(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepo) GetMessagesByCustomer(ctx context.Context, customerID string) ([]domain.Message, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]domain.Message), args.Error(1)
}

func (m *MockMessageRepo) GetMessagesByConversation(ctx context.Context, conversationID string) ([]domain.Message, error) {
	args := m.Called(ctx, conversationID)
	return args.Get(0).([]domain.Message), args.Error(1)
}

type MockConversationRepo struct {
	mock.Mock
}

func (m *MockConversationRepo) CreateConversation(ctx context.Context, conversation *domain.Conversation) error {
	args := m.Called(ctx, conversation)
	return args.Error(0)
}

func (m *MockConversationRepo) GetConversation(ctx context.Context, id string) (*domain.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockConversationRepo) GetActiveConversationByCustomer(ctx context.Context, customerID string) (*domain.Conversation, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockConversationRepo) UpdateConversation(ctx context.Context, conversation *domain.Conversation) error {
	args := m.Called(ctx, conversation)
	return args.Error(0)
}

type MockMessagePublisher struct {
	mock.Mock
}

func (m *MockMessagePublisher) PublishChatMessage(message *domain.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMessagePublisher) SubscribeToMessages(handler func(*domain.Message)) error {
	args := m.Called(handler)
	return args.Error(0)
}

func (m *MockMessagePublisher) Close() {
	m.Called()
}

func TestSaveMessage(t *testing.T) {
	messageRepo := new(MockMessageRepo)
	conversationRepo := new(MockConversationRepo)
	publisher := new(MockMessagePublisher)

	service := services.NewChatService(messageRepo, conversationRepo, publisher)

	t.Run("success", func(t *testing.T) {
		message := &domain.Message{
			Content:    "Hello",
			UserID:     "user123",
			CustomerID: "customer456",
			Type:       domain.UserMessage,
		}

		messageRepo.On("SaveMessage", mock.Anything, mock.MatchedBy(func(m *domain.Message) bool {
			return m.Content == "Hello" && m.UserID == "user123"
		})).Return(nil).Once()

		publisher.On("PublishChatMessage", mock.MatchedBy(func(m *domain.Message) bool {
			return m.Content == "Hello" && m.UserID == "user123"
		})).Return(nil).Once()

		err := service.SaveMessage(message)

		assert.NoError(t, err)
		assert.NotEmpty(t, message.ID)
		assert.False(t, message.Timestamp.IsZero())

		messageRepo.AssertExpectations(t)
		publisher.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		message := &domain.Message{
			Content:    "Hello",
			UserID:     "user123",
			CustomerID: "customer456",
			Type:       domain.UserMessage,
		}

		messageRepo.On("SaveMessage", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()

		err := service.SaveMessage(message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")

		messageRepo.AssertExpectations(t)
		publisher.AssertNotCalled(t, "PublishChatMessage")
	})

	t.Run("publisher error", func(t *testing.T) {
		message := &domain.Message{
			Content:    "Hello",
			UserID:     "user123",
			CustomerID: "customer456",
			Type:       domain.UserMessage,
		}

		messageRepo.On("SaveMessage", mock.Anything, mock.Anything).Return(nil).Once()
		publisher.On("PublishChatMessage", mock.Anything).Return(errors.New("publish error")).Once()

		err := service.SaveMessage(message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "publish error")

		messageRepo.AssertExpectations(t)
		publisher.AssertExpectations(t)
	})
}

func TestGetChatHistory(t *testing.T) {
	messageRepo := new(MockMessageRepo)
	conversationRepo := new(MockConversationRepo)
	publisher := new(MockMessagePublisher)

	service := services.NewChatService(messageRepo, conversationRepo, publisher)

	t.Run("success", func(t *testing.T) {
		customerID := "customer456"
		expectedMessages := []domain.Message{
			{ID: "1", Content: "Hello", UserID: "user123", CustomerID: customerID},
			{ID: "2", Content: "Hi there", UserID: "agent789", CustomerID: customerID},
		}

		messageRepo.On("GetMessagesByCustomer", mock.Anything, customerID).Return(expectedMessages, nil).Once()

		messages, err := service.GetChatHistory(customerID)

		assert.NoError(t, err)
		assert.Equal(t, expectedMessages, messages)

		messageRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		customerID := "customer456"

		messageRepo.On("GetMessagesByCustomer", mock.Anything, customerID).Return([]domain.Message{}, errors.New("db error")).Once()

		messages, err := service.GetChatHistory(customerID)

		assert.Error(t, err)
		assert.Empty(t, messages)
		assert.Contains(t, err.Error(), "db error")

		messageRepo.AssertExpectations(t)
	})
}

func TestCreateConversation(t *testing.T) {
	messageRepo := new(MockMessageRepo)
	conversationRepo := new(MockConversationRepo)
	publisher := new(MockMessagePublisher)

	service := services.NewChatService(messageRepo, conversationRepo, publisher)

	t.Run("create new conversation", func(t *testing.T) {
		customerID := "customer456"

		// No active conversation exists
		conversationRepo.On("GetActiveConversationByCustomer", mock.Anything, customerID).
			Return(nil, errors.New("not found")).Once()

		// Expect a call to create a new one
		conversationRepo.On("CreateConversation", mock.Anything, mock.MatchedBy(func(c *domain.Conversation) bool {
			return c.CustomerID == customerID && c.Status == "active"
		})).Return(nil).Once()

		conversation, err := service.CreateConversation(customerID)

		assert.NoError(t, err)
		assert.Equal(t, customerID, conversation.CustomerID)
		assert.Equal(t, "active", conversation.Status)
		assert.NotEmpty(t, conversation.ID)
		assert.False(t, conversation.StartedAt.IsZero())

		conversationRepo.AssertExpectations(t)
	})

	t.Run("return existing conversation", func(t *testing.T) {
		customerID := "customer456"
		existingConversation := &domain.Conversation{
			ID:         "conv123",
			CustomerID: customerID,
			StartedAt:  time.Now(),
			Status:     "active",
		}

		// Return an existing active conversation
		conversationRepo.On("GetActiveConversationByCustomer", mock.Anything, customerID).
			Return(existingConversation, nil).Once()

		// CreateConversation should not be called

		conversation, err := service.CreateConversation(customerID)

		assert.NoError(t, err)
		assert.Equal(t, existingConversation, conversation)

		conversationRepo.AssertExpectations(t)
	})

	t.Run("repository error during creation", func(t *testing.T) {
		customerID := "customer456"

		// No active conversation exists
		conversationRepo.On("GetActiveConversationByCustomer", mock.Anything, customerID).
			Return(nil, errors.New("not found")).Once()

		// Creation fails
		conversationRepo.On("CreateConversation", mock.Anything, mock.Anything).
			Return(errors.New("db error")).Once()

		conversation, err := service.CreateConversation(customerID)

		assert.Error(t, err)
		assert.Nil(t, conversation)
		assert.Contains(t, err.Error(), "db error")

		conversationRepo.AssertExpectations(t)
	})
}

func TestCloseConversation(t *testing.T) {
	messageRepo := new(MockMessageRepo)
	conversationRepo := new(MockConversationRepo)
	publisher := new(MockMessagePublisher)

	service := services.NewChatService(messageRepo, conversationRepo, publisher)

	t.Run("success", func(t *testing.T) {
		conversationID := "conv123"
		conversation := &domain.Conversation{
			ID:         conversationID,
			CustomerID: "customer456",
			StartedAt:  time.Now().Add(-1 * time.Hour),
			Status:     "active",
		}

		conversationRepo.On("GetConversation", mock.Anything, conversationID).Return(conversation, nil).Once()

		conversationRepo.On("UpdateConversation", mock.Anything, mock.MatchedBy(func(c *domain.Conversation) bool {
			return c.ID == conversationID && c.Status == "closed" && !c.EndedAt.IsZero()
		})).Return(nil).Once()

		err := service.CloseConversation(conversationID)

		assert.NoError(t, err)

		conversationRepo.AssertExpectations(t)
	})

	t.Run("conversation not found", func(t *testing.T) {
		conversationID := "conv123"

		conversationRepo.On("GetConversation", mock.Anything, conversationID).
			Return(nil, errors.New("not found")).Once()

		err := service.CloseConversation(conversationID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		conversationRepo.AssertExpectations(t)
		// UpdateConversation should not be called
		conversationRepo.AssertNotCalled(t, "UpdateConversation")
	})

	t.Run("update error", func(t *testing.T) {
		conversationID := "conv123"
		conversation := &domain.Conversation{
			ID:         conversationID,
			CustomerID: "customer456",
			StartedAt:  time.Now().Add(-1 * time.Hour),
			Status:     "active",
		}

		conversationRepo.On("GetConversation", mock.Anything, conversationID).Return(conversation, nil).Once()

		conversationRepo.On("UpdateConversation", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()

		err := service.CloseConversation(conversationID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")

		conversationRepo.AssertExpectations(t)
	})
}
