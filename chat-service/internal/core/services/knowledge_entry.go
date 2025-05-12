package services

import (
	"chat-service/internal/core/domain"
	"chat-service/internal/core/ports"
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

// KnowledgeBase manages Q&A entries for the chatbot
type KnowledgeBase struct {
	repository    ports.KnowledgeRepository
	cachedEntries []domain.KnowledgeEntry
	mutex         sync.RWMutex
	lastUpdate    time.Time
}

// NewKnowledgeBase creates a new knowledge base
func NewKnowledgeBase(repository ports.KnowledgeRepository) *KnowledgeBase {
	kb := &KnowledgeBase{
		repository: repository,
		lastUpdate: time.Time{}, // Zero time
	}

	// Add initial default entries if repo is empty
	go kb.initializeDefaultEntries()

	return kb
}

// initializeDefaultEntries adds default entries if none exist
func (kb *KnowledgeBase) initializeDefaultEntries() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	entries, err := kb.repository.GetAllEntries(ctx)
	if err != nil {
		log.Printf("Error checking existing knowledge entries: %v", err)
		return
	}

	if len(entries) == 0 {
		log.Println("No existing knowledge entries found, adding defaults")
		defaultEntries := []domain.KnowledgeEntry{
			{
				ID:       "greeting",
				Question: "hello",
				Answer:   "Hello! How can I assist you today?",
				Keywords: []string{"hi", "hello", "hey", "greetings"},
				Category: "general",
			},
			{
				ID:       "help",
				Question: "help",
				Answer:   "I'm here to help! What do you need assistance with?",
				Keywords: []string{"help", "assist", "support"},
				Category: "general",
			},
			{
				ID:       "order_status",
				Question: "order status",
				Answer:   "To check your order status, please provide your order number.",
				Keywords: []string{"order", "status", "track"},
				Category: "orders",
			},
			{
				ID:       "payment_methods",
				Question: "payment methods",
				Answer:   "We accept credit cards, PayPal, and bank transfers.",
				Keywords: []string{"payment", "pay", "credit card", "paypal"},
				Category: "payments",
			},
		}

		for _, entry := range defaultEntries {
			if err := kb.repository.CreateEntry(ctx, &entry); err != nil {
				log.Printf("Error adding default entry %s: %v", entry.ID, err)
			}
		}

		log.Println("Default knowledge entries added successfully")
	}

	// Initialize the cache
	kb.RefreshCache()
}

// refreshCache updates the internal cache from the repository
func (kb *KnowledgeBase) RefreshCache() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	log.Printf("Refreshing knowledge base cache...")

	entries, err := kb.repository.GetAllEntries(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to refresh knowledge base cache: %v", err)
		// Don't update lastUpdate time, so we'll try again soon
		return
	}

	kb.mutex.Lock()
	defer kb.mutex.Unlock()

	if len(entries) > 0 {
		kb.cachedEntries = entries
		kb.lastUpdate = time.Now()
		log.Printf("Knowledge base cache refreshed with %d entries", len(entries))
	} else {
		log.Printf("WARNING: Knowledge base query returned 0 entries")
		// Keep previous cache if we got zero entries (might be a temporary issue)
		// But do update the timestamp to prevent constant retries
		if len(kb.cachedEntries) > 0 {
			kb.lastUpdate = time.Now()
		}
	}
}

// FindBestMatch finds the best response with database error handling
func (kb *KnowledgeBase) FindBestMatch(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))

	// Check if we need to refresh cache
	if time.Since(kb.lastUpdate) > 5*time.Minute || len(kb.cachedEntries) == 0 {
		kb.RefreshCache()
	}

	kb.mutex.RLock()
	defer kb.mutex.RUnlock()

	// If cache is empty after refresh attempt, we might have database issues
	if len(kb.cachedEntries) == 0 {
		log.Printf("WARNING: Knowledge base cache is empty, possible database connectivity issue")
		return "I'm having trouble accessing my knowledge right now. Could you try again in a moment?"
	}

	// Try exact matches first
	for _, entry := range kb.cachedEntries {
		if strings.Contains(input, strings.ToLower(entry.Question)) {
			log.Printf("KB: Found exact match with entry: %s", entry.ID)
			return entry.Answer
		}
	}

	// Try keyword matches with scoring
	bestScore := 0
	var bestAnswer string

	for _, entry := range kb.cachedEntries {
		for _, keyword := range entry.Keywords {
			keyword = strings.ToLower(keyword)
			if strings.Contains(input, keyword) {
				// Score based on keyword length (longer keywords are more specific)
				score := len(keyword)
				if score > bestScore {
					bestScore = score
					bestAnswer = entry.Answer
				}
			}
		}
	}

	if bestScore > 0 {
		return bestAnswer
	}

	// No match found
	return "I'm not sure how to respond to that. Could you try phrasing your question differently?"
}

// LogEntries logs all entries in the knowledge base
func (kb *KnowledgeBase) LogEntries() {
	kb.mutex.RLock()
	defer kb.mutex.RUnlock()

	log.Printf("Knowledge Base has %d cached entries:", len(kb.cachedEntries))
	for _, entry := range kb.cachedEntries {
		keywords := strings.Join(entry.Keywords, ", ")
		log.Printf("- ID: %s, Question: '%s', Keywords: [%s]",
			entry.ID, entry.Question, keywords)
	}
}

// GetAllEntries returns all entries
func (kb *KnowledgeBase) GetAllEntries() []domain.KnowledgeEntry {
	kb.mutex.RLock()
	defer kb.mutex.RUnlock()

	result := make([]domain.KnowledgeEntry, len(kb.cachedEntries))
	copy(result, kb.cachedEntries)
	return result
}
