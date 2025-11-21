# Agent Instructions

This document provides instructions for AI agents and developers working on the Bible API Service codebase.

## Overview

The Bible API Service is a Go-based stateless microservice designed for serverless deployment (Google Cloud Run). It acts as a bridge between clients, the Bible Gateway website (via scraping), and Large Language Models (LLMs).

## Key Technologies & Architecture

-   **Go**: Primary language (Go 1.24+).
-   **Docker**: Multi-stage builds for optimized container images.
-   **Scraping**: Uses `goquery` to scrape `classic.biblegateway.com`. Logic distinguishes between **Prose** and **Poetry** to preserve formatting.
-   **LLM Clients**: Modular architecture in `internal/llm`. Supports multiple providers (OpenAI, Gemini, DeepSeek, Custom).
    -   **Fallback**: The `FallbackClient` (`internal/llm/client.go`) attempts providers in a priority list defined by `LLM_PROVIDERS` env var.
-   **Feature Flags**: Managed via `go-feature-flag`.
    -   **Source**: Primarily loaded from the GitHub repository (`julwrites/BibleAIAPI`).
    -   **Fallback**: Falls back to a local file (`configs/flags.yaml`) if GitHub retrieval fails.
-   **Secrets**: Managed via `internal/secrets`.
    -   **Source**: Attempts to fetch from Google Secret Manager first.
    -   **Fallback**: Falls back to environment variables (useful for local development).

## Project Structure

-   `cmd/server`: Application entry point.
-   `internal/biblegateway`: Scraper logic. **Critical**: Must handle `div.poetry` for poetry preservation.
-   `internal/handlers`: HTTP handlers. `QueryHandler` routes requests based on payload (`instruction` > `chat_prompt` > `verses`).
-   `internal/llm`: Modular LLM provider implementations.
-   `internal/middleware`: Auth middleware (`X-API-KEY`) and logging.
-   `internal/secrets`: Secret retrieval logic.
-   `configs`: Local fallback configuration files.
-   `docs`: Detailed documentation (API, Architecture, Deployment).

## Scraper Implementation Details

The scraper (`internal/biblegateway/scraper.go`) is sensitive to HTML structure.
-   **Allowed Tags**: `h1`, `h2`, `h3`, `h4`, `p`, `span`, `i`, `br`, `sup` (for verse numbers).
-   **Sanitization**: Aggressive cleanup for prose; structure preservation for poetry.
-   **Non-breaking Spaces**: Must be converted to regular spaces (`\u00a0` -> ` `).

## Development Workflow

1.  **Understand the Goal**: Read `README.md` and `docs/` first.
2.  **Verify State**: Always check `go.mod` and existing code before assuming dependencies or logic.
3.  **Testing**:
    -   **Unit Tests**: Write table-driven tests. Use `go test ./...`.
    -   **Coverage**: Strict ratcheting coverage is enforced. Coverage must typically exceed 80%. Use `go tool cover` to check.
    -   **Scraper Tests**: Use realistic HTML snippets in tests, not mock strings.
    -   **Integration Tests**: Run `main` in a goroutine and make real HTTP requests (mocking external calls if needed).
4.  **Documentation**: Update `docs/` and `README.md` if logic changes.
5.  **Pre-Commit**: Ensure all tests pass and coverage is sufficient before submitting.

## Deployment

-   Refer to `docs/deployment.md` for detailed instructions.
-   **Secrets**: `API_KEY`, `GCP_PROJECT_ID`, and relevant LLM keys (`OPENAI_API_KEY`, etc.) must be set.
