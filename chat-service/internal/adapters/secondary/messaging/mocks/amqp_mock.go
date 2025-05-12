// internal/adapters/secondary/messaging/mocks/amqp_mock.go

package mocks

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"
)

// MockConnection is a mock for amqp.Connection
type MockConnection struct {
	mock.Mock
}

func (m *MockConnection) Channel() (*amqp.Channel, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*amqp.Channel), args.Error(1)
}

func (m *MockConnection) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockChannel is a mock for amqp.Channel
type MockChannel struct {
	mock.Mock
}

func (m *MockChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	mockArgs := m.Called(name, kind, durable, autoDelete, internal, noWait, args)
	return mockArgs.Error(0)
}

func (m *MockChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	mockArgs := m.Called(exchange, key, mandatory, immediate, msg)
	return mockArgs.Error(0)
}

func (m *MockChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	mockArgs := m.Called(name, durable, autoDelete, exclusive, noWait, args)
	return mockArgs.Get(0).(amqp.Queue), mockArgs.Error(1)
}

func (m *MockChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	mockArgs := m.Called(name, key, exchange, noWait, args)
	return mockArgs.Error(0)
}

func (m *MockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	mockArgs := m.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return mockArgs.Get(0).(<-chan amqp.Delivery), mockArgs.Error(1)
}

func (m *MockChannel) Close() error {
	args := m.Called()
	return args.Error(0)
}
