
package middleware

import (
    "net/http"
    "sync"
    "time"
)

// Simple rate limiter implementation
type rateLimiter struct {
    requestCount  map[string]int
    requestTimers map[string]time.Time
    mu            sync.Mutex
    limit         int           // Requests per period
    period        time.Duration // Time period for rate limiting
}

var limiter = &rateLimiter{
    requestCount:  make(map[string]int),
    requestTimers: make(map[string]time.Time),
    limit:         100,         // 100 requests per minute by default
    period:        time.Minute, // 1 minute period
}

// RateLimit middleware limits the number of requests per IP address
func RateLimit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get IP address from request
        ip := r.RemoteAddr
        
        // Also consider X-Forwarded-For header if behind proxy
        forwardedFor := r.Header.Get("X-Forwarded-For")
        if forwardedFor != "" {
            ip = forwardedFor
        }
        
        // Check if request is allowed
        if !isAllowed(ip) {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusTooManyRequests)
            w.Write([]byte(`{"error": "Rate limit exceeded", "status": 429}`))
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// isAllowed checks if the request from the given IP is allowed
func isAllowed(ip string) bool {
    limiter.mu.Lock()
    defer limiter.mu.Unlock()
    
    now := time.Now()
    lastRequest, exists := limiter.requestTimers[ip]
    
    // If this is a new IP or the period has passed, reset count
    if !exists || now.Sub(lastRequest) > limiter.period {
        limiter.requestCount[ip] = 1
        limiter.requestTimers[ip] = now
        return true
    }
    
    // Increment request count
    limiter.requestCount[ip]++
    limiter.requestTimers[ip] = now
    
    // Check if limit is exceeded
    return limiter.requestCount[ip] <= limiter.limit
}