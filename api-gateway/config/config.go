package config

import (

    "os"
)

// Config holds application configuration
type Config struct {
    Port          string
    UserServiceURL string
    ChatServiceURL string
    CRMServiceURL  string
    JWTSecret     string
    AllowedOrigins string
    Environment   string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
    port := getEnv("PORT", "8080")
    userServiceURL := getEnv("USER_SERVICE_URL", "http://user-service:8082")
    chatServiceURL := getEnv("CHAT_SERVICE_URL", "http://chat-service:8091")
    crmServiceURL := getEnv("CRM_SERVICE_URL", "http://crm-service:8092")
    jwtSecret := getEnv("JWT_SECRET", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0LXVzZXItaWQiLCJyb2xlcyI6InVzZXIiLCJleHAiOjE5OTk5OTk5OTl9.iqfOjB4bABvLeZCYgGgRbWp4L9kFQBcdfSjVZZDKcqM")
    allowedOrigins := getEnv("ALLOWED_ORIGINS", "*")
    environment := getEnv("ENVIRONMENT", "development")
    
    return &Config{
        Port:           port,
        UserServiceURL: userServiceURL,
        ChatServiceURL: chatServiceURL,
        CRMServiceURL:  crmServiceURL,
        JWTSecret:      jwtSecret,
        AllowedOrigins: allowedOrigins,
        Environment:    environment,
    }
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
    value, exists := os.LookupEnv(key)
    if !exists {
        value = defaultValue
    }
    return value
}