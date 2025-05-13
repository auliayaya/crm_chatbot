// internal/adapters/primary/websocket/handler.go
package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWS handles WebSocket requests from clients
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Add detailed logging
	log.Printf("WS CONNECT REQUEST: Method=%s Path=%s Query=%s Headers=%v",
		r.Method, r.URL.Path, r.URL.RawQuery, r.Header)

	// Extract user ID and customer ID from request
	userID := r.URL.Query().Get("user_id")
	customerID := r.URL.Query().Get("customer_id")
	log.Printf("WS CONNECT: user_id=%s, customer_id=%s", userID, customerID)

	if userID == "" || customerID == "" {
		log.Println("Missing user_id or customer_id parameter")
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Create or get existing conversation
	conversation, err := hub.chatService.CreateConversation(customerID)
	if err != nil {
		log.Printf("Error creating conversation: %v", err)
		http.Error(w, "Error creating conversation", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("WS REGISTER: New client registered userID=%s, customerID=%s, addr=%s",
		userID, customerID, conn.RemoteAddr().String())
	client := &Client{
		hub:            hub,
		conn:           conn,
		send:           make(chan []byte, 256),
		userID:         userID,
		customerID:     customerID,
		conversationID: conversation.ID,
	}

	client.hub.register <- client

	// Start goroutines to handle messages
	go client.writePump()
	go client.readPump()
}
