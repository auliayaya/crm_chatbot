package http

import (
	"crm-service/internal/core/domain"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GetTickets handles GET /tickets
func (h *Handlers) GetTickets(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 50 // Default limit
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0 // Default offset
	}

	// Calculate current page
	page := 1
	if limit > 0 {
		page = (offset / limit) + 1
	}

	// Get tickets from service
	tickets, err := h.ticketService.GetTickets(r.Context(), limit, offset)
	if err != nil {
		logger.Printf("Error retrieving tickets: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tickets")
		return
	}

	// Get total count
	total, err := h.ticketService.GetTicketsCount(r.Context())
	if err != nil {
		logger.Printf("Error retrieving ticket count: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve ticket count")
		return
	}

	// Create response structure
	response := map[string]interface{}{
		"tickets":  tickets,
		"total":    total,
		"page":     page,
		"pageSize": limit,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetTicket handles GET /tickets/{id}
func (h *Handlers) GetTicket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	ticket, err := h.ticketService.GetTicketByID(r.Context(), id)
	if err != nil {
		logger.Printf("Error retrieving ticket %s: %v", id, err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve ticket")
		return
	}

	if ticket == nil {
		h.respondWithError(w, http.StatusNotFound, "Ticket not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, ticket)
}

// CreateTicket handles POST /tickets
func (h *Handlers) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var ticket domain.Ticket

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if ticket.CustomerID == "" || ticket.Subject == "" || ticket.Description == "" {
		h.respondWithError(w, http.StatusBadRequest, "Customer ID, subject, and description are required")
		return
	}

	// Create ticket
	if err := h.ticketService.CreateTicket(r.Context(), &ticket); err != nil {
		logger.Printf("Error creating ticket: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create ticket")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, ticket)
}

// UpdateTicket handles PUT /tickets/{id}
func (h *Handlers) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var ticket domain.Ticket

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Ensure ID in path matches body
	ticket.ID = id

	// Update ticket
	if err := h.ticketService.UpdateTicket(r.Context(), &ticket); err != nil {
		logger.Printf("Error updating ticket %s: %v", id, err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update ticket")
		return
	}

	h.respondWithJSON(w, http.StatusOK, ticket)
}

// AssignTicket handles POST /tickets/{id}/assign
func (h *Handlers) AssignTicket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Extract agent ID from request body
	var request struct {
		AgentID string `json:"agent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if request.AgentID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Agent ID is required")
		return
	}

	// Assign ticket
	if err := h.ticketService.AssignTicketToAgent(r.Context(), id, request.AgentID); err != nil {
		logger.Printf("Error assigning ticket %s: %v", id, err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to assign ticket")
		return
	}

	// Get updated ticket to return
	ticket, err := h.ticketService.GetTicketByID(r.Context(), id)
	if err != nil || ticket == nil {
		h.respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
		return
	}

	h.respondWithJSON(w, http.StatusOK, ticket)
}

// AddTicketComment handles POST /tickets/{id}/comments
func (h *Handlers) AddTicketComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var request struct {
		UserID  string `json:"user_id"`
		Content string `json:"content"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if request.UserID == "" || request.Content == "" {
		h.respondWithError(w, http.StatusBadRequest, "User ID and content are required")
		return
	}

	// Add comment
	if err := h.ticketService.AddTicketComment(r.Context(), id, request.UserID, request.Content); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "Ticket not found")
		} else {
			logger.Printf("Error adding comment to ticket %s: %v", id, err)
			h.respondWithError(w, http.StatusInternalServerError, "Failed to add comment")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// CloseTicket handles POST /tickets/{id}/close
func (h *Handlers) CloseTicket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var request struct {
		Resolution string `json:"resolution"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if request.Resolution == "" {
		h.respondWithError(w, http.StatusBadRequest, "Resolution is required")
		return
	}

	// Close ticket
	if err := h.ticketService.CloseTicket(r.Context(), id, request.Resolution); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "Ticket not found")
		} else {
			logger.Printf("Error closing ticket %s: %v", id, err)
			h.respondWithError(w, http.StatusInternalServerError, "Failed to close ticket")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// GetTicketHistory handles GET /tickets/{id}/history
func (h *Handlers) GetTicketHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	events, err := h.ticketService.GetTicketHistory(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "Ticket not found")
		} else {
			logger.Printf("Error retrieving ticket history for %s: %v", id, err)
			h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve ticket history")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, events)
}

// GetCustomerTickets handles GET /customers/{id}/tickets
func (h *Handlers) GetCustomerTickets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID := vars["id"]

	// Verify customer exists
	customer, err := h.customerService.GetCustomerByID(r.Context(), customerID)
	if err != nil {
		logger.Printf("Error retrieving customer %s: %v", customerID, err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve customer")
		return
	}

	if customer == nil {
		h.respondWithError(w, http.StatusNotFound, "Customer not found")
		return
	}

	// Get customer tickets
	tickets, err := h.ticketService.GetTicketsByCustomer(r.Context(), customerID)
	if err != nil {
		logger.Printf("Error retrieving tickets for customer %s: %v", customerID, err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tickets")
		return
	}

	h.respondWithJSON(w, http.StatusOK, tickets)
}
