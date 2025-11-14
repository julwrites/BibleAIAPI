# Architecture

This document provides an overview of the system architecture for the Bible API Service.

## System Overview

The Bible API Service is a stateless microservice designed for serverless environments like Google Cloud Run. It is built in Go and containerized with Docker.

The service has two primary modes of operation:
1.  **Direct Query**: When no instruction is provided, the service directly queries Bible Gateway for verses or search results.
2.  **Instruction-Based Query**: When an instruction is provided, the service uses a Large Language Model (LLM) to process the query and context, based on a predefined prompt and response schema.

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

-   **Handlers**: Contain the main business logic for processing API requests.
-   **Bible Gateway Client**: A client for scraping verse data and search results from `classic.biblegateway.com`.
-   **LLM Client**: A modular client for interacting with LLMs (e.g., OpenAI, Gemini) via `langchaingo`.
-   **Feature Flag Service**: Integrates with `go-feature-flag` to retrieve prompts and schemas from `configs/flags.yaml`.

### 3. Configuration (`configs`)

-   `flags.yaml`: Stores the configuration for the feature flags, including prompts and schemas for different instructions.

### 4. Containerization (`Dockerfile`)

-   A multi-stage `Dockerfile` is used to build a small, optimized container image for the service.

## Data Flow

### Without Instruction

1.  Client sends a request with a Bible verse, word, or open-ended query.
2.  The API server authenticates the request.
3.  The handler calls the Bible Gateway client to fetch the requested information.
4.  The response is formatted and returned to the client.

### With Instruction

1.  Client sends a request with an instruction, query, and context.
2.  The API server authenticates the request.
3.  The handler retrieves the corresponding prompt and schema from the feature flag service.
4.  The handler calls the LLM client with the prompt, schema, and context.
5.  The LLM returns a structured response.
6.  The response is returned to the client.
