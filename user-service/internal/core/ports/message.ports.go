package ports

// MessagePublisher defines the interface for publishing messages
type MessagePublisher interface {
	PublishUserEvent(eventType string, payload interface{}) error
	Close()
}
