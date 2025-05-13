package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// WebSocketProxy creates a specialized reverse proxy for WebSocket connections.
// targetBaseURLStr: The base URL of the upstream WebSocket service (e.g., "http://chat-service:8080").
// targetWSPath: The specific path on the upstream service for WebSocket connections (e.g., "/ws").
func WebSocketProxy(targetBaseURLStr string, targetWSPath string) http.Handler {
	targetURL, err := url.Parse(targetBaseURLStr)
	if err != nil {
		log.Fatalf("WebSocketProxy: Error parsing target base URL '%s': %v", targetBaseURLStr, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// Call the original director first. It sets req.URL.Scheme, req.URL.Host,
		// and potentially req.URL.Path if the targetURL had a path.
		originalDirector(req)

		// Override the path to the specific WebSocket path on the target.
		req.URL.Path = targetWSPath
		// Ensure the Host header matches the target's host.
		// This is important for services that use the Host header for routing or validation.
		req.Host = targetURL.Host

		// httputil.ReverseProxy automatically handles critical WebSocket headers:
		// - 'Connection'
		// - 'Upgrade'
		// It also copies most other headers.

		// Log the details of the request being proxied.
		log.Printf("WebSocketProxy Director: Forwarding to -> Scheme: %s, Host: %s, Path: %s, RawQuery: %s",
			req.URL.Scheme, req.URL.Host, req.URL.Path, req.URL.RawQuery)
		log.Printf("WebSocketProxy Director: Original Request URI: %s, Method: %s", req.RequestURI, req.Method)
	}

	// ErrorHandler to log errors from the proxy.
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("WebSocketProxy ErrorHandler: Proxy error for %s %s: %v", r.Method, r.RequestURI, err)
		// Avoid writing an HTTP error if the connection has already been hijacked for WebSocket.
		// Check if headers have been written (a common check, though not foolproof for hijacked conns).
		// For WebSockets, if an error occurs after the handshake, sending an HTTP error is usually not possible.
		// The primary purpose here is logging.
		// if !w.(http.Flusher).Flushed() { // This check is problematic as Flusher might not be available or indicative
		// http.Error(w, "Proxy Error", http.StatusBadGateway)
		// }
	}

	// ModifyResponse can be useful for debugging the handshake response from the backend.
	proxy.ModifyResponse = func(resp *http.Response) error {
		log.Printf("WebSocketProxy ModifyResponse: Backend response status for %s: %d", resp.Request.URL.String(), resp.StatusCode)
		if resp.StatusCode == http.StatusSwitchingProtocols {
			log.Printf("WebSocketProxy ModifyResponse: Backend successfully switched protocols.")
		}
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("WebSocketProxy Handler: Received request for %s (Upgrade header: '%s')", r.URL.Path, r.Header.Get("Upgrade"))
		proxy.ServeHTTP(w, r)
	})
}
