package main

import (
	"context"
	httphandlers "crm-service/internal/adapters/primary/http"
	"crm-service/internal/adapters/secondary/messaging"
	"crm-service/internal/adapters/secondary/repository"
	"crm-service/internal/config"
	"crm-service/internal/core/ports"
	"crm-service/internal/core/services"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create database connection
	repo, err := repository.NewPostgresRepository(
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	customerRepo := repository.NewCustomerRepository(repo.GetDB())
	ticketRepo := repository.NewTicketRepository(repo.GetDB())
	agentRepo := repository.NewAgentRepository(repo.GetDB())

	// Initialize messaging
	rabbitMQClient, err := messaging.NewRabbitMQClient(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v", err)
		log.Printf("Messaging functionality will be limited")
	}

	// Create message publisher
	var messagePublisher ports.MessagePublisher
	if rabbitMQClient != nil {
		messagePublisher = messaging.NewRabbitMQAdapter(rabbitMQClient)
		defer rabbitMQClient.Close()
	} else {
		messagePublisher = messaging.NewNoOpMessagePublisher()
	}

	// Create services
	customerService := services.NewCustomerService(customerRepo)
	ticketService := services.NewTicketService(ticketRepo, customerRepo, agentRepo, messagePublisher)
	agentService := services.NewAgentService(agentRepo, ticketRepo)

	// Create HTTP handlers
	handlers := httphandlers.NewHandlers(customerService, ticketService, agentService)

	// Create router and register routes
	router := httphandlers.NewRouter(handlers)

	// Configure HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("CRM Service starting on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for current operations to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}
