// internal/adapters/secondary/messaging/rabbitmq_client_test.go
package messaging

import (
	"chat-service/internal/core/domain"
	"encoding/json"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Define interfaces for what we need to mock


// Mock implementations with correct parameter types
type MockAMQPConnection struct {
	mock.Mock
}

func (m *MockAMQPConnection) Channel() (*amqp.Channel, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*amqp.Channel), args.Error(1)
}

func (m *MockAMQPConnection) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockAMQPChannel struct {
	mock.Mock
}

func (m *MockAMQPChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	callArgs := m.Called(name, kind, durable, autoDelete, internal, noWait, args)
	return callArgs.Error(0)
}

func (m *MockAMQPChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	callArgs := m.Called(exchange, key, mandatory, immediate, msg)
	return callArgs.Error(0)
}

func (m *MockAMQPChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	callArgs := m.Called(name, durable, autoDelete, exclusive, noWait, args)
	if q, ok := callArgs.Get(0).(amqp.Queue); ok {
		return q, callArgs.Error(1)
	}
	return amqp.Queue{}, callArgs.Error(1)
}

func (m *MockAMQPChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	callArgs := m.Called(name, key, exchange, noWait, args)
	return callArgs.Error(0)
}

func (m *MockAMQPChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	callArgs := m.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	if ch, ok := callArgs.Get(0).(<-chan amqp.Delivery); ok {
		return ch, callArgs.Error(1)
	}
	return nil, callArgs.Error(1)
}

func (m *MockAMQPChannel) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestRabbitMQClientWithMocks tests the RabbitMQ client with mocks
func TestRabbitMQClientWithMocks(t *testing.T) {
	t.Run("PublishChatMessage", func(t *testing.T) {
		// Setup
		mockConn := new(MockAMQPConnection)
		mockChan := new(MockAMQPChannel)

		// Don't expect ExchangeDeclare in createMockRabbitMQClient
		// since we're going to test that separately

		// Create client with our custom constructor that uses interfaces
		client := &RabbitMQClient{
			conn:    mockConn,
			channel: mockChan,
		}

		// Test message
		message := &domain.Message{
			ID:         "test-id",
			Content:    "Test message",
			UserID:     "user1",
			CustomerID: "customer1",
			Type:       domain.UserMessage,
			Timestamp:  time.Now(),
		}

		// Expect Publish with the correct parameters
		mockChan.On("Publish",
			"chat_events",
			"chat.messages",
			false,
			false,
			mock.MatchedBy(func(msg amqp.Publishing) bool {
				var decodedMsg domain.Message
				err := json.Unmarshal(msg.Body, &decodedMsg)
				return err == nil &&
					msg.ContentType == "application/json" &&
					decodedMsg.ID == message.ID
			})).Return(nil)

		// Act
		err := client.PublishChatMessage(message)

		// Assert
		assert.NoError(t, err)
		mockChan.AssertExpectations(t)
	})

	t.Run("SubscribeToMessages", func(t *testing.T) {
		// Setup
		mockConn := new(MockAMQPConnection)
		mockChan := new(MockAMQPChannel)

		client := &RabbitMQClient{
			conn:    mockConn,
			channel: mockChan,
		}

		// Create a queue for the result
		queue := amqp.Queue{Name: "chat_service_queue"}

		// Create a channel for test deliveries
		deliveries := make(chan amqp.Delivery)

		// Setup expectations with correct parameter types
		mockChan.On("QueueDeclare",
			"chat_service_queue",
			true,
			false,
			false,
			false,
			amqp.Table(nil)).Return(queue, nil)

		mockChan.On("QueueBind",
			"chat_service_queue",
			"bot.responses",
			"chat_events",
			false,
			amqp.Table(nil)).Return(nil)

		mockChan.On("Consume",
			"chat_service_queue",
			"",
			true,
			false,
			false,
			false,
			amqp.Table(nil)).Return((<-chan amqp.Delivery)(deliveries), nil)

		// Message handler for testing
		var receivedMsg *domain.Message
		handler := func(msg *domain.Message) {
			receivedMsg = msg
		}

		// Act - subscribe to messages
		err := client.SubscribeToMessages(handler)
		assert.NoError(t, err)

		// Simulate a message delivery
		testMessage := &domain.Message{
			ID:        "msg1",
			Content:   "Hello from bot",
			Type:      domain.BotMessage,
			Timestamp: time.Now(),
		}

		msgBytes, _ := json.Marshal(testMessage)

		// Send test message through the channel
		go func() {
			deliveries <- amqp.Delivery{Body: msgBytes}
		}()

		// Wait a bit for processing
		time.Sleep(100 * time.Millisecond)

		// Assert the message was processed
		assert.NotNil(t, receivedMsg)
		if receivedMsg != nil {
			assert.Equal(t, testMessage.ID, receivedMsg.ID)
			assert.Equal(t, testMessage.Content, receivedMsg.Content)
		}

		mockChan.AssertExpectations(t)
	})

	t.Run("Close", func(t *testing.T) {
		// Setup
		mockConn := new(MockAMQPConnection)
		mockChan := new(MockAMQPChannel)

		client := &RabbitMQClient{
			conn:    mockConn,
			channel: mockChan,
		}

		// Expect close calls
		mockChan.On("Close").Return(nil)
		mockConn.On("Close").Return(nil)

		// Act
		client.Close()

		// Assert
		mockChan.AssertExpectations(t)
		mockConn.AssertExpectations(t)
	})
}
