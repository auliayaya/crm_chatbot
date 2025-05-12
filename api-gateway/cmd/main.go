package main

import (
	"api-gateway/config"
	"api-gateway/handlers"
	"api-gateway/middleware"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	log.Printf("Starting API Gateway on port %s", cfg.Port)
	log.Printf("Services: CRM=%s, Chat=%s, User=%s",
		cfg.CRMServiceURL, cfg.ChatServiceURL, cfg.UserServiceURL)

	// Create router
	router := mux.NewRouter()

	// Add middleware
	router.Use(middleware.Logging)
	router.Use(middleware.CORS)

	// Health check endpoint
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// API documentation
	router.HandleFunc("/api-docs", handlers.APIDocumentation).Methods("GET")

	// Set up auth routes (no auth required)
	authRouter := router.PathPrefix("/auth").Subrouter()
	handlers.RegisterAuthRoutes(authRouter, cfg.UserServiceURL)

	// Protected routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	authMiddleware := middleware.NewAuthenticationMiddleware([]byte(cfg.JWTSecret))
	apiRouter.Use(authMiddleware)
	apiRouter.Use(middleware.RateLimit)

	// Register routes for each service
	handlers.RegisterCRMRoutes(apiRouter.PathPrefix("/crm").Subrouter(), cfg.CRMServiceURL)
	handlers.RegisterChatRoutes(apiRouter.PathPrefix("/chat").Subrouter(), cfg.ChatServiceURL)

	// Admin routes with stricter authentication
	adminRouter := apiRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AdminOnly)
	handlers.RegisterAdminRoutes(adminRouter, cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down API Gateway...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("API Gateway stopped")
}
