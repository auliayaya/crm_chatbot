// filepath: /Users/auliaillahi/VSCodeProject/GenAIBuild/crm-chatbot/api-gateway/middleware/logging.go
package middleware

import (
    "log"
    "net/http"
    "time"
)

// Logging middleware logs request details
func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Create response wrapper to capture status code
        wrapped := newResponseWriter(w)
        
        // Process request
        next.ServeHTTP(wrapped, r)
        
        // Log request details
        duration := time.Since(start)
        log.Printf(
            "%s %s %s %d %s %s",
            r.RemoteAddr,
            r.Method,
            r.URL.Path,
            wrapped.status,
            duration,
            r.UserAgent(),
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