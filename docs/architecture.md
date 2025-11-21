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
-   **Feature Flag Service**: Integrates with `go-feature-flag`. It attempts to retrieve configuration from the GitHub repository (`julwrites/BibleAIAPI`) and falls back to a local file (`configs/flags.yaml`) if needed.
-   **Secret Service**: Abstraction for secret retrieval. It prioritizes Google Secret Manager but falls back to environment variables for local development.

### 3. Configuration (`configs`)

-   `flags.yaml`: Local fallback configuration for feature flags.

## Data Flow

### Request Routing Logic
The `QueryHandler` inspects the request payload and routes it based on the following precedence:
1.  **Instruction**: If `context.instruction` is present -> LLM Instruction Flow.
2.  **Chat Prompt**: If `query.chat_prompt` is present -> LLM Chat Flow.
3.  **Verses**: If `query.verses` is present -> Verse Retrieval (Scraper).
4.  **Words**: If `query.words` is present -> Word Search (Scraper).
5.  **Open Query**: If `query.oquery` is present -> Open Query (LLM).

### Verse Retrieval Flow
1.  Client sends a request with verse references.
2.  Handler calls `BibleGatewayClient`.
3.  Client scrapes `classic.biblegateway.com`, parses HTML (handling poetry/prose), and sanitizes output.
4.  Formatted HTML is returned.

### LLM Instruction/Chat Flow
1.  Client sends a request with instruction/prompt and context.
2.  Handler retrieves the relevant prompt template/schema from the Feature Flag Service (GitHub/Local).
3.  Handler calls the `LLMClient`.
4.  `LLMClient` attempts to call the configured providers (defined in `LLM_PROVIDERS`) in order.
5.  If a provider fails, the next one is tried (Fallback).
6.  Structured response is returned to the client.
