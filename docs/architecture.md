# Azure AI Proxy - Architecture Documentation

## Overview

The Azure AI Proxy is a lightweight, transparent HTTP proxy designed to intercept and log API interactions between client applications and Azure OpenAI services. It sits between client applications and Azure OpenAI endpoints, transparently forwarding requests while capturing complete request/response pairs for logging and analysis.

## Purpose

This proxy serves several important purposes:

1. **Full Response Capture**: Intercepts streaming and non-streaming responses to ensure complete responses are logged
2. **Transparent Operation**: Acts as a pass-through proxy that doesn't modify the request/response content
3. **API Interaction Logging**: Records all API interactions for analysis, debugging, and monitoring
4. **Non-Opinionated Design**: Maintains the client's original request structure without imposing changes
5. **Simple Authentication**: Optional API key authentication to prevent unauthorized access to the proxy

## System Architecture

```
┌───────────┐       ┌──────────────┐       ┌───────────────┐
│           │       │              │       │               │
│  Client   │◄─────►│  Azure AI    │◄─────►│ Azure OpenAI  │
│Application│       │   Proxy      │       │   Service     │
│           │       │              │       │               │
└───────────┘       └──────────────┘       └───────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │             │
                    │   Log File  │
                    │             │
                    └─────────────┘
```

The proxy uses Go's built-in HTTP and reverse proxy capabilities to:

1. Receive client requests on a configurable listening address
2. Authenticate requests using an optional API key mechanism
3. Forward authenticated requests to the Azure OpenAI service
4. Capture and log the complete request/response interaction
5. Return the unmodified response to the client

## Key Components

### 1. Main Application (`main.go`)

Serves as the entry point for the application with the following responsibilities:
- Loading configuration from environment variables or defaults
- Initializing the logging system
- Creating and starting the proxy server

```go
func main() {
    // Load configuration
    cfg := config.NewDefaultConfig()
    
    // Create a logger
    logger, err := logging.NewFileLogger(cfg.LogFilePath)
    // ...
    
    // Create and start the proxy server
    server := proxy.New(targetURL, logger, cfg.APIKey)
    // ...
}
```

### 2. Configuration (`config/config.go`)

Provides configurable settings for the application through environment variables:
- `AZURE_OPENAI_ENDPOINT`: The Azure OpenAI service endpoint URL
- `LISTEN_ADDR`: The local address and port for the proxy to listen on
- `LOG_FILE_PATH`: The file path for request/response logs
- `PROXY_API_KEY`: Optional API key for authenticating requests to the proxy

Default values are provided when environment variables are not set.

### 3. Proxy Server (`internal/proxy/proxy.go`)

The core component of the system, implemented as a reverse proxy with the following architecture:

#### 3.1 Server Structure

```go
type Server struct {
    targetURL *url.URL
    proxy     *httputil.ReverseProxy
    logger    logging.Logger
}
```

The server uses Go's standard `httputil.ReverseProxy` with customized Director and Transport implementations.

#### 3.2 Request Handling Flow

1. **Authentication**:
   - If an API key is configured, the proxy checks for the `X-API-Key` header
   - Unauthorized requests are rejected with a 401 Unauthorized response
   - If no API key is configured, all requests are allowed

2. **Request Interception**:
   - The proxy intercepts incoming requests in the `ServeHTTP` method
   - The request body is read and stored for logging
   - The body is then repackaged for forwarding
   - Context values (path, method, etc.) are stored for logging

3. **Request Forwarding**:
   - The Director function preserves the Host header when forwarding
   - The request is sent to the Azure OpenAI service

4. **Response Processing**:
   - The custom `loggingTransport` intercepts the response
   - The complete response body is read and stored
   - A new response body is created for the client with the same content
   - The appropriate logging takes place

5. **Streaming Response Handling**:
   - The `processStreamingResponse` function handles server-sent events format
   - It extracts and concatenates content from streaming chunks
   - The full content is reconstructed for logging purposes

### 4. Logging System (`internal/logging/logger.go`)

The logging system provides a clean interface for recording request/response interactions:

#### 4.1 Logger Interface

```go
type Logger interface {
    LogRequest(entry Entry)
    Close()
}
```

#### 4.2 Entry Structure

Each log entry contains:
- Timestamp
- Request body (parsed JSON or string)
- Response body (parsed JSON or string)
- Duration (request processing time)
- Path (API endpoint path)
- Method (HTTP method)

#### 4.3 File Logger Implementation

The default implementation writes log entries as JSON to a file:
- Each entry is JSON formatted with indentation
- Entries are appended to the log file
- File operations are properly handled with error logging

## Data Flow

1. **Client Request**:
   - Client sends a request to the proxy at the configured listening address
   - Proxy reads, stores, and repackages the request
   - Request is forwarded to the Azure OpenAI service

2. **Service Response**:
   - Azure OpenAI service processes the request and sends a response
   - Proxy intercepts and reads the full response (whether streaming or not)
   - For streaming responses, content chunks are aggregated
   - The full response is logged

3. **Response Delivery**:
   - Original response is sent back to the client
   - Request/response pair is logged to the configured log file

## Technical Implementation Details

### Request Body Handling

```go
// Read and store the request body
bodyBytes, err := io.ReadAll(r.Body)
// ...
r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
```

The request body is captured and then repackaged to allow forwarding while preserving the content.

### Response Body Processing

```go
// Read the full response body
bodyBytes, err := io.ReadAll(resp.Body)
// ...
resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
```

The response body is fully read, which is crucial for streaming responses, and then repackaged for client delivery.

### Streaming Response Processing

```go
func processStreamingResponse(responseText string) map[string]interface{} {
    fullContent := ""
    lines := strings.Split(responseText, "\n")
    // Process each line of the stream
    // ...
    return map[string]interface{}{/* reconstructed response */}
}
```

The proxy handles Server-Sent Events (SSE) format used by the Azure OpenAI streaming API by:
1. Splitting the stream by line breaks
2. Processing each `data:` prefixed line
3. Parsing JSON chunks from each data line
4. Extracting and concatenating content from each chunk
5. Reconstructing a complete response object with the full content

## Performance Considerations

1. **Memory Usage**:
   - The proxy reads complete request/response bodies into memory
   - For large responses, this can increase memory usage

2. **Response Time**:
   - Reading the complete response before forwarding adds latency
   - This is a necessary trade-off to ensure complete logging

3. **Logging Volume**:
   - Large request/response pairs generate significant log data
   - Log rotation or management should be considered for production use

## Security Considerations

1. **API Keys and Secrets**:
   - The proxy logs complete request bodies, which may contain API keys
   - Sensitive information should be redacted in a production environment

2. **Data Exposure**:
   - Logged responses may contain sensitive or personal information
   - Appropriate log file permissions and handling are important

3. **Network Security**:
   - By default, the proxy listens on localhost only
   - Exposing the proxy to a wider network requires additional security measures

4. **Authentication**:
   - The proxy supports optional API key authentication via the `PROXY_API_KEY` environment variable
   - When enabled, clients must provide the same key in the `X-API-Key` header
   - If no API key is configured, authentication is disabled and all requests are allowed

## Authentication

The proxy implements a simple API key authentication mechanism:

1. **Configuration**:
   - Set the `PROXY_API_KEY` environment variable to enable authentication
   - If this variable is empty or not set, authentication is disabled

2. **Client Usage**:
   - Clients must include the API key in the `X-API-Key` header
   - Example: `X-API-Key: your-secret-key-here`

3. **Security Level**:
   - This is a basic authentication mechanism and should be used in conjunction with HTTPS in production
   - For internal or development use, it provides a simple access control mechanism

Example of setting the API key:
```bash
# Set the API key
export PROXY_API_KEY="your-secret-key-here"
# Run the proxy
./azure-ai-proxy
```

Example of using the API key in a client request:
```bash
curl -H "X-API-Key: your-secret-key-here" -H "Content-Type: application/json" -d '{"messages":[{"role":"user","content":"Say hello"}]}' http://localhost:8080/openai/deployments/gpt-4o/chat/completions
```

## Deployment Considerations

1. **Environment Configuration**:
   - Use environment variables to configure the proxy
   - Consider a more robust configuration system for production

2. **Log Management**:
   - Implement log rotation or a more sophisticated logging solution
   - Monitor log file size and storage requirements

3. **Scaling**:
   - The current implementation is designed for single-instance use
   - For higher throughput, consider a load-balanced deployment

## Future Enhancements

1. **Metrics Collection**:
   - Add response time, token usage, and other performance metrics
   - Implement prometheus or similar monitoring integration

2. **Request Filtering**:
   - Add capability to filter certain requests from logging
   - Implement path-based or content-based filtering

3. **Security Enhancements**:
   - Add authentication for the proxy itself
   - Implement redaction of sensitive information

4. **Advanced Logging**:
   - Support for different logging backends (e.g., database, cloud storage)
   - Structured logging for better analysis

5. **UI Dashboard**:
   - Add a web interface for viewing and analyzing logged interactions