package middleware

import (
	"api-gateway/config"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret []byte
var authenticationOverride func(http.Handler) http.Handler

// InitAuth initializes authentication middleware with JWT secret
func InitAuth(cfg *config.Config) {
	jwtSecret = []byte(cfg.JWTSecret)
}

// Authentication middleware validates JWT tokens
func Authentication(next http.Handler) http.Handler {
	// Use override for testing if set
	if authenticationOverride != nil {
		return authenticationOverride(next)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Parse Bearer token
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), "user_id", claims["sub"])
		if roles, ok := claims["roles"]; ok {
			ctx = context.WithValue(ctx, "user_roles", roles)
		}

		// Call next handler with enhanced context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminOnly middleware restricts access to admin users
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roles := r.Context().Value("user_roles")

		// Check if user has admin role
		if roles == nil || !strings.Contains(roles.(string), "admin") {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SetAuthenticationOverride allows tests to override the Authentication middleware
func SetAuthenticationOverride(override func(http.Handler) http.Handler) func() {
	prev := authenticationOverride
	authenticationOverride = override
	// Return a function that restores the previous override
	return func() {
		authenticationOverride = prev
	}
}

// NewAuthenticationMiddleware creates a new authentication middleware
func NewAuthenticationMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Parse Bearer token
			tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
			fmt.Println("Token String:", tokenString)
			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return jwtSecret, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), "user_id", claims["sub"])
			if roles, ok := claims["roles"]; ok {
				ctx = context.WithValue(ctx, "user_roles", roles)
			}

			// Call next handler with enhanced context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
