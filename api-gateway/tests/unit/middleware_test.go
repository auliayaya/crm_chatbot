package unit

import (
	"api-gateway/config"
	"api-gateway/middleware"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationMiddleware(t *testing.T) {
	// Set up a test configuration
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	middleware.InitAuth(cfg)

	// Create a simple handler for testing
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if user_id made it to the context
		userID := r.Context().Value("user_id")
		if userID != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"user_id":"` + userID.(string) + `"}`))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// Create the middleware handler
	handler := middleware.Authentication(nextHandler)

	t.Run("Token tests with real middleware", func(t *testing.T) {
		t.Run("Valid JWT token should pass", func(t *testing.T) {
			// Create a test token
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub":   "test-user-id",
				"roles": "user",
				"exp":   time.Now().Add(time.Hour).Unix(),
			})
			tokenString, err := token.SignedString([]byte("test-secret"))
			assert.NoError(t, err)

			// Create request with token
			req := httptest.NewRequest("GET", "/api/resource", nil)
			req.Header.Set("Authorization", "Bearer "+tokenString)

			// Test middleware
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusOK, recorder.Code)
		})

		t.Run("Missing token should fail", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/resource", nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		})

		t.Run("Invalid token should fail", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/resource", nil)
			req.Header.Set("Authorization", "Bearer invalid-token")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		})

		t.Run("Token debug info", func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub":   "test-user-id",
				"roles": "user",
				"exp":   time.Now().Add(time.Hour).Unix(),
			})

			tokenString, err := token.SignedString([]byte("test-secret"))
			assert.NoError(t, err)
			t.Logf("Valid token for testing: %s", tokenString)

			// Try to parse the hardcoded token with our secret
			hardcodedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0LXVzZXItaWQiLCJyb2xlcyI6InVzZXIiLCJleHAiOjE5OTk5OTk5OTl9.iqfOjB4bABvLeZCYgGgRbWp4L9kFQBcdfSjVZZDKcqM"
			_, err = jwt.Parse(hardcodedToken, func(t *jwt.Token) (interface{}, error) {
				return []byte("test-secret"), nil
			})
			if err != nil {
				t.Logf("Error parsing hardcoded token with test-secret: %v", err)
			} else {
				t.Log("Hardcoded token is valid with test-secret")
			}
		})
	})

	t.Run("Tests with mock middleware", func(t *testing.T) {
		// Set up the override for this test
		restore := middleware.SetAuthenticationOverride(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Always add the user_id to context for testing
				ctx := context.WithValue(r.Context(), "user_id", "test-user-id")
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		// Ensure we restore at the end of the test
		defer restore()

		// Your tests with the mocked middleware
		req := httptest.NewRequest("GET", "/api/resource", nil)
		recorder := httptest.NewRecorder()
		middleware.Authentication(nextHandler).ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestAdminOnlyMiddleware(t *testing.T) {
	// Create a simple handler for testing
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create the middleware handler
	handler := middleware.AdminOnly(nextHandler)

	t.Run("User with admin role should pass", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/admin/resource", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, "user_roles", "admin")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("User without admin role should fail", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/admin/resource", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, "user_roles", "user")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusForbidden, recorder.Code)
	})
}
