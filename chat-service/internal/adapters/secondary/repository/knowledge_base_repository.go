package repository

import (
	"chat-service/internal/core/domain"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

// PostgresKnowledgeRepository implements KnowledgeRepository
type PostgresKnowledgeRepository struct {
	db *sql.DB
}

// NewPostgresKnowledgeRepository creates a new PostgresKnowledgeRepository
func NewPostgresKnowledgeRepository(db *sql.DB) *PostgresKnowledgeRepository {
	return &PostgresKnowledgeRepository{db: db}
}
func (r *PostgresKnowledgeRepository) GetDB() *sql.DB {
    return r.db
}

// InitSchema creates the required tables if they don't exist
func (r *PostgresKnowledgeRepository) InitSchema(ctx context.Context) error {
    // Create knowledge entries table
    _, err := r.db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS knowledge_entries (
            id VARCHAR(100) PRIMARY KEY,
            question TEXT NOT NULL,
            answer TEXT NOT NULL,
            keywords TEXT[] NOT NULL,
            category VARCHAR(100),
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        )
    `)
    if err != nil {
        return err
    }
    
    // Create a GIN index for efficient keyword searching
    _, err = r.db.ExecContext(ctx, `
        CREATE INDEX IF NOT EXISTS idx_knowledge_entries_keywords 
        ON knowledge_entries USING GIN (keywords)
    `)
    
    return err
}
// GetAllEntries fetches all knowledge base entries
func (r *PostgresKnowledgeRepository) GetAllEntries(ctx context.Context) ([]domain.KnowledgeEntry, error) {
	query := `SELECT id, question, answer, keywords, category, created_at, updated_at 
              FROM knowledge_entries ORDER BY updated_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.KnowledgeEntry
	for rows.Next() {
		var entry domain.KnowledgeEntry
		if err := rows.Scan(
			&entry.ID,
			&entry.Question,
			&entry.Answer,
			pq.Array(&entry.Keywords),
			&entry.Category,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetEntryByID fetches a knowledge entry by ID
func (r *PostgresKnowledgeRepository) GetEntryByID(ctx context.Context, id string) (*domain.KnowledgeEntry, error) {
	query := `SELECT id, question, answer, keywords, category, created_at, updated_at 
              FROM knowledge_entries WHERE id = $1`

	var entry domain.KnowledgeEntry
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.Question,
		&entry.Answer,
		pq.Array(&entry.Keywords),
		&entry.Category,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found, return nil
		}
		return nil, err
	}

	return &entry, nil
}

// CreateEntry creates a new knowledge base entry
func (r *PostgresKnowledgeRepository) CreateEntry(ctx context.Context, entry *domain.KnowledgeEntry) error {
	query := `INSERT INTO knowledge_entries (id, question, answer, keywords, category, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`

	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	_, err := r.db.ExecContext(
		ctx,
		query,
		entry.ID,
		entry.Question,
		entry.Answer,
		pq.Array(entry.Keywords),
		entry.Category,
		entry.CreatedAt,
		entry.UpdatedAt,
	)

	return err
}

// UpdateEntry updates an existing knowledge base entry
func (r *PostgresKnowledgeRepository) UpdateEntry(ctx context.Context, entry *domain.KnowledgeEntry) error {
	query := `UPDATE knowledge_entries 
              SET question = $2, answer = $3, keywords = $4, category = $5, updated_at = $6
              WHERE id = $1`

	entry.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		entry.ID,
		entry.Question,
		entry.Answer,
		pq.Array(entry.Keywords),
		entry.Category,
		entry.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("entry not found")
	}

	return nil
}

// DeleteEntry deletes a knowledge base entry
func (r *PostgresKnowledgeRepository) DeleteEntry(ctx context.Context, id string) error {
	query := `DELETE FROM knowledge_entries WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("entry not found")
	}

	return nil
}

// SearchEntries searches for entries matching keywords
func (r *PostgresKnowledgeRepository) SearchEntries(ctx context.Context, query string) ([]domain.KnowledgeEntry, error) {
	sqlQuery := `
        SELECT id, question, answer, keywords, category, created_at, updated_at 
        FROM knowledge_entries 
        WHERE question ILIKE $1 
           OR $1 = ANY(keywords)
        ORDER BY updated_at DESC
    `

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.KnowledgeEntry
	for rows.Next() {
		var entry domain.KnowledgeEntry
		if err := rows.Scan(
			&entry.ID,
			&entry.Question,
			&entry.Answer,
			pq.Array(&entry.Keywords),
			&entry.Category,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
