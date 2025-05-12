// internal/core/ports/secondary.go
package ports

import (
	"chat-service/internal/core/domain"
	"context"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, message *domain.Message) error
	GetMessagesByCustomer(ctx context.Context,customerID string) ([]domain.Message, error)
	GetMessagesByConversation(ctx context.Context, conversationID string) ([]domain.Message, error)
}

type ConversationRepository interface {
	CreateConversation(ctx context.Context,conversation *domain.Conversation) error
	GetConversation(ctx context.Context,id string) (*domain.Conversation, error)
	GetActiveConversationByCustomer(ctx context.Context,customerID string) (*domain.Conversation, error)
	UpdateConversation(ctx context.Context,conversation *domain.Conversation) error
}

type MessagePublisher interface {
	PublishChatMessage(message *domain.Message) error
	SubscribeToMessages(handler func(*domain.Message)) error
	Close()
}

type KnowledgeRepository interface {
    GetAllEntries(ctx context.Context) ([]domain.KnowledgeEntry, error)
    GetEntryByID(ctx context.Context, id string) (*domain.KnowledgeEntry, error)
    CreateEntry(ctx context.Context, entry *domain.KnowledgeEntry) error
    UpdateEntry(ctx context.Context, entry *domain.KnowledgeEntry) error
    DeleteEntry(ctx context.Context, id string) error
    SearchEntries(ctx context.Context, query string) ([]domain.KnowledgeEntry, error)
}