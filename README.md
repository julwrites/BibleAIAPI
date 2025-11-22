# Bible API Service

A stateless microservice that provides a unified API for querying Bible verses, performing word searches, and interacting with LLMs for biblical context. Designed for serverless platforms like Google Cloud Run.

## Features

-   **Verse Retrieval**: Fetch verses by reference (e.g., `John 3:16`) with formatting preserved.
-   **Word Search**: Find verses by keywords.
-   **LLM Integration**: Ask questions or provide instructions (e.g., "Summarize", "Cross-reference") using various LLM providers (OpenAI, Gemini, DeepSeek).
-   **Smart Routing**: Routes queries based on whether they are verse lookups, word searches, or LLM prompts.
-   **Feature Flags**: Dynamic configuration via GitHub-hosted feature flags.

## API Reference

For detailed API documentation, see the [OpenAPI specification](./docs/api/openapi.yaml).

## Getting Started

### Prerequisites

-   [Go](https://golang.org/) (1.24+)
-   [Docker](https://www.docker.com/)

### Running Locally

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/julwrites/BibleAIAPI.git
    cd BibleAIAPI
    ```

2.  **Set up Environment Variables:**
    Create a `.env` file or export these variables.
    *   `API_KEY`: Arbitrary secret for local auth (e.g., `secret`).
    *   `LLM_PROVIDERS`: Comma-separated list of providers to enable (e.g., `openai,gemini`).
    *   `OPENAI_API_KEY`: Required if using OpenAI.
    *   `GEMINI_API_KEY`: Required if using Gemini.
    *   `DEEPSEEK_API_KEY`: Required if using DeepSeek.
    *   `GCP_PROJECT_ID`: (Optional) Required if you want to test Google Secret Manager integration; otherwise, it falls back to env vars.

3.  **Run the service:**
    ```bash
    go run cmd/server/main.go
    ```
    The server starts on port `8080`.

4.  **Test a Request:**
    ```bash
    curl -X POST http://localhost:8080/query \
      -H "X-API-KEY: secret" \
      -d '{"query": {"verses": ["John 3:16"]}}'
    ```
    *Note: For verse queries, the `context` object is not allowed.*

    **LLM Prompt Request:**
    ```bash
    curl -X POST http://localhost:8080/query \
      -H "X-API-KEY: secret" \
      -d '{"query": {"prompt": "Explain this verse"}, "context": {"verses": ["John 3:16"], "user": {"version": "ESV"}}}'
    ```

    **LLM Prompt Request with Word Search Context:**
    ```bash
    curl -X POST http://localhost:8080/query \
      -H "X-API-KEY: secret" \
      -d '{"query": {"prompt": "Summarize the verses containing this word"}, "context": {"words": ["Grace"], "user": {"version": "ESV"}}}'
    ```

### Testing Locally

You can use the provided script to automatically build, run, and verify the API service:

```bash
./scripts/test_api.sh
```

This script will:
1.  Build the server binary.
2.  Start the server in the background.
3.  Run `curl` requests for Verse lookup (Prose & Poetry) and Word search.
4.  Skip LLM tests if relevant API keys (`OPENAI_API_KEY`, etc.) are not found in the environment.

### Building with Docker

```bash
docker build -t bible-api-service .
docker run -p 8080:8080 --env-file .env bible-api-service
```

## Configuration

-   **Feature Flags**: Managed via `go-feature-flag`. The service retrieves flags from the [GitHub repository](https://github.com/julwrites/BibleAIAPI) by default, falling back to `configs/flags.yaml` locally.
-   **Secrets**: The service attempts to fetch secrets from Google Secret Manager. If unavailable (e.g., local dev), it falls back to environment variables.

## Project Structure

-   `cmd/server`: Main entry point.
-   `internal/biblegateway`: Scraper logic (Prose/Poetry handling).
-   `internal/llm`: LLM provider implementations.
-   `internal/handlers`: Request routing and processing.
-   `internal/secrets`: Secret management (GSM/Env).
-   `docs`: Architecture and deployment documentation.
