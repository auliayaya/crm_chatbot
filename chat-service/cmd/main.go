package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	httphandlers "chat-service/internal/adapters/primary/http"
	"chat-service/internal/adapters/primary/websocket"
	"chat-service/internal/adapters/secondary/messaging"
	"chat-service/internal/adapters/secondary/repository"
	"chat-service/internal/config"
	"chat-service/internal/core/services"
)

func main() {
	cfg := config.LoadConfig()

	repo, err := repository.NewPostgresRepository(
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	rabbitMQClient, err := messaging.NewRabbitMQClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQClient.Close()
	messagePublisher := messaging.NewRabbitMQAdapter(rabbitMQClient)

	messageRepository := repo

	// Create knowledge repository with error handling
	knowledgeRepo := repository.NewPostgresKnowledgeRepository(repo.GetDB())

	// Initialize schema with proper timeout handling
	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := knowledgeRepo.InitSchema(initCtx); err != nil {
		log.Printf("Warning: Failed to initialize knowledge base schema: %v", err)
		log.Printf("Knowledge base functionality may be limited")
	}
	initCancel()

	// Create knowledge base service with repository
	knowledgeBase := services.NewKnowledgeBase(knowledgeRepo)

	// Create other services
	chatService := services.NewChatService(messageRepository, messageRepository, messagePublisher)

	// Pass knowledge base to bot agent
	botAgent := services.NewBotAgent("bot-1", "Support Bot", cfg.UseAI,
		messageRepository, messagePublisher, knowledgeBase)

	if cfg.UseAI {
		fmt.Println("Setting up AI client")
		botAgent.SetupAIClient(cfg.OpenAIKey)
	}

	hub := websocket.NewHub(chatService, botAgent)
	botAgent.SetHub(hub) // Connect hub to bot agent

	if err := hub.SubscribeToBotMessages(); err != nil {
		log.Fatalf("Failed to subscribe to messages: %v", err)
	}


	// Create admin handlers with repository
	adminHandlers := httphandlers.NewAdminHandlers(knowledgeRepo, knowledgeBase)
	adminHandlers.RegisterRoutes(http.DefaultServeMux)

	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWS(hub, w, r)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Chat service running on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
