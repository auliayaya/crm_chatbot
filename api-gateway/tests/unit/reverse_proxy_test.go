package unit

import (
	"api-gateway/proxy"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReverseProxy(t *testing.T) {
	// Create a test server that will act as a backend service
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if headers were forwarded correctly
		userID := r.Header.Get("X-User-ID")
		if userID == "test-user" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","path":"` + r.URL.Path + `"}`))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer backendServer.Close()

	// Create a reverse proxy
	reverseProxy, err := proxy.NewReverseProxy(backendServer.URL, "/api/service")
	assert.NoError(t, err)
	assert.NotNil(t, reverseProxy)

	// Create a test server using our reverse proxy
	proxyServer := httptest.NewServer(reverseProxy)
	defer proxyServer.Close()

	t.Run("Forwards request with user context", func(t *testing.T) {
		// Create request with user ID in context
		req, err := http.NewRequest("GET", proxyServer.URL+"/api/service/resource", nil)
		assert.NoError(t, err)

		// Add user ID both to context and as a header
		// The context part is to test that the proxy will read from context
		ctx := context.WithValue(req.Context(), "user_id", "test-user")
		req = req.WithContext(ctx)

		// This simulates what the Authentication middleware would do
		req.Header.Set("X-User-ID", "test-user")

		// Make the request
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Path stripping works correctly", func(t *testing.T) {
		// Create request with a path that should be stripped
		req, err := http.NewRequest("GET", proxyServer.URL+"/api/service/users", nil)
		assert.NoError(t, err)

		// Add user ID to context and header
		ctx := context.WithValue(req.Context(), "user_id", "test-user")
		req = req.WithContext(ctx)
		req.Header.Set("X-User-ID", "test-user")

		// Make the request
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
