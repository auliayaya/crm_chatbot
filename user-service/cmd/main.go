package main

import (
	"log"
	"net/http"
	"os"
	httpHandler "user-service/internal/adapters/primary/http"

	"user-service/internal/adapters/secondary/rabbitmq"
	"user-service/internal/adapters/secondary/repository"
	"user-service/internal/core/services"
)

func main() {
	// Initialize repositories
	userRepo := repository.NewPostgresUserRepository(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// Initialize RabbitMQ
	rabbitMQ := rabbitmq.NewRabbitMQClient(os.Getenv("RABBITMQ_URL"))
	defer rabbitMQ.Close()

	// Initialize services
	authService := services.NewAuthService(userRepo, []byte(os.Getenv("JWT_SECRET")))

	// Initialize HTTP handlers
	handler := httpHandler.NewHandler(authService, rabbitMQ)

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/register", handler.Register)
	mux.HandleFunc("/login", handler.Login)
	mux.HandleFunc("/verify", handler.VerifyToken)

	log.Println("User service running on :8082")
	log.Fatal(http.ListenAndServe("0.0.0.0:8082", mux))
}
