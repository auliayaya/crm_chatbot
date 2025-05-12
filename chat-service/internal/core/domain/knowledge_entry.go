package domain

import "time"

// KnowledgeEntry represents a Q&A pair
type KnowledgeEntry struct {
	ID        string    `json:"id" db:"id"`
	Question  string    `json:"question" db:"question"`
	Answer    string    `json:"answer" db:"answer"`
	Keywords  []string  `json:"keywords,omitempty" db:"keywords"`
	Category  string    `json:"category,omitempty" db:"category"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
