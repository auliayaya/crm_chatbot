package main

import (
	"api-gateway/config"
	"api-gateway/handlers"
	"api-gateway/middleware"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	log.Printf("Starting API Gateway on port %s", cfg.Port)
	log.Printf("Services: CRM=%s, Chat=%s, User=%s",
		cfg.CRMServiceURL, cfg.ChatServiceURL, cfg.UserServiceURL)
	middleware.InitCORS(cfg)

	// Clear handler for WebSocket connections - this bypasses all middleware
	http.HandleFunc("/chat/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("WebSocket connection request from %s to %s", r.RemoteAddr, r.URL.Path)

		// Parse the target URL
		targetURL := url.URL{
			Scheme:   "ws",
			Host:     "chat-service:8081",
			Path:     "/ws",
			RawQuery: r.URL.RawQuery,
		}

		log.Printf("Proxying WebSocket from %s to %s", r.URL.String(), targetURL.String())

		// Create a WebSocket dialer with options
		dialer := websocket.DefaultDialer

		// Copy headers to make sure we include auth if needed
		requestHeader := http.Header{}
		for k, vs := range r.Header {
			if k == "Sec-Websocket-Key" ||
				k == "Sec-Websocket-Version" ||
				k == "Sec-Websocket-Extensions" ||
				k == "Sec-Websocket-Protocol" ||
				k == "Upgrade" ||
				k == "Connection" {
				// Skip WebSocket headers as they'll be set by the dialer
				continue
			}
			for _, v := range vs {
				requestHeader.Add(k, v)
			}
		}

		// Connect to the backend WebSocket server
		backConn, resp, err := dialer.Dial(targetURL.String(), requestHeader)
		if err != nil {
			log.Printf("Failed to connect to backend: %v", err)
			if resp != nil {
				log.Printf("Backend response: %d %s", resp.StatusCode, resp.Status)
				// Copy error response
				for k, v := range resp.Header {
					w.Header()[k] = v
				}
				w.WriteHeader(resp.StatusCode)
				io.Copy(w, resp.Body)
			} else {
				http.Error(w, "Failed to connect to backend", http.StatusBadGateway)
			}
			return
		}
		defer backConn.Close()

		// Define WebSocket upgrader for client connection
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for testing
			},
		}

		// Upgrade the client connection
		clientConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade client connection: %v", err)
			return
		}
		defer clientConn.Close()

		log.Printf("WebSocket connection established")

		// Function to copy messages from one connection to another
		copyWebSocketMessages := func(dest *websocket.Conn, src *websocket.Conn, done chan<- bool) {
			defer func() { done <- true }()
			for {
				messageType, message, err := src.ReadMessage()
				if err != nil {
					if !websocket.IsCloseError(err,
						websocket.CloseNormalClosure,
						websocket.CloseGoingAway) {
						log.Printf("Error reading from WebSocket: %v", err)
					}
					break
				}

				err = dest.WriteMessage(messageType, message)
				if err != nil {
					log.Printf("Error writing to WebSocket: %v", err)
					break
				}

				// Log message for debugging
				if messageType == websocket.TextMessage {
					log.Printf("WebSocket message: %s", string(message))
				}
			}
		}

		// Create channels to signal when connections are closed
		clientDone := make(chan bool, 1)
		backendDone := make(chan bool, 1)

		// Copy messages in both directions
		go copyWebSocketMessages(backConn, clientConn, clientDone)
		go copyWebSocketMessages(clientConn, backConn, backendDone)

		// Wait for either connection to close
		select {
		case <-clientDone:
			log.Printf("Client connection closed")
		case <-backendDone:
			log.Printf("Backend connection closed")
		}
	})

	// Create router
	router := mux.NewRouter()

	// Add middleware
	router.Use(middleware.Logging)
	router.Use(middleware.CORS)

	// Health check and other public endpoints
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
	router.HandleFunc("/api-docs", handlers.APIDocumentation).Methods("GET")

	// Keep your existing /chat/ws handler as well
	router.HandleFunc("/chat/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Forwarding WebSocket request to chat service from /chat/ws")
		// Forward to chat service directly
		targetURL := cfg.ChatServiceURL + "/ws?" + r.URL.RawQuery

		// Create a reverse proxy for WebSocket
		target, _ := url.Parse(targetURL)
		proxy := httputil.NewSingleHostReverseProxy(target)

		// Update the request URL
		r.URL.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.URL.Path = "/ws"
		r.Host = target.Host

		// Make sure the connection header is preserved
		r.Header.Set("Connection", "Upgrade")
		r.Header.Set("Upgrade", "websocket")

		// Forward the request
		proxy.ServeHTTP(w, r)
	}).Methods("GET")

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

	// Log registered routes
	logRegisteredRoutes(router)

	// Create a custom handler that routes WebSocket requests to the specific handler
	// and all other requests to the router
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Let the dedicated handler process WebSocket requests
		if r.URL.Path == "/chat/ws" && strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
			http.DefaultServeMux.ServeHTTP(w, r)
			return
		}

		// Use router for all other paths
		router.ServeHTTP(w, r)
	})

	// Configure your server with the combined handler
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      finalHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
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

func logRegisteredRoutes(r *mux.Router) {
	log.Println("=== REGISTERED API ROUTES ===")
	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			pathTemplate = "<unknown path>"
		}

		methods, err := route.GetMethods()
		if err != nil {
			methods = []string{"ANY"}
		}

		// Skip subrouters without specific path templates
		if pathTemplate == "/" && len(ancestors) > 0 {
			return nil
		}

		// Check if route has a handler
		handler := route.GetHandler()
		hasHandler := handler != nil

		// Print route details
		for _, method := range methods {
			middlewares := ""
			if len(ancestors) > 0 {
				middlewares = fmt.Sprintf(" (Middlewares: %d)", len(ancestors))
			}

			status := "✓"
			if !hasHandler {
				status = "✗"
			}

			log.Printf("[%s] %s\t%s %s", status, method, pathTemplate, middlewares)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking routes: %v", err)
	}
	log.Println("=============================")
}
