# Architecture

This document provides an overview of the system architecture for the Bible API Service.

## System Overview

The Bible API Service is a stateless microservice designed for serverless environments like Google Cloud Run. It is built in Go and containerized with Docker.

## Components

### 1. API Server (`cmd/server`)

-   **Language**: Go
-   **Framework**: Standard `net/http`
-   **Responsibilities**:
    -   Handles incoming HTTP requests.
    -   Implements API key authentication.
    -   Provides structured logging.
    -   Routes requests to the appropriate handlers in the `internal` package.

### 2. Core Logic (`internal`)

-   **Handlers**: Contain the main business logic. The `QueryHandler` determines the type of request.
-   **Bible Gateway Client**: Scrapes verse data from `classic.biblegateway.com`. It intelligently parses HTML, distinguishing between prose and poetry to preserve formatting.
-   **LLM Client**: A modular client for interacting with LLMs. It supports multiple providers (OpenAI, Gemini, DeepSeek, etc.) via a common interface and includes a fallback mechanism.
-   **Chat Service**: Orchestrates the interaction between the API handler and the LLM client, managing context and schemas.
-   **Feature Flag Service**: Integrates with `go-feature-flag`. It attempts to retrieve configuration from the GitHub repository (`julwrites/BibleAIAPI`) and falls back to a local file (`configs/flags.yaml`) if needed.
-   **Secret Service**: Abstraction for secret retrieval. It prioritizes Google Secret Manager but falls back to environment variables for local development.

### 3. Configuration (`configs`)

-   `flags.yaml`: Local fallback configuration for feature flags.

## Data Flow

### Request Routing Logic
The `QueryHandler` inspects the request payload (`query` object) and routes it based on which field is present. The API enforces that exactly one of the following is present:

1.  **Prompt**: If `query.prompt` is present -> LLM Prompt Flow.
2.  **Verses**: If `query.verses` is present -> Verse Retrieval (Scraper).
3.  **Words**: If `query.words` is present -> Word Search (Scraper).

### Verse Retrieval Flow
1.  Client sends a request with verse references (`query.verses`).
2.  Handler calls `BibleGatewayClient.GetVerse`.
3.  Client scrapes `classic.biblegateway.com`, parses HTML (handling poetry/prose), and sanitizes output.
4.  Formatted HTML is returned.

### Word Search Flow
1.  Client sends a request with words (`query.words`).
2.  Handler calls `BibleGatewayClient.SearchWords`.
3.  Client scrapes search results from `classic.biblegateway.com`.
4.  List of results (verse reference and text snippet) is returned.

### LLM Prompt Flow
1.  Client sends a request with a prompt (`query.prompt`) and optional context (`context` object).
    -   **Context**: Can include `verses` (for specific verses), `words` (for word search results), `history` (chat history), and `schema` (JSON schema for response).
2.  Handler constructs a `ChatRequest` using the prompt and context.
    -   If `context.schema` is not provided, a default "Open Query" schema is used.
3.  Handler calls the `ChatService`.
4.  `ChatService` invokes the `LLMClient`.
5.  `LLMClient` attempts to call the configured providers (defined in `LLM_PROVIDERS`) in order.
6.  If a provider fails, the next one is tried (Fallback).
7.  Structured response is returned to the client.
