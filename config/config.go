package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	AzureOpenAIEndpoint string
	ListenAddr          string
	LogFilePath         string
	APIKey              string
}

// NewDefaultConfig returns a config with values from environment variables or defaults
func NewDefaultConfig() *Config {
	return &Config{
		AzureOpenAIEndpoint: getEnvOrDefault("AZURE_OPENAI_ENDPOINT", "your-deployment.openai.azure.com/"),
		ListenAddr:          getEnvOrDefault("LISTEN_ADDR", ":8080"),
		LogFilePath:         getEnvOrDefault("LOG_FILE_PATH", "openai_proxy.json"),
		APIKey:              getEnvOrDefault("PROXY_API_KEY", ""),
	}
}

// getEnvOrDefault returns the value of the environment variable or the default if not set
func getEnvOrDefault(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}
