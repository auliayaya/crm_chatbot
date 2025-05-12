package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	httpHandler "user-service/internal/adapters/primary/http"
	"user-service/internal/adapters/secondary/repository"
	"user-service/internal/core/services"
)

// MockRabbitMQ implements a mock version of the RabbitMQ client
type MockRabbitMQ struct{}

func (m *MockRabbitMQ) PublishUserEvent(eventType string, payload interface{}) error {
	return nil
}

func (m *MockRabbitMQ) Close() {}

func setupTestServer() http.Handler {
	// Use in-memory repository for testing
	repo := repository.NewInMemoryUserRepo()
	rabbitMQ := &MockRabbitMQ{}
	authService := services.NewAuthService(repo, []byte("test-secret"))
	handler := httpHandler.NewHandler(authService, rabbitMQ)

	// Create a router for our endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("/register", handler.Register)
	mux.HandleFunc("/login", handler.Login)
	mux.HandleFunc("/verify", handler.VerifyToken)

	return mux
}

func TestRegister(t *testing.T) {
	server := setupTestServer()

	t.Run("successful registration", func(t *testing.T) {
		// Create registration payload
		payload := map[string]string{
			"email":    "test@example.com",
			"username": "testuser",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(payload)

		// Create request
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Record response
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		// Check status code
		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status %d; got %d", http.StatusCreated, rec.Code)
		}

		// Check response body
		var response map[string]string
		json.Unmarshal(rec.Body.Bytes(), &response)
		if msg, exists := response["message"]; !exists || msg != "User registered successfully" {
			t.Error("Unexpected response message")
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		// Create incomplete registration payload
		payload := map[string]string{
			"email": "test@example.com",
			// Missing username and password
		}
		jsonData, _ := json.Marshal(payload)

		// Create request
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Record response
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		// Check status code
		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("duplicate user", func(t *testing.T) {
		// Register first user
		payload := map[string]string{
			"email":    "duplicate@example.com",
			"username": "duplicate",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		// Try to register same user again
		req = httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		// Check status code
		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, rec.Code)
		}
	})
}

func TestLogin(t *testing.T) {
	server := setupTestServer()

	// Register a user for login tests
	registerPayload := map[string]string{
		"email":    "login@example.com",
		"username": "loginuser",
		"password": "password123",
	}
	jsonRegister, _ := json.Marshal(registerPayload)

	reqRegister := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonRegister))
	reqRegister.Header.Set("Content-Type", "application/json")
	recRegister := httptest.NewRecorder()
	server.ServeHTTP(recRegister, reqRegister)

	t.Run("successful login", func(t *testing.T) {
		// Create login payload
		payload := map[string]string{
			"username": "loginuser",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(payload)

		// Create request
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Record response
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		// Check status code
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d; got %d", http.StatusOK, rec.Code)
		}

		// Check response contains token
		var response map[string]string
		json.Unmarshal(rec.Body.Bytes(), &response)
		if _, exists := response["token"]; !exists {
			t.Error("Response missing token")
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		// Create login payload with wrong password
		payload := map[string]string{
			"username": "loginuser",
			"password": "wrongpassword",
		}
		jsonData, _ := json.Marshal(payload)

		// Create request
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Record response
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		// Check status code
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		// Create login payload with non-existent username
		payload := map[string]string{
			"username": "nonexistent",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(payload)

		// Create request
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Record response
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		// Check status code
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}

func TestVerifyToken(t *testing.T) {
	server := setupTestServer()

	// Register a user and get token for verification tests
	registerPayload := map[string]string{
		"email":    "verify@example.com",
		"username": "verifyuser",
		"password": "password123",
	}
	jsonRegister, _ := json.Marshal(registerPayload)

	reqRegister := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonRegister))
	reqRegister.Header.Set("Content-Type", "application/json")
	recRegister := httptest.NewRecorder()
	server.ServeHTTP(recRegister, reqRegister)

	// Login to get token
	loginPayload := map[string]string{
		"username": "verifyuser",
		"password": "password123",
	}
	jsonLogin, _ := json.Marshal(loginPayload)

	reqLogin := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonLogin))
	reqLogin.Header.Set("Content-Type", "application/json")
	recLogin := httptest.NewRecorder()
	server.ServeHTTP(recLogin, reqLogin)

	var loginResponse map[string]string
	json.Unmarshal(recLogin.Body.Bytes(), &loginResponse)
	validToken := loginResponse["token"]

	t.Run("valid token", func(t *testing.T) {
		// Create request with valid token
		req := httptest.NewRequest("POST", "/verify", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		// Record response
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		// Check status code
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d; got %d", http.StatusOK, rec.Code)
		}

		// Check response contains user info
		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		if username, ok := response["username"]; !ok || username != "verifyuser" {
			t.Error("Response missing or incorrect username")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		// Create request with invalid token
		req := httptest.NewRequest("POST", "/verify", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")

		// Record response
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		// Check status code
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		// Create request without token
		req := httptest.NewRequest("POST", "/verify", nil)

		// Record response
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		// Check status code
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}
