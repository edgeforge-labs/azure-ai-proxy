# Azure AI Proxy

A Go application that proxies requests to Azure OpenAI services to capture diagnostics about prompts and their responses. It sits between your client applications and Azure's AI endpoints, logging complete request/response cycles including streaming responses.

## Features

- Transparently proxies requests to Azure OpenAI services
- Captures complete request and response data, including streaming responses
- Minimal impact on request/response flow
- Optional API key authentication
- JSON-formatted logging to file

## Architecture

For detailed information about the architecture of this proxy, please see [architecture.md](docs/architecture.md).

## Installation

### Using Pre-built Binary

Download the latest version from the releases page.

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/azure-ai-proxy.git
cd azure-ai-proxy

# Build the application
go build .
```

### Build & run the container

```bash
docker build -t azure-ai-proxy:local --build-arg IMAGE_NAME=azure-ai-proxy .
docker run azure-ai-proxy:local
```

## Configuration

Azure AI Proxy can be configured using environment variables:

| Environment Variable  | Description                                 | Default Value                     |
| --------------------- | ------------------------------------------- | --------------------------------- |
| AZURE_OPENAI_ENDPOINT | URL of the Azure OpenAI service endpoint    | your-deployment.openai.azure.com/ |
| LISTEN_ADDR           | Address and port for the proxy to listen on | localhost:8080                    |
| LOG_FILE_PATH         | File path for request/response logs         | openai_proxy.json                 |
| PROXY_API_KEY         | API key for proxy authentication (optional) | (none)                            |

## Authentication

When `PROXY_API_KEY` is set for extra hardening, the proxy requires clients to include the key in the `X-API-Key` header, effectively requiring 2 API keys.

## Runing the proxy

**Windows**
```batch
set AZURE_OPENAI_ENDPOINT=https://your-endpoint.openai.azure.com/
set PROXY_API_KEY=your-secret-key

azure-ai-proxy.exe
```

**Linux**
```sh
export AZURE_OPENAI_ENDPOINT=https://your-endpoint.openai.azure.com/
export PROXY_API_KEY=your-secret-key

./azure-ai-proxy
```

## Using the Proxy

After running the proxy, you can use it to send requests to Azure OpenAI services:

**bash**
```bash
export PROXY_API_KEY="yourapikey"
export AZUREAI_API_KEY="yourapikey"
curl -X POST \
  "http://localhost:8080/openai/deployments/gpt-4o/chat/completions?api-version=2025-01-01-preview" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $PROXY_API_KEY" \
  -H "api-key: $AZUREAI_API_KEY" \
  -d '{"messages":[{"role":"user","content":"Say hello"}],"max_tokens":1000}'
```

**powershell**
```powershell
$env:PROXY_API_KEY="yourapikey"
$env:AZUREAI_API_KEY="yourapikey"
curl -X POST `
  "http://localhost:8080/openai/deployments/gpt-4o/chat/completions?api-version=2025-01-01-preview" `
  -H "Content-Type: application/json" `
  -H "X-API-Key: $($env:PROXY_API_KEY)" `
  -H "api-key: $($env:AZUREAI_API_KEY)" `
  -d '{"messages":[{"role":"user","content":"Say hello"}],"max_tokens":1000}'
```
