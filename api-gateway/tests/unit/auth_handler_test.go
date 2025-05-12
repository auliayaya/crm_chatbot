package unit

import (
    "api-gateway/handlers"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gorilla/mux"
    "github.com/stretchr/testify/assert"
)

// Test RegisterAuthRoutes to ensure it configures all expected routes
func TestRegisterAuthRoutes(t *testing.T) {
    router := mux.NewRouter()
    
    // Start mock user service
    mockUserSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    }))
    defer mockUserSrv.Close()
    
    // Register auth routes with the mock server URL
    handlers.RegisterAuthRoutes(router.PathPrefix("/auth").Subrouter(), mockUserSrv.URL)
    
    // Create test server using the router
    testServer := httptest.NewServer(router)
    defer testServer.Close()
    
    // Test cases for auth endpoints
    testCases := []struct {
        name     string
        path     string
        method   string
        expected int
    }{
        {"Login", "/auth/login", http.MethodPost, http.StatusOK},
        {"Register", "/auth/register", http.MethodPost, http.StatusOK},
        {"ForgotPassword", "/auth/forgot-password", http.MethodPost, http.StatusOK},
        {"ResetPassword", "/auth/reset-password", http.MethodPost, http.StatusOK},
        {"RefreshToken", "/auth/refresh-token", http.MethodPost, http.StatusOK},
    }
    
    // Run tests for each endpoint
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            req, err := http.NewRequest(tc.method, testServer.URL+tc.path, nil)
            assert.NoError(t, err)
            
            resp, err := http.DefaultClient.Do(req)
            assert.NoError(t, err)
            defer resp.Body.Close()
            
            assert.Equal(t, tc.expected, resp.StatusCode)
        })
    }
}

// Test RegisterAdminRoutes to ensure admin routes are configured
func TestRegisterAdminRoutes(t *testing.T) {
    router := mux.NewRouter()
    
    // Register admin routes
    handlers.RegisterAdminRoutes(router.PathPrefix("/admin").Subrouter(), nil)
    
    // Create test server
    testServer := httptest.NewServer(router)
    defer testServer.Close()
    
    // Test cases for admin endpoints
    testCases := []struct {
        name     string
        path     string
        method   string
        expected int
    }{
        {"Dashboard", "/admin/dashboard", http.MethodGet, http.StatusOK},
        {"System", "/admin/system", http.MethodGet, http.StatusOK},
    }
    
    // Run tests for each endpoint
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            req, err := http.NewRequest(tc.method, testServer.URL+tc.path, nil)
            assert.NoError(t, err)
            
            resp, err := http.DefaultClient.Do(req)
            assert.NoError(t, err)
            defer resp.Body.Close()
            
            assert.Equal(t, tc.expected, resp.StatusCode)
        })
    }
}