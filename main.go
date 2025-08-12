package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"azure-ai-proxy/config"
	"azure-ai-proxy/internal/logging"
	"azure-ai-proxy/internal/proxy"
)

func run() error {
	// Load configuration
	cfg := config.NewDefaultConfig()

	// Create a logger
	logger, err := logging.NewFileLogger(cfg.LogFilePath)
	if err != nil {
		return fmt.Errorf("failed to create logger: %v", err)
	}
	defer logger.Close()

	// Parse the target URL
	targetURL, err := url.Parse(cfg.AzureOpenAIEndpoint)
	if err != nil {
		return fmt.Errorf("failed to parse target URL: %v", err)
	}

	// Create and start the proxy server
	server := proxy.New(targetURL, logger, cfg.APIKey)
	log.Printf("Logging requests and responses to %s", cfg.LogFilePath)
	if err := server.Run(cfg.ListenAddr); err != nil {
		return fmt.Errorf("server error: %v", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
