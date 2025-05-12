// internal/core/services/bot_agent.go
package services

import (
	"chat-service/internal/core/domain"
	"chat-service/internal/core/ports"
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/openai/openai-go"        // New import
	"github.com/openai/openai-go/option" // New import
	"github.com/openai/openai-go/packages/param"
	"golang.org/x/time/rate"
)

type BotAgent struct {
	ID            string
	Name          string
	repository    ports.MessageRepository
	publisher     ports.MessagePublisher
	useAI         bool
	conversations map[string][]openai.ChatCompletionMessageParamUnion
	mutex         sync.Mutex
	aiClient      openai.Client
	rateLimiter   *rate.Limiter

	// Response caching
	responseCache map[string]string
	cacheMutex    sync.RWMutex

	// Add a Hub field to BotAgent
	hub ports.MessageHub // Add this field

	// Add this field to the BotAgent struct
	knowledgeBase *KnowledgeBase
}

var _ ports.BotService = (*BotAgent)(nil)

func NewBotAgent(id, name string, useAi bool, repo ports.MessageRepository, pub ports.MessagePublisher, knowledgeBase *KnowledgeBase) *BotAgent {
	return &BotAgent{
		ID:            id,
		Name:          name,
		repository:    repo,
		publisher:     pub,
		useAI:         useAi,
		conversations: make(map[string][]openai.ChatCompletionMessageParamUnion),
		rateLimiter:   rate.NewLimiter(rate.Every(6*time.Second), 1), // Updated to 6 seconds
		responseCache: make(map[string]string),
		knowledgeBase: knowledgeBase,
	}
}

// Add this method to initialize the hub
func (b *BotAgent) SetHub(hub ports.MessageHub) {
	b.hub = hub
}

func (b *BotAgent) ProcessMessage(ctx context.Context, message *domain.Message) error {
	log.Printf("Bot received message: %s from user: %s (type: %s)",
		message.Content, message.UserID, message.Type)

	// Only process user messages
	if message.Type != domain.UserMessage {
		log.Printf("Bot ignoring non-user message type: %s", message.Type)
		return nil
	}

	// Create bot response with robust error handling
	var responseText string

	// Try to generate response with timeout
	responseCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		defer func() {
			done <- true
		}()

		// First check cache
		cacheKey := message.CustomerID + ":" + message.Content
		b.cacheMutex.RLock()
		cachedResp, found := b.responseCache[cacheKey]
		b.cacheMutex.RUnlock()

		if found {
			responseText = cachedResp
			log.Printf("Using cached response for '%s'", message.Content)
		} else {
			// Generate new response
			responseText = b.generateResponse(message.Content)

			// Cache the response
			b.cacheMutex.Lock()
			b.responseCache[cacheKey] = responseText
			b.cacheMutex.Unlock()
		}
	}()

	// Wait for response generation or timeout
	select {
	case <-done:
		// Response generated successfully
	case <-responseCtx.Done():
		log.Printf("Response generation timed out for: '%s'", message.Content)
		responseText = "I'm sorry, it's taking me longer than expected to respond. Please try asking again."
	}

	// Create bot response
	response := &domain.Message{
		ID:         uuid.New().String(),
		Content:    responseText,
		UserID:     b.ID, // Bot's ID
		CustomerID: message.CustomerID,
		Type:       domain.BotMessage,
		Timestamp:  time.Now(),
	}

	// Store message with error handling
	if err := b.repository.SaveMessage(ctx, response); err != nil {
		log.Printf("Error saving bot response: %v", err)
		// Continue anyway - don't fail the whole process if DB write fails
	}

	// Try both delivery methods
	// 1. Publish to message broker
	if err := b.publisher.PublishChatMessage(response); err != nil {
		log.Printf("Error publishing bot response: %v", err)
		// Try direct delivery if publishing fails
	}

	// 2. Direct delivery via hub (more reliable)
	if b.hub != nil {
		b.hub.SendBotResponse(response)
		log.Printf("Bot response sent via hub")
	}

	return nil
}

// Legacy rule-based response generator
func (b *BotAgent) generateRuleBasedResponse(input string) string {
	return b.knowledgeBase.FindBestMatch(input)
}

// SetupAIClient configures the OpenAI client
func (b *BotAgent) SetupAIClient(apiKey string) {
	if apiKey == "" {
		log.Printf("Warning: Empty OpenAI API key provided, AI features will be disabled")
		b.useAI = false
		return
	}

	b.aiClient = openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	b.useAI = true
	log.Printf("AI capabilities enabled for chat bot")
}

// AI-powered response generation with conversation history
func (b *BotAgent) generateAIResponse(ctx context.Context, message *domain.Message) string {
	// Wait for rate limiter
	if err := b.rateLimiter.Wait(ctx); err != nil {
		log.Printf("Rate limit wait canceled: %v", err)
		return b.generateRuleBasedResponse(message.Content)
	}

	b.mutex.Lock()
	if _, exists := b.conversations[message.CustomerID]; !exists {
		// Initialize with system prompt
		b.conversations[message.CustomerID] = []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful customer support assistant. Be concise and professional."),
		}
		log.Printf("Creating new conversation for customer: %s", message.CustomerID)
	}

	// Add user's message to history
	b.conversations[message.CustomerID] = append(b.conversations[message.CustomerID],
		openai.UserMessage(message.Content),
	)

	// Cap history length to prevent token overflow (keep last 10 exchanges)
	if len(b.conversations[message.CustomerID]) > 21 {
		b.conversations[message.CustomerID] = append(
			b.conversations[message.CustomerID][:1],
			b.conversations[message.CustomerID][len(b.conversations[message.CustomerID])-20:]...,
		)
	}

	// Get conversation for this customer
	conversation := b.conversations[message.CustomerID]
	b.mutex.Unlock()

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Retry logic for API calls
	var responseContent string

	backoff := 1 * time.Second
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		// Wait for rate limiter before making request
		if err := b.rateLimiter.Wait(ctx); err != nil {
			log.Printf("Rate limit wait canceled: %v", err)
			return b.generateRuleBasedResponse(message.Content)
		}

		// Create chat completion with new client
		chatCompletion, err := b.aiClient.Chat.Completions.New(
			timeoutCtx,
			openai.ChatCompletionNewParams{
				Model:       openai.ChatModelGPT3_5Turbo,    // Remove openai.F()
				Messages:    conversation,                   // Remove openai.F()
				MaxTokens:   param.Opt[int64]{Value: 100},   // Remove openai.F()
				Temperature: param.Opt[float64]{Value: 0.7}, // Properly wrap in param.Opt
			},
		)

		if err == nil && len(chatCompletion.Choices) > 0 {
			responseContent = chatCompletion.Choices[0].Message.Content
			break // Success, exit retry loop
		}

		if err != nil {
			if strings.Contains(err.Error(), "429") {
				log.Printf("Rate limited by OpenAI (attempt %d/%d), retrying in %v", i+1, maxRetries, backoff)
				time.Sleep(backoff)
				backoff *= 2 // Exponential backoff
			} else {
				log.Printf("AI service error: %v", err)
				return b.generateRuleBasedResponse(message.Content)
			}
		} else {
			log.Printf("API returned empty choices (attempt %d/%d)", i+1, maxRetries)
			err = errors.New("empty response from API")
		}
	}

	// After retry loop, check if we got a valid response
	if responseContent == "" {
		log.Printf("Failed to get response after %d retries", maxRetries)
		return b.generateRuleBasedResponse(message.Content) + " (AI service unavailable)"
	}

	// Add AI response to conversation history
	b.mutex.Lock()
	b.conversations[message.CustomerID] = append(b.conversations[message.CustomerID],
		openai.AssistantMessage(responseContent),
	)
	b.mutex.Unlock()

	return responseContent
}

// generateResponse creates a response using knowledge base first, then AI if needed
func (b *BotAgent) generateResponse(input string) string {
	input = strings.TrimSpace(input)
	normalizedInput := strings.ToLower(input)

	log.Printf("Bot generating response for: '%s'", input)

	// Step 1: Try knowledge base first with timeout and error handling
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var knowledgeResponse string
	var knowledgeErr error

	// Use a separate goroutine to query knowledge base with timeout
	done := make(chan bool, 1)
	go func() {
		defer func() {
			done <- true
		}()

		// Check if knowledge base is available
		if b.knowledgeBase != nil {
			knowledgeResponse = b.knowledgeBase.FindBestMatch(input)
			if knowledgeResponse != "" && !strings.Contains(knowledgeResponse, "I'm not sure") {
				log.Printf("Bot found knowledge base match for: '%s'", input)
			}
		} else {
			knowledgeErr = errors.New("knowledge base not available")
		}
	}()

	// Wait for knowledge base query or timeout
	select {
	case <-done:
		// Query completed
	case <-ctx.Done():
		log.Printf("Knowledge base query timed out for: '%s'", input)
		knowledgeErr = ctx.Err()
	}

	// If we got a good response from knowledge base, return it
	if knowledgeErr == nil && knowledgeResponse != "" &&
		!strings.Contains(knowledgeResponse, "I'm not sure") {
		return knowledgeResponse
	}

	// Step 2: Fall back to rule-based for common patterns
	if strings.Contains(normalizedInput, "hello") ||
		strings.Contains(normalizedInput, "hi") ||
		strings.Contains(normalizedInput, "help") {
		return b.generateRuleBasedResponse(input)
	}

	// Step 3: Fall back to AI if enabled and input is complex
	if b.useAI {
		return b.generateAIResponse(context.Background(), &domain.Message{Content: input})
	}

	// Step 4: Last resort - use basic rule-based
	return b.generateRuleBasedResponse(input)
}
