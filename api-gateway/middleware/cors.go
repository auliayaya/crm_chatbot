package middleware

import (
	"api-gateway/config"
	"log"
	"net/http"
	"strings"
)

var allowedOrigins []string

// InitCORS initializes CORS middleware
func InitCORS(cfg *config.Config) {
	allowedOrigins = strings.Split(cfg.AllowedOrigins, ",")
}

// CORS middleware handles Cross-Origin Resource Sharing
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Debug for WebSocket connections
		if strings.Contains(r.URL.Path, "/ws") {
			log.Printf("CORS middleware processing WebSocket request: %s", r.URL.Path)
		}

		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if origin != "" {
			allowOrigin := false

			// Check allowed origins
			for _, allowed := range allowedOrigins {
				if allowed == "*" || allowed == origin {
					allowOrigin = true
					break
				}
			}

			if allowOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
		}

		// For WebSocket upgrade requests, add specific headers
		if r.Header.Get("Upgrade") == "websocket" {
			log.Printf("WebSocket upgrade request detected")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
