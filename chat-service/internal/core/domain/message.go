// internal/core/domain/message.go
package domain

import "time"

type MessageType string

const (
	UserMessage   MessageType = "user"
	SystemMessage MessageType = "system"
	BotMessage    MessageType = "bot"
)

type Message struct {
	ID         string      `json:"id"`
	Content    string      `json:"content"`
	UserID     string      `json:"user_id"`
	CustomerID string      `json:"customer_id"`
	Type       MessageType `json:"type"`
	Timestamp  time.Time   `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type Conversation struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at,omitempty"`
	Status     string    `json:"status"` // "active", "closed"
}
