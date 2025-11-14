# API Design

This document outlines the design principles for the Bible API Service.

## Guiding Principles

- **Stateless**: The API is stateless. Clients are responsible for providing all necessary context in each request.
- **JSON**: The API uses JSON for all request and response bodies.
- **Clear and Consistent**: The API is designed to be clear, consistent, and predictable.
- **Secure**: All endpoints are protected by API key authentication.
- **Formal Specification**: The API is formally defined using the OpenAPI 3.0 specification.

## API Specification

The OpenAPI specification can be found in [`docs/api/openapi.yaml`](./api/openapi.yaml). This is the single source of truth for the API's contract.

## Authentication

Authentication is handled via an API key. The client must provide the API key in the `X-API-KEY` header of each request.

## Error Handling

The API uses a standardized JSON error structure for all error responses:

```json
{
  "error": {
    "code": <integer>,
    "message": "<string>"
  }
}
```

Common HTTP status codes are used to indicate the success or failure of a request.
- `200 OK`: The request was successful.
- `400 Bad Request`: The request was malformed (e.g., invalid JSON).
- `401 Unauthorized`: The API key is missing or invalid.
- `404 Not Found`: The requested resource was not found.
- `500 Internal Server Error`: An unexpected error occurred on the server.
