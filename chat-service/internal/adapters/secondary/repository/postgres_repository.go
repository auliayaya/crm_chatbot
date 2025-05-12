// internal/adapters/secondary/repository/postgres_repository.go
package repository

import (
	"chat-service/internal/core/domain"
	"chat-service/internal/core/ports"
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(host, user, password, dbname string, connStringOpt ...string) (*PostgresRepository, error) {
	var connStr string

	// If a direct connection string is provided, use it
	if len(connStringOpt) > 0 && connStringOpt[0] != "" {
		connStr = connStringOpt[0]
	} else {
		// Otherwise build from individual parameters
		connStr = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
			host, user, password, dbname)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Initialize the tables
	err = initTables(db)
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}
func (r *PostgresRepository) GetDB() *sql.DB {
    return r.db
}

func initTables(db *sql.DB) error {
	// Create conversations table
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS conversations (
            id VARCHAR(36) PRIMARY KEY,
            customer_id VARCHAR(36) NOT NULL,
            started_at TIMESTAMP NOT NULL,
            ended_at TIMESTAMP,
            status VARCHAR(20) NOT NULL
        )
    `)
	if err != nil {
		return err
	}

	// Create messages table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS messages (
            id VARCHAR(36) PRIMARY KEY,
            content TEXT NOT NULL,
            user_id VARCHAR(36) NOT NULL,
            customer_id VARCHAR(36) NOT NULL,
            conversation_id VARCHAR(36) REFERENCES conversations(id),
            type VARCHAR(20) NOT NULL,
            timestamp TIMESTAMP NOT NULL
        )
    `)
	return err
}

// Update SaveMessage to include context
func (r *PostgresRepository) SaveMessage(ctx context.Context, message *domain.Message) error {
	// Get the active conversation for this customer
	conversation, err := r.GetActiveConversationByCustomer(ctx, message.CustomerID)
	if err != nil {
		return err
	}

	// Insert the message - use ExecContext to pass the context
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO messages (id, content, user_id, customer_id, conversation_id, type, timestamp)
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		message.ID, message.Content, message.UserID, message.CustomerID,
		conversation.ID, message.Type, message.Timestamp,
	)
	return err
}

func (r *PostgresRepository) GetMessagesByCustomer(ctx context.Context, customerID string) ([]domain.Message, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, content, user_id, customer_id, type, timestamp
         FROM messages
         WHERE customer_id = $1
         ORDER BY timestamp ASC`,
		customerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(&msg.ID, &msg.Content, &msg.UserID, &msg.CustomerID, &msg.Type, &msg.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *PostgresRepository) GetMessagesByConversation(ctx context.Context, conversationID string) ([]domain.Message, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, content, user_id, customer_id, type, timestamp 
         FROM messages 
         WHERE conversation_id = $1
         ORDER BY timestamp ASC`,
		conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(&msg.ID, &msg.Content, &msg.UserID, &msg.CustomerID, &msg.Type, &msg.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// Implement ConversationRepository interface
func (r *PostgresRepository) CreateConversation(ctx context.Context, conversation *domain.Conversation) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO conversations (id, customer_id, started_at, status)
         VALUES ($1, $2, $3, $4)`,
		conversation.ID, conversation.CustomerID, conversation.StartedAt, conversation.Status,
	)
	return err
}

func (r *PostgresRepository) GetConversation(ctx context.Context, id string) (*domain.Conversation, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, customer_id, started_at, ended_at, status 
         FROM conversations 
         WHERE id = $1`,
		id,
	)

	var conversation domain.Conversation
	var endedAt sql.NullTime

	err := row.Scan(
		&conversation.ID,
		&conversation.CustomerID,
		&conversation.StartedAt,
		&endedAt,
		&conversation.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("conversation not found")
		}
		return nil, err
	}

	if endedAt.Valid {
		conversation.EndedAt = endedAt.Time
	}

	return &conversation, nil
}

func (r *PostgresRepository) GetActiveConversationByCustomer(ctx context.Context, customerID string) (*domain.Conversation, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, customer_id, started_at, ended_at, status 
         FROM conversations 
         WHERE customer_id = $1 AND status = 'active'
         ORDER BY started_at DESC
         LIMIT 1`,
		customerID,
	)

	var conversation domain.Conversation
	var endedAt sql.NullTime

	err := row.Scan(
		&conversation.ID,
		&conversation.CustomerID,
		&conversation.StartedAt,
		&endedAt,
		&conversation.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no active conversation found")
		}
		return nil, err
	}

	if endedAt.Valid {
		conversation.EndedAt = endedAt.Time
	}

	return &conversation, nil
}

func (r *PostgresRepository) UpdateConversation(ctx context.Context, conversation *domain.Conversation) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations 
         SET status = $1, ended_at = $2
         WHERE id = $3`,
		conversation.Status,
		conversation.EndedAt,
		conversation.ID,
	)
	return err
}

// Implement both interfaces with a single struct
var _ ports.MessageRepository = (*PostgresRepository)(nil)
var _ ports.ConversationRepository = (*PostgresRepository)(nil)
