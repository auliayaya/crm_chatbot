// filepath: /Users/auliaillahi/VSCodeProject/GenAIBuild/crm-chatbot/api-gateway/middleware/logging.go
package middleware

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// Ensure responseWriter implements http.Hijacker interface
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	// Cast the original ResponseWriter to http.Hijacker and call its Hijack method
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Logging is middleware to log all requests
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// For WebSocket connections, bypass wrapping to avoid Hijacker issues
		if r.Header.Get("Upgrade") == "websocket" {
			log.Printf("Logging Middleware: WebSocket request from %s: %s %s. Bypassing response writer wrapping.", r.RemoteAddr, r.Method, r.URL.Path)
			next.ServeHTTP(w, r) // Pass original http.ResponseWriter
			return
		}

		// Normal HTTP requests use wrapped ResponseWriter
		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK, // Default to 200 OK
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Printf("%s %s %s %d %d %s", r.RemoteAddr, r.Method, r.URL.Path, rw.status, rw.size, duration)
	})
}
