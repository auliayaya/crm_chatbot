package http

import (
	"chat-service/internal/core/domain"
	"chat-service/internal/core/ports"
	"chat-service/internal/core/services"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// AdminHandlers handles admin API requests
type AdminHandlers struct {
	knowledgeRepo ports.KnowledgeRepository
	knowledgeBase *services.KnowledgeBase
}

// NewAdminHandlers creates a new AdminHandlers
func NewAdminHandlers(repo ports.KnowledgeRepository, kb *services.KnowledgeBase) *AdminHandlers {
	return &AdminHandlers{
		knowledgeRepo: repo,
		knowledgeBase: kb,
	}
}

// RegisterRoutes registers HTTP routes
func (h *AdminHandlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/admin/knowledge", h.handleKnowledge)
	mux.HandleFunc("/admin/knowledge/", h.handleKnowledgeEntry)
	mux.HandleFunc("/admin/knowledge/search", h.handleKnowledgeSearch)
}

// handleKnowledge handles GET (list all) and POST (create) operations
func (h *AdminHandlers) handleKnowledge(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listKnowledgeEntries(w, r)
	case http.MethodPost:
		h.createKnowledgeEntry(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleKnowledgeEntry handles GET/PUT/DELETE operations on a specific entry
func (h *AdminHandlers) handleKnowledgeEntry(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/admin/knowledge/"):]
	if id == "" {
		http.Error(w, "Missing entry ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getKnowledgeEntry(w, r, id)
	case http.MethodPut:
		h.updateKnowledgeEntry(w, r, id)
	case http.MethodDelete:
		h.deleteKnowledgeEntry(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleKnowledgeSearch handles search operations
func (h *AdminHandlers) handleKnowledgeSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query parameter 'q'", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	entries, err := h.knowledgeRepo.SearchEntries(ctx, query)
	if err != nil {
		log.Printf("Error searching knowledge entries: %v", err)
		http.Error(w, "Error searching entries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// listKnowledgeEntries lists all knowledge entries
func (h *AdminHandlers) listKnowledgeEntries(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	entries, err := h.knowledgeRepo.GetAllEntries(ctx)
	if err != nil {
		log.Printf("Error fetching knowledge entries: %v", err)
		http.Error(w, "Error fetching entries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// getKnowledgeEntry fetches a specific entry by ID
func (h *AdminHandlers) getKnowledgeEntry(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	entry, err := h.knowledgeRepo.GetEntryByID(ctx, id)
	if err != nil {
		log.Printf("Error fetching knowledge entry: %v", err)
		http.Error(w, "Error fetching entry", http.StatusInternalServerError)
		return
	}

	if entry == nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// createKnowledgeEntry creates a new knowledge entry
func (h *AdminHandlers) createKnowledgeEntry(w http.ResponseWriter, r *http.Request) {
	var entry domain.KnowledgeEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.knowledgeRepo.CreateEntry(ctx, &entry); err != nil {
		log.Printf("Error creating knowledge entry: %v", err)
		http.Error(w, "Error creating entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)

	// Refresh knowledge base cache
	h.knowledgeBase.RefreshCache()
}

// updateKnowledgeEntry updates an existing entry
func (h *AdminHandlers) updateKnowledgeEntry(w http.ResponseWriter, r *http.Request, id string) {
	var entry domain.KnowledgeEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure ID in URL matches entry
	entry.ID = id

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.knowledgeRepo.UpdateEntry(ctx, &entry); err != nil {
		log.Printf("Error updating knowledge entry: %v", err)
		http.Error(w, "Error updating entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)

	// Refresh knowledge base cache
	h.knowledgeBase.RefreshCache()
}

// deleteKnowledgeEntry deletes an entry
func (h *AdminHandlers) deleteKnowledgeEntry(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.knowledgeRepo.DeleteEntry(ctx, id); err != nil {
		log.Printf("Error deleting knowledge entry: %v", err)
		http.Error(w, "Error deleting entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Refresh knowledge base cache
	h.knowledgeBase.RefreshCache()
}
