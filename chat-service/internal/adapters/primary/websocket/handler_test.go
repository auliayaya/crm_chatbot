// internal/adapters/primary/websocket/handler_test.go
package websocket_test

import (
	webSock "chat-service/internal/adapters/primary/websocket"
	"chat-service/internal/core/domain"
	"chat-service/internal/core/ports"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock chat service for testing
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

// Ensure MockChatService implements ChatService interface
var _ ports.ChatService = (*MockChatService)(nil)

// Mock bot service for testing
type MockBotService struct {
	mock.Mock
}

func (m *MockChatService) SubscribeToMessages(handler func(*domain.Message)) error {
	args := m.Called(handler)
	return args.Error(0)
}

func (m *MockBotService) ProcessMessage(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockBotService) SetupAIClient(apiKey string) {
	m.Called(apiKey)
}

// Ensure MockBotService implements BotService interface
var _ ports.BotService = (*MockBotService)(nil)

func TestServeWS(t *testing.T) {
	t.Run("missing parameters", func(t *testing.T) {
		// Create mock services
		mockChatService := new(MockChatService)
		mockBotService := new(MockBotService)

		// Create a test HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hub := webSock.NewHub(mockChatService, mockBotService)
			go hub.Run()
			webSock.ServeWS(hub, w, r)
		}))
		defer server.Close()

		// Convert http:// to ws://
		wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/ws"

		// Connect without required parameters
		_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

		// Should fail with parameters missing
		assert.Error(t, err)
	})

	t.Run("conversation creation failure", func(t *testing.T) {
		// Create mock services
		mockChatService := new(MockChatService)
		mockBotService := new(MockBotService)

		// Mock conversation creation failure
		mockChatService.On("CreateConversation", "customer123").Return(nil, errors.New("creation failed")).Once()

		// Create a test HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hub := webSock.NewHub(mockChatService, mockBotService)
			go hub.Run()
			webSock.ServeWS(hub, w, r)
		}))
		defer server.Close()

		// Convert http:// to ws://
		wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/ws?user_id=user123&customer_id=customer123"

		// Connect with parameters
		_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

		// Should fail due to conversation creation error
		assert.Error(t, err)

		mockChatService.AssertExpectations(t)
	})

	t.Run("successful connection", func(t *testing.T) {
		// Create mock services
		mockChatService := new(MockChatService)
		mockBotService := new(MockBotService)

		// Mock successful conversation creation
		conversation := &domain.Conversation{
			ID:         "conv123",
			CustomerID: "customer123",
			StartedAt:  time.Now(),
			Status:     "active",
		}
		mockChatService.On("CreateConversation", "customer123").Return(conversation, nil).Once()

		// Mock empty chat history
		mockChatService.On("GetChatHistory", "customer123").Return([]domain.Message{}, nil).Once()

		// Create a test HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hub := webSock.NewHub(mockChatService, mockBotService)
			go hub.Run()
			webSock.ServeWS(hub, w, r)
		}))
		defer server.Close()

		// Convert http:// to ws://
		wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/ws?user_id=user123&customer_id=customer123"

		// Connect with parameters
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

		// Should succeed
		assert.NoError(t, err)

		if conn != nil {
			defer conn.Close()
		}

		mockChatService.AssertExpectations(t)
	})
}

func TestBotInteraction(t *testing.T) {
	// Create mock services
	mockChatService := new(MockChatService)
	mockBotService := new(MockBotService)

	// Mock successful conversation creation
	conversation := &domain.Conversation{
		ID:         "conv123",
		CustomerID: "customer123",
		StartedAt:  time.Now(),
		Status:     "active",
	}
	mockChatService.On("CreateConversation", "customer123").Return(conversation, nil)
	mockChatService.On("GetChatHistory", "customer123").Return([]domain.Message{}, nil)

	// Mock successful message saving
	mockChatService.On("SaveMessage", mock.MatchedBy(func(msg *domain.Message) bool {
		return msg.CustomerID == "customer123" && msg.Type == domain.UserMessage
	})).Return(nil)

	// Mock bot processing with captured message
	mockBotService.On("ProcessMessage", mock.Anything, mock.MatchedBy(func(msg *domain.Message) bool {
		return msg.CustomerID == "customer123" && msg.Content == "Hello bot"
	})).Return(nil).Run(func(args mock.Arguments) {
		// Create the bot response
		botResponse := &domain.Message{
			ID:         "bot-msg-123",
			Content:    "Hello human, how can I help?",
			UserID:     "bot-1",
			CustomerID: "customer123",
			Type:       domain.BotMessage,
			Timestamp:  time.Now(),
		}

		// Actually use the botResponse variable when setting up the mock
		mockChatService.On("SaveMessage", botResponse).Return(nil).Once()

		// OR if you need to use a matcher but still want to reference the response:
		mockChatService.On("SaveMessage", mock.MatchedBy(func(msg *domain.Message) bool {
			return msg.ID == botResponse.ID &&
				msg.Type == domain.BotMessage &&
				msg.UserID == "bot-1"
		})).Return(nil).Once()
	})

	// Set up subscription mocking
	var messageHandler func(*domain.Message)
	mockChatService.On("SubscribeToMessages", mock.AnythingOfType("func(*domain.Message)")).
		Return(nil).
		Run(func(args mock.Arguments) {
			// Capture the handler function to call it later
			messageHandler = args.Get(0).(func(*domain.Message))
		})

	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub := webSock.NewHub(mockChatService, mockBotService)
		if err := hub.SubscribeToBotMessages(); err != nil {
			t.Fatalf("Failed to subscribe to messages: %v", err)
		}
		go hub.Run()
		webSock.ServeWS(hub, w, r)
	}))
	defer server.Close()

	// Connect to WebSocket
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	u.Path = "/ws"
	u.RawQuery = "user_id=user123&customer_id=customer123"

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(t, err)
	defer conn.Close()

	// Send a message
	userMessage := domain.Message{
		Content:    "Hello bot",
		UserID:     "user123",
		CustomerID: "customer123",
		Type:       domain.UserMessage,
	}
	err = conn.WriteJSON(userMessage)
	assert.NoError(t, err)

	// Read and discard the echo of our own message
	var echoResponse domain.Message
	err = conn.ReadJSON(&echoResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Hello bot", echoResponse.Content) // Optional verification

	// Allow time for processing
	time.Sleep(100 * time.Millisecond)

	// Verify expectations
	mockBotService.AssertExpectations(t)

	// Simulate bot response coming through the message subscription
	if messageHandler != nil {
		botResponse := &domain.Message{
			ID:         "bot-msg-123",
			Content:    "Hello human, how can I help?",
			UserID:     "bot-1",
			CustomerID: "customer123",
			Type:       domain.BotMessage,
			Timestamp:  time.Now(),
		}
		messageHandler(botResponse)

		// Now read the bot's response (second message)
		var botResponseReceived domain.Message
		err = conn.ReadJSON(&botResponseReceived)
		assert.NoError(t, err)
		assert.Equal(t, domain.BotMessage, botResponseReceived.Type)
		assert.Equal(t, "Hello human, how can I help?", botResponseReceived.Content)
	}
}
