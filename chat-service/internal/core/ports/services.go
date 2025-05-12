// internal/core/ports/services.go
package ports

import (
	"chat-service/internal/core/domain"
	"context"
)

// BotService defines the methods for a bot agent
type BotService interface {
	ProcessMessage(ctx context.Context, message *domain.Message) error
}

type MessageHub interface {
    SendBotResponse(message *domain.Message)
}