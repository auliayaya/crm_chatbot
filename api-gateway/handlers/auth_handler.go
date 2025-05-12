package handlers

import (
	"api-gateway/proxy"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterAuthRoutes registers authentication-related routes
func RegisterAuthRoutes(router *mux.Router, userServiceURL string) {
	authProxy, err := proxy.NewReverseProxy(userServiceURL, "/auth")
	if err != nil {
		log.Fatalf("Failed to create authentication proxy: %v", err)
	}

	// Login endpoint
	router.Path("/login").Handler(authProxy)

	// Register endpoint
	router.Path("/register").Handler(authProxy)

	// Password reset endpoints
	router.Path("/forgot-password").Handler(authProxy)
	router.Path("/reset-password").Handler(authProxy)

	// Refresh token
	router.Path("/refresh-token").Handler(authProxy)
}

// RegisterAdminRoutes registers admin routes
func RegisterAdminRoutes(router *mux.Router, cfg interface{}) {
	// Admin dashboard routes
	router.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "admin dashboard"}`))
	}).Methods("GET")

	// System monitoring
	router.HandleFunc("/system", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "system monitoring"}`))
	}).Methods("GET")
}
