package http

import (
	"net/http"
	"time"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, generate a UUID
		// requestID := uuid.New().String()
		// w.Header().Set("X-Request-ID", requestID)
		// r = r.WithContext(context.WithValue(r.Context(), "requestID", requestID))
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a wrapper for the response writer to capture status code
		wrapped := newResponseWriter(w)

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Calculate duration after processing
		duration := time.Since(start)

		// Log request details after completion
		logger.Printf(
			"%s %s %d %s",
			r.Method,
			r.URL.Path,
			wrapped.status,
			duration,
		)
	})
}

// responseWriter is a wrapper for http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

// newResponseWriter creates a new responseWriter
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ContentTypeJSONMiddleware sets the Content-Type header to application/json
func ContentTypeJSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
