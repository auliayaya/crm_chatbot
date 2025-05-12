package http

import (
	"crm-service/internal/core/ports"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// Response represents a generic API response
type Response struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Handlers holds all API handlers
type Handlers struct {
	customerService ports.CustomerService
	ticketService   ports.TicketService
	agentService    ports.AgentService
}

// Logger for HTTP handlers
var logger = log.New(os.Stdout, "[HTTP] ", log.LstdFlags)

// NewHandlers creates new handler instances
func NewHandlers(
	customerService ports.CustomerService,
	ticketService ports.TicketService,
	agentService ports.AgentService,
) *Handlers {
	return &Handlers{
		customerService: customerService,
		ticketService:   ticketService,
		agentService:    agentService,
	}
}

// HealthCheck handles GET /health
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	h.respondWithJSON(w, http.StatusOK, status)
}

// Helper methods for consistent response handling

// respondWithError sends an error response with a specific status code
func (h *Handlers) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, Response{
		Success: false,
		Error:   message,
	})
}

// respondWithJSON sends a JSON response with a specific status code
func (h *Handlers) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		logger.Printf("Error marshalling JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success":false,"error":"Internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
