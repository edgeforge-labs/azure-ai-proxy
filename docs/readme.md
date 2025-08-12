# Azure AI Proxy documentation

Here we'll provide some additional documentation about this proxy implementation.

## Client Libraries

You can use any OpenAI client library by simply changing the base URL to point to your proxy:

#### Python Example

```python
import openai

client = openai.AzureOpenAI(
    azure_endpoint="http://localhost:8080",  # Your proxy address
    api_key="your-azure-api-key",            # Your Azure API key
    api_version="2023-05-15",
    headers={
        "X-API-Key": "your-proxy-api-key"    # Your proxy authentication key
    }
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Say hello"}],
    max_tokens=1000
)
print(response.choices[0].message.content)
```

#### JavaScript Example

```javascript
import { OpenAIClient } from "@azure/openai";
import { AzureKeyCredential } from "@azure/core-auth";

const client = new OpenAIClient(
  "http://localhost:8080",                // Your proxy address
  new AzureKeyCredential("your-azure-api-key"),
  {
    headers: { "X-API-Key": "your-proxy-api-key" }  // Your proxy authentication key
  }
);

const response = await client.getChatCompletions(
  "gpt-4o",
  [{ role: "user", content: "Say hello" }],
  { maxTokens: 1000 }
);
console.log(response.choices[0].message.content);
```

## Log Format

Logs are stored in JSON format with the following structure:

```json
{"Timestamp":"2025-07-08T21:46:39.1604573+02:00","RequestBody":{"max_tokens":1000,"messages":[{"content":"Say hello","role":"user"}]},"Response":{"choices":[{"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"protected_material_code":{"detected":false,"filtered":false},"protected_material_text":{"detected":false,"filtered":false},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"finish_reason":"stop","index":0,"logprobs":null,"message":{"annotations":[],"content":"Hello! 😊 How can I assist you today?","refusal":null,"role":"assistant"}}],"created":1752003998,"id":"chatcmpl-Br8Yw6JYDAZyrQEmSeqGk05GVuSE9","model":"gpt-4o-2024-11-20","object":"chat.completion","prompt_filter_results":[{"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"jailbreak":{"detected":false,"filtered":false},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"prompt_index":0}],"system_fingerprint":"fp_ee1d74bde0","usage":{"completion_tokens":11,"completion_tokens_details":{"accepted_prediction_tokens":0,"audio_tokens":0,"reasoning_tokens":0,"rejected_prediction_tokens":0},"prompt_tokens":9,"prompt_tokens_details":{"audio_tokens":0,"cached_tokens":0},"total_tokens":20}},"Duration":507298300,"Path":"/openai/deployments/gpt-4o/chat/completions","Method":"POST"}
```