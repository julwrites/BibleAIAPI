# Agent Instructions

This document provides instructions for AI agents working on this codebase.

## Overview

This is a Go-based API service designed to be a stateless microservice. It interacts with the Bible Gateway website and Large Language Models (LLMs) to provide various functionalities related to Bible verses.

## Key Technologies

- **Go**: The primary programming language. The service uses the standard `net/http` library for the web server.
- **Docker**: Used for containerization. The `Dockerfile` is a multi-stage build to create a small, optimized image.
- **go-feature-flag**: Used for managing instructions and prompts. The configuration is in `configs/flags.yaml`.
- **langchaingo**: Used for interacting with LLMs. The LLM client is designed to be modular to support multiple providers.

## Project Structure

- `cmd/server`: The main entry point for the application.
- `internal`: Contains the core application logic. This is where you'll find the handlers, services, and clients for interacting with external services.
- `pkg`: For shared code that can be used across the application.
- `configs`: All configuration files, including the feature flags for `go-feature-flag`.
- `docs`: Documentation, including the OpenAPI specification in `docs/api`.

## Development Workflow

1.  **Understand the API**: Before making changes, review the OpenAPI specification in `docs/api/openapi.yaml` to understand the API contract.
2.  **Modify the code**: Make your changes to the Go source files. Remember to follow Go best practices.
3.  **Update dependencies**: If you add or change dependencies, use `go get` and `go mod tidy` to update the `go.mod` and `go.sum` files.
4.  **Test your changes**: Ensure that your changes are covered by tests.
5.  **Update documentation**: If you change the API, update the OpenAPI specification and any other relevant documentation.

## Running Tests

To run the tests, use the following command:

```bash
go test ./...
```

## Code Coverage

This project enforces a high standard of code coverage. All new code should be accompanied by tests, and the overall code coverage should not drop. The CI/CD pipeline will fail if the code coverage drops.

To run the tests with coverage, use the following command:

```bash
go test -v -cover ./...
```

## Code Style

This project follows standard Go formatting. Use `gofmt` to format your code before committing.
