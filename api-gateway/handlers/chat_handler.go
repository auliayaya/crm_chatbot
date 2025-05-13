package handlers

import (
	"api-gateway/proxy"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	// Keep for createHTTPProxy if needed

	"github.com/gorilla/mux"
)

// RegisterChatRoutes registers routes for the Chat service
// chatServiceURL should be the base URL of the chat service, e.g., "http://chat-service:8081"
func RegisterChatRoutes(router *mux.Router, chatServiceURL string) {
	// General HTTP proxy for regular endpoints
	httpProxy, err := createHTTPProxy(chatServiceURL)
	if err != nil {
		log.Fatalf("Failed to create Chat service HTTP proxy: %v", err)
	}

	// Regular HTTP endpoints (assuming they are relative to chatServiceURL)
	router.PathPrefix("/sessions").Handler(httpProxy)
	router.PathPrefix("/messages").Handler(httpProxy)

	// WebSocket endpoint
	// The target path on the backend chat service is "/ws"
	wsProxyHandler := proxy.WebSocketProxy(chatServiceURL, "/ws")

	// This route handles requests like /api/chat/ws (if router is mounted under /api/chat)
	// or just /ws (if router is the main router or mounted at root for this path)
	router.Handle("/ws", wsProxyHandler).Methods("GET")

	log.Printf("Registered WebSocket handler for path ending in /ws to proxy to %s/ws", chatServiceURL)
}

// createHTTPProxy is for non-WebSocket endpoints
func createHTTPProxy(targetBaseURL string) (http.Handler, error) {
	targetURL, err := url.Parse(targetBaseURL)
	if err != nil {
		return nil, err
	}

	p := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := p.Director
	p.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host // Important for the backend

		// Path adjustment logic:
		// If this handler is mounted under "/api/chat" by a parent router,
		// and a request comes for "/api/chat/sessions/123",
		// req.URL.Path might be "/api/chat/sessions/123".
		// If the backend expects "/sessions/123", you need to strip "/api/chat".
		// This depends on your main router setup.
		// For now, assuming the path passed to this director is what the backend expects
		// relative to targetBaseURL, or that the main router handles stripping.
		// Example stripping (if needed):
		// if strings.HasPrefix(req.URL.Path, "/api/chat/") {
		// req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/chat")
		// }
		log.Printf("HTTPProxy Director: Forwarding to %s%s", req.URL.Host, req.URL.Path)
	}
	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("HTTPProxy ErrorHandler: %v for request %s", err, r.URL.Path)
		http.Error(w, "Proxy error", http.StatusBadGateway)
	}
	return p, nil
}
