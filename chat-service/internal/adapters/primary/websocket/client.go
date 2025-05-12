// internal/adapters/primary/websocket/client.go
package websocket

import (
	"bytes"
	"chat-service/internal/core/domain"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from client.
	maxMessageSize = 10240
)

// Client represents a connected WebSocket client
type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	send           chan []byte
	userID         string
	customerID     string
	conversationID string
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg domain.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error decoding message: %v", err)
			continue
		}

		// Add user and customer IDs from the connection
		msg.UserID = c.userID
		msg.CustomerID = c.customerID
		msg.Metadata = map[string]string{
			"conversation_id": c.conversationID,
			"originalSender":        c.userID,
			"clientID":     c.conn.RemoteAddr().String(),
		}

		// Process the message in the hub
		jsonMsg, _ := json.Marshal(msg)
		c.hub.broadcast <- jsonMsg
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			// Add debug logging
			log.Printf("CLIENT DEBUG: Received message to write, size=%d", len(message))

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Debug successful write
			log.Printf("CLIENT DEBUG: Successfully wrote message to websocket")

			// Add queued messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(bytes.TrimSpace([]byte{'\n'}))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
