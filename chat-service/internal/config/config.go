// internal/config/config.go
package config

import (
	"os"
	"strconv"
	"log"
)

type Config struct {
	Port        string
	DBHost      string
	DBUser      string
	DBPassword  string
	DBName      string
	RabbitMQURL string
	UseAI	  bool
	OpenAIKey   string
	OpenAIURL   string
	OpenAIModel string
	OpenAITemperature float64
	OpenAIChatModel string
	OpenAIChatTemperature float64
	OpenAIChatMaxTokens int
	OpenAIChatTopP float64
	OpenAIChatFrequencyPenalty float64
	OpenAIChatPresencePenalty float64
	OpenAIChatStop []string
	OpenAIChatUser string
	OpenAIChatSystem string
	OpenAIChatAssistant string
	OpenAIChatUserName string
	OpenAIChatAssistantName string
	OpenAIChatUserRole string
	OpenAIChatAssistantRole string
	OpenAIChatUserRoleName string
	OpenAIChatAssistantRoleName string
	OpenAIChatUserRoleDescription string
	OpenAIChatAssistantRoleDescription string
	OpenAIChatUserRoleDescriptionName string
	OpenAIChatAssistantRoleDescriptionName string
}

func LoadConfig() Config {
	return Config{
		Port:        getEnv("PORT", "8081"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "crm"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		UseAI:      getEnv("USE_AI", "true") == "true",
		OpenAIKey:   getEnv("OPENAI_KEY", ""),
		OpenAITemperature: mustParseFloat(getEnv("OPENAI_TEMPERATURE", "0.7")),
		OpenAIChatModel: getEnv("OPENAI_CHAT_MODEL", "gpt-3.5-turbo"),
		OpenAIChatTemperature: mustParseFloat(getEnv("OPENAI_CHAT_TEMPERATURE", "0.7")),
		OpenAIChatMaxTokens: mustParseInt(getEnv("OPENAI_CHAT_MAX_TOKENS", "100")),
		OpenAIChatTopP: mustParseFloat(getEnv("OPENAI_CHAT_TOP_P", "1.0")),
		OpenAIChatFrequencyPenalty: mustParseFloat(getEnv("OPENAI_CHAT_FREQUENCY_PENALTY", "0.0")),
		OpenAIChatPresencePenalty: mustParseFloat(getEnv("OPENAI_CHAT_PRESENCE_PENALTY", "0.0")),
		OpenAIChatStop: []string{getEnv("OPENAI_CHAT_STOP", "\n")},
		OpenAIChatUser: getEnv("OPENAI_CHAT_USER", "user"),
		OpenAIChatSystem: getEnv("OPENAI_CHAT_SYSTEM", "system"),
		OpenAIChatAssistant: getEnv("OPENAI_CHAT_ASSISTANT", "assistant"),
		OpenAIChatUserName: getEnv("OPENAI_CHAT_USER_NAME", "User"),
		OpenAIChatAssistantName: getEnv("OPENAI_CHAT_ASSISTANT_NAME", "Assistant"),
		OpenAIChatUserRole: getEnv("OPENAI_CHAT_USER_ROLE", "user"),
		OpenAIChatAssistantRole: getEnv("OPENAI_CHAT_ASSISTANT_ROLE", "assistant"),
		OpenAIChatUserRoleName: getEnv("OPENAI_CHAT_USER_ROLE_NAME", "User"),
		OpenAIChatAssistantRoleName: getEnv("OPENAI_CHAT_ASSISTANT_ROLE_NAME", "Assistant"),
		OpenAIChatUserRoleDescription: getEnv("OPENAI_CHAT_USER_ROLE_DESCRIPTION", "User"),
		OpenAIChatAssistantRoleDescription: getEnv("OPENAI_CHAT_ASSISTANT_ROLE_DESCRIPTION", "Assistant"),
		OpenAIChatUserRoleDescriptionName: getEnv("OPENAI_CHAT_USER_ROLE_DESCRIPTION_NAME", "User"),
		OpenAIChatAssistantRoleDescriptionName: getEnv("OPENAI_CHAT_ASSISTANT_ROLE_DESCRIPTION_NAME", "Assistant"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func mustParseFloat(val string) float64 {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Fatalf("Failed to parse float from %q: %v", val, err)
	}
	return f
}

func mustParseInt(val string) int {
	i, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("Failed to parse int from %q: %v", val, err)
	}
	return i
}

