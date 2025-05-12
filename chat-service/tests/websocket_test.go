// tests/websocket_test.go
package tests

import (
	webSock "chat-service/internal/adapters/primary/websocket"
	"chat-service/internal/core/domain"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBotService struct {
	mock.Mock
}

func (m *MockBotService) ProcessMessage(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockBotService) SetupAIClient(apiKey string) {
	m.Called(apiKey)
}

// MockChatService mocks the chat service
type MockChatService struct {
	mock.Mock
}

func (m *MockChatService) SaveMessage(message *domain.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockChatService) GetChatHistory(customerID string) ([]domain.Message, error) {
	args := m.Called(customerID)
	return args.Get(0).([]domain.Message), args.Error(1)
}

func (m *MockChatService) GetConversation(conversationID string) (*domain.Conversation, error) {
	args := m.Called(conversationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockChatService) CreateConversation(customerID string) (*domain.Conversation, error) {
	args := m.Called(customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockChatService) CloseConversation(conversationID string) error {
	args := m.Called(conversationID)
	return args.Error(0)
}

// SubscribeToMessages mocks the subscription to message events
func (m *MockChatService) SubscribeToMessages(handler func(*domain.Message)) error {
	args := m.Called(handler)
	return args.Error(0)
}

func TestWebSocketEndToEnd(t *testing.T) {
	// Create mock services
	mockService := new(MockChatService)
	mockBotService := new(MockBotService)

	// Create a conversation for testing
	conversation := &domain.Conversation{
		ID:         "test-conv",
		CustomerID: "customer123",
		StartedAt:  time.Now(),
		Status:     "active",
	}

	// Set up the history to return
	history := []domain.Message{
		{
			ID:         "msg1",
			Content:    "Previous message",
			UserID:     "user123",
			CustomerID: "customer123",
			Type:       domain.UserMessage,
			Timestamp:  time.Now().Add(-1 * time.Minute),
		},
	}

	// Set up mock expectations
	mockService.On("CreateConversation", "customer123").Return(conversation, nil)
	mockService.On("GetChatHistory", "customer123").Return(history, nil)
	mockService.On("SaveMessage", mock.AnythingOfType("*domain.Message")).Return(nil)

	// Add this expectation for the bot service
	mockBotService.On("ProcessMessage", mock.Anything, mock.MatchedBy(func(msg *domain.Message) bool {
		return msg.Content == "Hello from test" &&
			msg.UserID == "user123" &&
			msg.CustomerID == "customer123"
	})).Return(nil)

	// Set up subscription expectations
	mockService.On("SubscribeToMessages", mock.AnythingOfType("func(*domain.Message)")).Return(nil)

	// Create a WebSocket hub
	hub := webSock.NewHub(mockService, mockBotService)

	// Subscribe to bot messages
	err := hub.SubscribeToBotMessages()
	assert.NoError(t, err)

	go hub.Run()

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			webSock.ServeWS(hub, w, r)
		}
	}))
	defer server.Close()

	// Convert to WebSocket URL
	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/ws?user_id=user123&customer_id=customer123"

	// Connect to WebSocket
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Error connecting to WebSocket: %v", err)
	}
	defer c.Close()

	// Create channels for communication
	receivedMsg := make(chan []byte, 10)
	errorChan := make(chan error, 1)

	// Goroutine to read messages
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				errorChan <- err
				return
			}
			receivedMsg <- message
		}
	}()

	// First message should be the history
	select {
	case msg := <-receivedMsg:
		var messages []domain.Message
		err := json.Unmarshal(msg, &messages)
		assert.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Equal(t, "Previous message", messages[0].Content)
	case err := <-errorChan:
		t.Fatalf("Error reading message: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for history message")
	}

	// Send a new message
	newMessage := map[string]interface{}{
		"content": "Hello from test",
		"type":    string(domain.UserMessage),
	}
	messageBytes, _ := json.Marshal(newMessage)
	err = c.WriteMessage(websocket.TextMessage, messageBytes)
	assert.NoError(t, err)

	// We should receive the same message back (broadcast)
	select {
	case msg := <-receivedMsg:
		var receivedMessage map[string]interface{}
		err := json.Unmarshal(msg, &receivedMessage)
		assert.NoError(t, err)
		assert.Equal(t, "Hello from test", receivedMessage["content"])
		assert.Equal(t, "user123", receivedMessage["user_id"])
		assert.Equal(t, "customer123", receivedMessage["customer_id"])
	case err := <-errorChan:
		t.Fatalf("Error reading message: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for echo message")
	}

	// Verify our mock expectations
	mockService.AssertExpectations(t)
}
