package integration

import (
	"api-gateway/config"
	"api-gateway/handlers"
	"api-gateway/middleware"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestFullAPIGateway(t *testing.T) {
	// Start mock services
	mockUserService := startMockUserService()
	defer mockUserService.Close()

	mockCRMService := startMockCRMService()
	defer mockCRMService.Close()

	mockChatService := startMockChatService()
	defer mockChatService.Close()

	// Configure gateway
	cfg := &config.Config{
		UserServiceURL: mockUserService.URL,
		CRMServiceURL:  mockCRMService.URL,
		ChatServiceURL: mockChatService.URL,
		JWTSecret:      "test-secret",
		AllowedOrigins: "*",
	}

	// Initialize middleware
	middleware.InitAuth(cfg)
	middleware.InitCORS(cfg)

	// Create main router
	router := mux.NewRouter()

	// Add middleware
	router.Use(middleware.Logging)
	router.Use(middleware.CORS)

	// Health check endpoint
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// Auth routes
	authRouter := router.PathPrefix("/auth").Subrouter()
	handlers.RegisterAuthRoutes(authRouter, cfg.UserServiceURL)

	// Protected routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.Authentication)

	// Register service routes
	handlers.RegisterCRMRoutes(apiRouter.PathPrefix("/crm").Subrouter(), cfg.CRMServiceURL)
	handlers.RegisterChatRoutes(apiRouter.PathPrefix("/chat").Subrouter(), cfg.ChatServiceURL)

	// Admin routes
	adminRouter := apiRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AdminOnly)
	handlers.RegisterAdminRoutes(adminRouter, cfg)

	// Create test server
	gateway := httptest.NewServer(router)
	defer gateway.Close()

	// Run integration tests
	t.Run("Health check returns OK", func(t *testing.T) {
		resp, err := http.Get(gateway.URL + "/health")
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Authentication flow works", func(t *testing.T) {
		// 1. Login to get token
		loginBody := map[string]string{
			"email":    "user@example.com",
			"password": "password",
		}
		loginJSON, _ := json.Marshal(loginBody)

		resp, err := http.Post(gateway.URL+"/auth/login", "application/json", bytes.NewBuffer(loginJSON))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse token from response
		var loginResponse struct {
			Token string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&loginResponse)

		// 2. Use token to access protected endpoint
		req, _ := http.NewRequest("GET", gateway.URL+"/api/crm/customers", nil)
		req.Header.Set("Authorization", "Bearer "+loginResponse.Token)

		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Register user", func(t *testing.T) {
		registerBody := map[string]string{
			"email":     "newuser@example.com",
			"password":  "securepassword",
			"firstName": "John",
			"lastName":  "Doe",
		}
		registerJSON, _ := json.Marshal(registerBody)

		resp, err := http.Post(gateway.URL+"/auth/register", "application/json", bytes.NewBuffer(registerJSON))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify response contains user ID
		var registerResponse struct {
			UserID  string `json:"user_id"`
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&registerResponse)
		assert.NotEmpty(t, registerResponse.UserID)
	})

	t.Run("Forgot password", func(t *testing.T) {
		forgotBody := map[string]string{
			"email": "user@example.com",
		}
		forgotJSON, _ := json.Marshal(forgotBody)

		resp, err := http.Post(gateway.URL+"/auth/forgot-password", "application/json", bytes.NewBuffer(forgotJSON))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check for success message
		var forgotResponse struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&forgotResponse)
		assert.True(t, forgotResponse.Success)
	})

	t.Run("Reset password", func(t *testing.T) {
		resetBody := map[string]string{
			"token":    "valid-reset-token",
			"password": "newpassword123",
		}
		resetJSON, _ := json.Marshal(resetBody)

		resp, err := http.Post(gateway.URL+"/auth/reset-password", "application/json", bytes.NewBuffer(resetJSON))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Check for success message
		var resetResponse struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&resetResponse)
		assert.True(t, resetResponse.Success)
	})

	t.Run("Refresh token", func(t *testing.T) {
		// First get a valid token through login
		loginBody := map[string]string{
			"email":    "user@example.com",
			"password": "password",
		}
		loginJSON, _ := json.Marshal(loginBody)

		loginResp, err := http.Post(gateway.URL+"/auth/login", "application/json", bytes.NewBuffer(loginJSON))
		assert.NoError(t, err)

		var loginResponse struct {
			Token        string `json:"token"`
			RefreshToken string `json:"refresh_token"`
		}
		json.NewDecoder(loginResp.Body).Decode(&loginResponse)
		loginResp.Body.Close()

		// Now use the refresh token to get a new access token
		refreshBody := map[string]string{
			"refresh_token": loginResponse.RefreshToken,
		}
		refreshJSON, _ := json.Marshal(refreshBody)

		resp, err := http.Post(gateway.URL+"/auth/refresh-token", "application/json", bytes.NewBuffer(refreshJSON))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify we get a new token
		var refreshResponse struct {
			Token string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&refreshResponse)
		assert.NotEmpty(t, refreshResponse.Token)
	})

	// Test CRM endpoints
	t.Run("CRM: Get customer list", func(t *testing.T) {
		// Get a token first
		token := getAuthToken(t, gateway.URL)

		req, _ := http.NewRequest("GET", gateway.URL+"/api/crm/customers", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("CRM: Create ticket", func(t *testing.T) {
		token := getAuthToken(t, gateway.URL)

		ticketBody := map[string]interface{}{
			"customer_id": "1",
			"subject":     "Test Ticket from API Gateway Test",
			"description": "This is a test ticket created during API Gateway testing",
			"priority":    "high",
		}
		ticketJSON, _ := json.Marshal(ticketBody)

		req, _ := http.NewRequest("POST", gateway.URL+"/api/crm/tickets", bytes.NewBuffer(ticketJSON))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test Chat endpoints
	t.Run("Chat: Get sessions", func(t *testing.T) {
		token := getAuthToken(t, gateway.URL)

		req, _ := http.NewRequest("GET", gateway.URL+"/api/chat/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// Helper function to get an auth token for testing
func getAuthToken(t *testing.T, gatewayURL string) string {
	loginBody := map[string]string{
		"email":    "user@example.com",
		"password": "password",
	}
	loginJSON, _ := json.Marshal(loginBody)

	resp, err := http.Post(gatewayURL+"/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	assert.NoError(t, err)
	defer resp.Body.Close()

	var loginResponse struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResponse)

	return loginResponse.Token
}

// Helper functions to start mock services

func startMockUserService() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Handle login requests
		if r.URL.Path == "/login" && r.Method == "POST" {
			// Generate a valid JWT token using the same secret as the middleware
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub":   "test-user-id",
				"roles": "user",
				"exp":   time.Now().Add(time.Hour * 24).Unix(),
			})

			// Sign the token with the test secret
			tokenString, _ := token.SignedString([]byte("test-secret"))

			refreshToken := "valid-refresh-token-" + tokenString[10:20]

			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
                "token": "%s",
                "refresh_token": "%s",
                "user_id": "test-user-id",
                "expires_in": 86400
            }`, tokenString, refreshToken)
			return
		}

		// Handle register requests
		if r.URL.Path == "/register" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
                "user_id": "new-user-123",
                "message": "User registered successfully",
                "success": true
            }`))
			return
		}

		// Handle forgot password requests
		if r.URL.Path == "/forgot-password" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
                "success": true,
                "message": "If the email exists, a password reset link has been sent"
            }`))
			return
		}

		// Handle reset password requests
		if r.URL.Path == "/reset-password" && r.Method == "POST" {
			// Parse the request to verify the token
			var resetReq struct {
				Token string `json:"token"`
			}
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &resetReq)

			// Check if the token is the expected one
			if resetReq.Token == "valid-reset-token" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
                    "success": true,
                    "message": "Password has been reset successfully"
                }`))
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{
                    "success": false,
                    "message": "Invalid or expired token"
                }`))
			}
			return
		}

		// Handle refresh token requests
		if r.URL.Path == "/refresh-token" && r.Method == "POST" {
			// Check the refresh token
			var refreshReq struct {
				RefreshToken string `json:"refresh_token"`
			}
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &refreshReq)

			if strings.HasPrefix(refreshReq.RefreshToken, "valid-refresh-token-") {
				// Generate a new token
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"sub":   "test-user-id",
					"roles": "user",
					"exp":   time.Now().Add(time.Hour * 24).Unix(),
				})
				tokenString, _ := token.SignedString([]byte("test-secret"))

				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{
                    "token": "%s",
                    "expires_in": 86400
                }`, tokenString)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{
                    "error": "invalid_grant",
                    "message": "Invalid refresh token"
                }`))
			}
			return
		}

		// Handle other user service endpoints
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok", "path": "%s", "method": "%s"}`, r.URL.Path, r.Method)
	}))
}

func startMockCRMService() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check auth headers
		if r.Header.Get("X-User-ID") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Different responses based on path
		switch {
		case r.URL.Path == "/customers":
			w.Write([]byte(`{"customers": [{"id": "1", "name": "Test Customer"}]}`))
		case r.URL.Path == "/tickets":
			w.Write([]byte(`{"tickets": [{"id": "101", "subject": "Test Ticket"}]}`))
		default:
			fmt.Fprintf(w, `{"status": "ok", "path": "%s", "method": "%s"}`, r.URL.Path, r.Method)
		}
	}))
}

func startMockChatService() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check auth headers
		if r.Header.Get("X-User-ID") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Different responses based on path
		switch {
		case r.URL.Path == "/sessions":
			w.Write([]byte(`{"sessions": [{"id": "201", "customer_id": "1"}]}`))
		case r.URL.Path == "/messages":
			w.Write([]byte(`{"messages": [{"id": "301", "content": "Test message"}]}`))
		default:
			fmt.Fprintf(w, `{"status": "ok", "path": "%s", "method": "%s"}`, r.URL.Path, r.Method)
		}
	}))
}
