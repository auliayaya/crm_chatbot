// internal/adapters/primary/websocket/hub.go
package websocket

import (
	"chat-service/internal/core/domain"
	"chat-service/internal/core/ports"
	"context"
	"encoding/json"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Inbound messages from clients
	broadcast chan []byte

	// Chat service
	chatService ports.ChatService

	// Add bot agent
	botAgent ports.BotService // New field for bot agent
}

// NewHub creates a new Hub
func NewHub(chatService ports.ChatService, botAgent ports.BotService) *Hub {
	return &Hub{
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte),
		chatService: chatService,
		botAgent:    botAgent, // Include bot agent
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

			// Always get chat history for newly connected client
			history, err := h.chatService.GetChatHistory(client.customerID)
			// Send history only if there are messages and no error
			if err == nil && len(history) > 0 {
				historyJSON, err := json.Marshal(history)
				if err == nil {
					client.send <- historyJSON
				}
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			// Process and save the message
			var msg domain.Message
			if err := json.Unmarshal(message, &msg); err == nil {
				// Save the message
				if err := h.chatService.SaveMessage(&msg); err != nil {
					log.Printf("Error saving message: %v", err)
					continue
				}

				// If it's a user message, process with bot agent
				if msg.Type == domain.UserMessage {
					log.Printf("HUB: Processing user message: %s from user: %s, customer: %s",
						msg.Content, msg.UserID, msg.CustomerID)

					// Process with bot agent synchronously for testing
					if err := h.botAgent.ProcessMessage(context.Background(), &msg); err != nil {
						log.Printf("Error processing message with bot: %v", err)
					} else {
						log.Printf("HUB: Bot successfully processed message")
					}
				}
			}

			// Broadcast original message to all clients
			// (Bot response will come through message subscription)
			for client := range h.clients {
				// Only send to clients in the same conversation
				if client.customerID == msg.CustomerID {
					if client.userID == msg.UserID && msg.Type == domain.UserMessage {
						// Skip sending to the original sender
						continue
					}
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}

// SubscribeToBotMessages sets up subscription for bot messages
func (h *Hub) SubscribeToBotMessages() error {
	return h.chatService.SubscribeToMessages(func(msg *domain.Message) {
		// When a message comes from the subscription (e.g., bot response)
		// broadcast it to clients in the same conversation
		messageJSON, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error marshaling bot message: %v", err)
			return
		}

		for client := range h.clients {
			// Only send to clients in the same conversation
			if client.customerID == msg.CustomerID {
				select {
				case client.send <- messageJSON:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	})
}

// SendBotResponse sends a bot response directly to the appropriate clients
func (h *Hub) SendBotResponse(response *domain.Message) {
	// Debug the incoming response
	log.Printf("HUB DEBUG: Received bot response to send: ID=%s, Content=%s, Customer=%s",
		response.ID, response.Content, response.CustomerID)

	messageJSON, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling direct bot response: %v", err)
		return
	}

	// Debug the JSON
	log.Printf("HUB DEBUG: Marshaled response: %s", string(messageJSON))

	clientCount := 0
	for client := range h.clients {
		// Debug client information
		log.Printf("HUB DEBUG: Checking client with customerID=%s against response customerID=%s",
			client.customerID, response.CustomerID)

		if client.customerID == response.CustomerID {
			clientCount++
			select {
			case client.send <- messageJSON:
				log.Printf("HUB DEBUG: Successfully sent to client %p", client)
			default:
				log.Printf("HUB DEBUG: Failed to send to client - buffer full")
				close(client.send)
				delete(h.clients, client)
			}
		}
	}

	if clientCount == 0 {
		log.Printf("HUB WARNING: No clients found for customerID=%s", response.CustomerID)
	} else {
		log.Printf("HUB INFO: Bot response sent to %d clients", clientCount)
	}
}
