package handlers

import (
    "api-gateway/proxy"
    "log"


    "github.com/gorilla/mux"
)

// RegisterChatRoutes registers routes for the Chat service
func RegisterChatRoutes(router *mux.Router, chatServiceURL string) {
    chatProxy, err := proxy.NewReverseProxy(chatServiceURL, "/api/chat")
    if err != nil {
        log.Fatalf("Failed to create Chat service proxy: %v", err)
    }
    
    // Chat session endpoints
    router.PathPrefix("/sessions").Handler(chatProxy)
    
    // Message endpoints
    router.PathPrefix("/messages").Handler(chatProxy)
    
    // WebSocket endpoint for real-time chat
    router.PathPrefix("/ws").Handler(chatProxy)
}