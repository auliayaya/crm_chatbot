package http

import (
	"crm-service/internal/core/domain"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// GetAgents handles GET /agents
func (h *Handlers) GetAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := h.agentService.GetAgents(r.Context())
	if err != nil {
		logger.Printf("Error retrieving agents: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve agents")
		return
	}

	h.respondWithJSON(w, http.StatusOK, agents)
}

// GetAgent handles GET /agents/{id}
func (h *Handlers) GetAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	agent, err := h.agentService.GetAgentByID(r.Context(), id)
	if err != nil {
		logger.Printf("Error retrieving agent %s: %v", id, err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve agent")
		return
	}

	if agent == nil {
		h.respondWithError(w, http.StatusNotFound, "Agent not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, agent)
}

// CreateAgent handles POST /agents
func (h *Handlers) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var agent domain.Agent

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if agent.Email == "" || agent.FirstName == "" || agent.LastName == "" || agent.Department == "" {
		h.respondWithError(w, http.StatusBadRequest, "Email, first name, last name, and department are required")
		return
	}

	// Create agent
	if err := h.agentService.CreateAgent(r.Context(), &agent); err != nil {
		logger.Printf("Error creating agent: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create agent")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, agent)
}

// UpdateAgent handles PUT /agents/{id}
func (h *Handlers) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var agent domain.Agent

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Ensure ID in path matches body
	agent.ID = id

	// Update agent
	if err := h.agentService.UpdateAgent(r.Context(), &agent); err != nil {
		logger.Printf("Error updating agent %s: %v", id, err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update agent")
		return
	}

	h.respondWithJSON(w, http.StatusOK, agent)
}

// GetAgentWorkloads handles GET /agents/workloads
func (h *Handlers) GetAgentWorkloads(w http.ResponseWriter, r *http.Request) {
	workloads, err := h.agentService.GetAgentWorkloads(r.Context())
	if err != nil {
		logger.Printf("Error retrieving agent workloads: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve agent workloads")
		return
	}

	h.respondWithJSON(w, http.StatusOK, workloads)
}
