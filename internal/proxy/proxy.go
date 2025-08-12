package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"azure-ai-proxy/internal/logging"
)

// Define custom context key types to avoid collisions
type contextKey string

const (
	requestBodyKey contextKey = "requestBody"
	pathKey        contextKey = "path"
	methodKey      contextKey = "method"
	startTimeKey   contextKey = "startTime"
)

// Server represents the proxy server
type Server struct {
	targetURL *url.URL
	proxy     *httputil.ReverseProxy
	logger    logging.Logger
	apiKey    string
}

// New creates a new proxy server
func New(targetURL *url.URL, logger logging.Logger, apiKey string) *Server {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	server := &Server{
		targetURL: targetURL,
		proxy:     proxy,
		logger:    logger,
		apiKey:    apiKey,
	}

	// Override the Director function to modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
	}

	// Create a custom transport that captures the response
	originalTransport := http.DefaultTransport
	proxy.Transport = &loggingTransport{
		transport: originalTransport,
		logger:    logger,
	}

	return server
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check API key if configured
	if s.apiKey != "" {
		clientKey := r.Header.Get("X-API-Key")
		if clientKey != s.apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Read and store the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Printf("Error closing request body: %v", err)
	}

	// Create a new reader with the same content
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse the request body to log it
	var requestBody interface{}
	if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
		log.Printf("Warning: Could not parse request body as JSON: %v", err)
		requestBody = string(bodyBytes)
	}

	// Store the request in context for the transport to access
	ctx := r.Context()
	ctx = context.WithValue(ctx, requestBodyKey, requestBody)
	ctx = context.WithValue(ctx, pathKey, r.URL.Path)
	ctx = context.WithValue(ctx, methodKey, r.Method)
	ctx = context.WithValue(ctx, startTimeKey, time.Now())

	// Serve the request with the modified context
	s.proxy.ServeHTTP(w, r.WithContext(ctx))
}

// Run starts the proxy server
func (s *Server) Run(listenAddr string) error {
	log.Printf("Starting proxy server on %s, forwarding to %s", listenAddr, s.targetURL)
	if s.apiKey != "" {
		log.Printf("API key authentication enabled")
	} else {
		log.Printf("Warning: API key authentication disabled, proxy is open to all requests")
	}
	return http.ListenAndServe(listenAddr, s)
}

// loggingTransport is a custom transport that logs responses
type loggingTransport struct {
	transport http.RoundTripper
	logger    logging.Logger
}

// RoundTrip implements the http.RoundTripper interface
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Get the stored request info from context
	requestBody := req.Context().Value(requestBodyKey)
	path := req.Context().Value(pathKey).(string)
	method := req.Context().Value(methodKey).(string)
	startTime := req.Context().Value(startTimeKey).(time.Time)

	// Make the original request
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Capture correlation ID from response headers
	correlationID := resp.Header.Get("apim-request-id")
	if correlationID == "" {
		// Fallback to other possible correlation headers
		correlationID = resp.Header.Get("x-ms-request-id")
		if correlationID == "" {
			correlationID = resp.Header.Get("x-ms-correlation-request-id")
		}
	}

	// Read the full response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := resp.Body.Close(); err != nil {
		log.Printf("Error closing response body: %v", err)
	}

	// Create a new response body for the client
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse the response for logging
	var responseBody interface{}
	if err := json.Unmarshal(bodyBytes, &responseBody); err != nil {
		// For non-JSON responses or streaming responses, process differently
		if strings.Contains(string(bodyBytes), "data: ") {
			responseBody = processStreamingResponse(string(bodyBytes))
		} else {
			responseBody = string(bodyBytes)
		}
	}

	// Log the entry with the parsed response
	t.logger.LogRequest(logging.Entry{
		Timestamp:     time.Now(),
		RequestBody:   requestBody,
		Response:      responseBody,
		Duration:      time.Since(startTime),
		Path:          path,
		Method:        method,
		CorrelationID: correlationID,
	})

	return resp, nil
}

// processStreamingResponse extracts the full content from a server-sent events stream
func processStreamingResponse(responseText string) map[string]interface{} {
	fullContent := ""
	lines := strings.Split(responseText, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				continue
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			// Extract content from choices
			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					// Check for content in delta (streaming format)
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							fullContent += content
						}
					}
					// Check for content in message (non-streaming format)
					if message, ok := choice["message"].(map[string]interface{}); ok {
						if content, ok := message["content"].(string); ok {
							fullContent += content
						}
					}
				}
			}
		}
	}

	// Return a simplified response object with the full content
	return map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": fullContent,
				},
			},
		},
	}
}
