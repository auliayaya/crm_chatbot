package http

import (
	"crm-service/internal/core/domain"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GetCustomers handles GET /customers
func (h *Handlers) GetCustomers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 100 // Default limit
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0 // Default offset
	}

	// Get customers from service
	customers, err := h.customerService.GetCustomers(r.Context(), limit, offset)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve customers")
		return
	}

	h.respondWithJSON(w, http.StatusOK, customers)
}

// GetCustomer handles GET /customers/{id}
func (h *Handlers) GetCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	customer, err := h.customerService.GetCustomerByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve customer")
		return
	}

	if customer == nil {
		h.respondWithError(w, http.StatusNotFound, "Customer not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, customer)
}

// CreateCustomer handles POST /customers
func (h *Handlers) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer domain.Customer

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if customer.Email == "" || customer.FirstName == "" || customer.LastName == "" {
		h.respondWithError(w, http.StatusBadRequest, "Email, first name, and last name are required")
		return
	}

	// Create customer
	if err := h.customerService.CreateCustomer(r.Context(), &customer); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create customer")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, customer)
}

// UpdateCustomer handles PUT /customers/{id}
func (h *Handlers) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var customer domain.Customer

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Ensure ID in path matches body
	customer.ID = id

	// Update customer
	if err := h.customerService.UpdateCustomer(r.Context(), &customer); err != nil {
		if err == domain.ErrNotFound {
			h.respondWithError(w, http.StatusNotFound, "Customer not found")
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to update customer")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, customer)
}

// DeleteCustomer handles DELETE /customers/{id}
func (h *Handlers) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Delete customer
	if err := h.customerService.DeleteCustomer(r.Context(), id); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete customer")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// SearchCustomers handles GET /customers/search
func (h *Handlers) SearchCustomers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.respondWithError(w, http.StatusBadRequest, "Search query parameter 'q' is required")
		return
	}

	customers, err := h.customerService.SearchCustomers(r.Context(), query)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to search customers")
		return
	}

	h.respondWithJSON(w, http.StatusOK, customers)
}
