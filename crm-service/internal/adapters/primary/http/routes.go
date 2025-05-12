package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// NewRouter creates and configures a new HTTP router
func NewRouter(handlers *Handlers) http.Handler {
	router := mux.NewRouter()

	// Add middleware
	router.Use(loggingMiddleware)
	router.Use(jsonContentTypeMiddleware)

	// Health check
	router.HandleFunc("/health", handlers.HealthCheck).Methods(http.MethodGet)

	// Customer routes
	router.HandleFunc("/customers", handlers.GetCustomers).Methods(http.MethodGet)
	router.HandleFunc("/customers", handlers.CreateCustomer).Methods(http.MethodPost)
	router.HandleFunc("/customers/{id}", handlers.GetCustomer).Methods(http.MethodGet)
	router.HandleFunc("/customers/{id}", handlers.UpdateCustomer).Methods(http.MethodPut)
	router.HandleFunc("/customers/{id}", handlers.DeleteCustomer).Methods(http.MethodDelete)
	router.HandleFunc("/customers/search", handlers.SearchCustomers).Methods(http.MethodGet)

	// Ticket routes
	router.HandleFunc("/tickets", handlers.GetTickets).Methods(http.MethodGet)
	router.HandleFunc("/tickets", handlers.CreateTicket).Methods(http.MethodPost)
	router.HandleFunc("/tickets/{id}", handlers.GetTicket).Methods(http.MethodGet)
	router.HandleFunc("/tickets/{id}", handlers.UpdateTicket).Methods(http.MethodPut)
	router.HandleFunc("/tickets/{id}/assign", handlers.AssignTicket).Methods(http.MethodPost)
	router.HandleFunc("/tickets/{id}/comments", handlers.AddTicketComment).Methods(http.MethodPost)
	router.HandleFunc("/tickets/{id}/close", handlers.CloseTicket).Methods(http.MethodPost)
	router.HandleFunc("/tickets/{id}/history", handlers.GetTicketHistory).Methods(http.MethodGet)
	router.HandleFunc("/customers/{id}/tickets", handlers.GetCustomerTickets).Methods(http.MethodGet)

	// Agent routes
	router.HandleFunc("/agents", handlers.GetAgents).Methods(http.MethodGet)
	router.HandleFunc("/agents", handlers.CreateAgent).Methods(http.MethodPost)
	router.HandleFunc("/agents/{id}", handlers.GetAgent).Methods(http.MethodGet)
	router.HandleFunc("/agents/{id}", handlers.UpdateAgent).Methods(http.MethodPut)
	router.HandleFunc("/agents/workloads", handlers.GetAgentWorkloads).Methods(http.MethodGet)

	router.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	return router
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		// Log after request is complete
		duration := time.Since(start)

		// Log request details
		// Format: timestamp method path duration
		// e.g. 2023-04-05 12:34:56 GET /customers 17.523ms
		logger.Printf(
			"%s %s %s %s",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.RequestURI,
			duration,
		)
	})
}

// jsonContentTypeMiddleware sets the Content-Type header to application/json
func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
