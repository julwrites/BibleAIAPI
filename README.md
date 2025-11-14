# Bible API Service

This is a stateless microservice that provides a simple API for querying and interacting with Bible verses. It is designed to be hosted on serverless platforms like Google Cloud Run.

## Features

- **Verse Retrieval**: Fetch Bible verses by reference (e.g., `John 3:16`).
- **Word/Phrase Search**: Find verses related to specific words or phrases (e.g., `love`, `grace`).
- **Open-Ended Queries**: Ask questions and receive answers from an LLM, complete with biblical references (e.g., `Who was Moses?`).
- **Instruction-Based Processing**: Use instructions (e.g., `summarize`, `cross-reference`) to perform complex actions on a given context.
- **Customizable**: User preferences like Bible version can be specified in the request.

## API

For detailed information about the API, please see the [OpenAPI specification](./docs/api/openapi.yaml).

## Getting Started

### Prerequisites

- [Go](https://golang.org/)
- [Docker](https://www.docker.com/)

### Running Locally

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd <repository-name>
    ```

2.  **Set up environment variables:**
    Create a `.env` file in the root of the project and add the following variables:
    ```
    API_KEY=<your-secret-api-key>
    OPENAI_API_KEY=<your-openai-api-key>
    GEMINI_API_KEY=<your-gemini-api-key>
    ```

3.  **Run the service:**
    ```bash
    go run cmd/server/main.go
    ```
    The server will start on port `8080`.

### Building with Docker

To build the Docker image, run:
```bash
docker build -t bible-api-service .
```

To run the Docker container:
```bash
docker run -p 8080:8080 -e API_KEY=<your-secret-api-key> -e OPENAI_API_KEY=<your-openai-api-key> -e GEMINI_API_KEY=<your-gemini-api-key> bible-api-service
```

## Configuration

This service uses a feature flag system to manage instructions and prompts. The configuration for these flags is stored in `configs/flags.yaml`.

## Project Structure

- `cmd/server`: Main application entry point.
- `configs`: Feature flag and other service configurations.
- `docs`: Project documentation, including the OpenAPI specification.
- `internal`: Core application logic, including handlers, clients, and services.
- `pkg`: Shared libraries and utilities.
