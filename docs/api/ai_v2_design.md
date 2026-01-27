# AI API V2 Design

This document outlines the design for the next iteration of the AI Query API, introducing support for streaming responses, specific provider requests, and structured output formalization.

## 1. Request Schema

The `QueryRequest` object is extended to support user preferences and execution options.

### Structure

```json
{
  "query": {
    "prompt": "Explain John 3:16"
  },
  "context": {
    "user": {
      "version": "ESV",
      "ai_provider": "openai"  // Optional: Preferred provider ID (e.g., "openai", "deepseek")
    },
    "schema": "..." // Optional: JSON Schema for structured output.
  },
  "options": {
    "stream": true // Optional: Enable streaming response (default: false)
  }
}
```

### Validation Rules
1.  **Mutual Exclusion**: `options.stream` MUST be `false` (or omitted) if `context.schema` is present. Structured output is only supported in blocking mode.
2.  **Default Schema Bypass**: If `options.stream` is `true`, the system MUST NOT inject a default schema (e.g., `oquery_response`).
3.  **Provider Preference**: `ai_provider` is preferential. If the requested provider is unavailable, the system falls back to the default order, but includes the actual provider used in the response metadata.

## 2. Response Schema

### Blocking Response (Standard JSON)

When `stream` is false, the API returns a JSON object wrapping the result and metadata.

```json
{
  "data": {
    // The result.
    // If context.schema was provided (or defaulted), this is the structured JSON object.
    "text": "For God so loved the world...",
    "references": [...]
  },
  "meta": {
    "ai_provider": "openai", // The provider that fulfilled the request
    "model": "gpt-4o"       // Optional: specific model used
  }
}
```

### Streaming Response (Server-Sent Events)

When `stream` is true, the API responds with `Content-Type: text/event-stream`.

**Events:**

*   `meta`: Sent once (preferably first) containing metadata.
    ```json
    data: {"ai_provider": "openai"}
    ```
*   `chunk`: Sent repeatedly with content deltas.
    ```json
    data: {"delta": "For God "}
    ```
    ```json
    data: {"delta": "so loved "}
    ```
*   `error`: Sent if an error occurs mid-stream.
    ```json
    data: {"error": "Connection lost"}
    ```
*   `done`: Sent to indicate end of stream (optional if connection closes, but good practice).
    ```json
    data: "[DONE]"
    ```

## 3. Internal Interface Changes

To support these features, the internal `LLMClient` interface and implementations must be updated.

### `LLMClient` Interface

```go
type LLMClient interface {
    // Query performs a blocking request.
    // Returns the raw response string, the provider name, and error.
    Query(ctx context.Context, prompt string, schema string) (response string, providerName string, err error)

    // Stream performs a streaming request.
    // Returns a channel for chunks, the provider name, and error (immediate).
    // Note: The channel closes when the stream is done.
    Stream(ctx context.Context, prompt string) (<-chan string, string, error)

    // Name returns the identifier of the provider (e.g., "openai").
    Name() string
}
```

### `FallbackClient` Logic

The `FallbackClient` will be refactored to:
1.  Maintain a map of initialized clients (e.g., `map[string]provider.LLMClient`).
2.  Maintain a default priority list.
3.  Accept a "preferred provider" argument (via context or method argument).
4.  Attempt the preferred provider first.
5.  If it fails, iterate through the remaining providers in the configured priority order.

### Chat Service

The `ChatService` must be updated to:
1.  Accept `Options` in the request.
2.  If `Stream` is true, invoke `client.Stream`.
3.  If `Stream` is false, invoke `client.Query`.
4.  Return a result type that can encapsulate either a parsed JSON object (blocking) or a channel (streaming), along with metadata.
