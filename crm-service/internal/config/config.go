// filepath: crm-service/internal/config/config.go
package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Port           int
	DBHost         string
	DBUser         string
	DBPassword     string
	DBName         string
	RabbitMQURL    string
	ChatServiceURL string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	port, err := strconv.Atoi(getEnv("PORT", "8092"))
	if err != nil {
		log.Printf("Invalid PORT, using default: 8092")
		port = 8092
	}

	return &Config{
		Port:           port,
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "crm"),
		RabbitMQURL:    getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		ChatServiceURL: getEnv("CHAT_SERVICE_URL", "http://chat-service:8091"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
