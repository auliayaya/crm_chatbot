package proxy

import (
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
)

// NewReverseProxy creates a new reverse proxy to a target service
func NewReverseProxy(targetURL string, stripPrefix string) (*httputil.ReverseProxy, error) {
    url, err := url.Parse(targetURL)
    if err != nil {
        return nil, err
    }
    
    proxy := httputil.NewSingleHostReverseProxy(url)
    
    // Modify director to handle path rewriting
    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        
        // Strip API prefix
        if stripPrefix != "" {
            req.URL.Path = strings.TrimPrefix(req.URL.Path, stripPrefix)
            // Ensure path starts with /
            if !strings.HasPrefix(req.URL.Path, "/") {
                req.URL.Path = "/" + req.URL.Path
            }
        }
        
        // Forward user info from JWT
        if userID := req.Context().Value("user_id"); userID != nil {
            req.Header.Set("X-User-ID", userID.(string))
        }
        
        // Forward user roles if available
        if roles := req.Context().Value("user_roles"); roles != nil {
            req.Header.Set("X-User-Roles", roles.(string))
        }
    }
    
    // Custom error handler
    proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
        w.WriteHeader(http.StatusBadGateway)
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"error": "Service Unavailable", "status": 502}`))
    }
    
    return proxy, nil
}