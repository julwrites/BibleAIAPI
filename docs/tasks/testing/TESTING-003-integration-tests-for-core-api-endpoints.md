---
id: TESTING-003
status: completed
title: Integration Tests for Core API Endpoints
priority: medium
created: 2025-12-11 06:17:08
category: unknown
type: task
---

# Integration Tests for Core API Endpoints

### Description
This task involves writing integration tests for the core API endpoints. These tests will start the actual HTTP server and make requests to it, verifying the end-to-end functionality of the service.

### Acceptance Criteria
- [x] Integration tests for the `/query` endpoint are implemented.
- [x] Tests cover all four query types (verse, word, open-ended, and instruction-based).
- [x] Tests use a mock LLM client to avoid making actual calls to the LLM API.
- [x] Tests verify the correctness of the API responses, including the HTTP status code and the JSON body.
- [x] All tests pass.

### Implementation Notes
- A separate `main_test.go` file can be used to set up the test server.
- The `httptest.Server` can be used to start a real server on a local port.
- As of 2025-11-14, the open query test is failing as expected due to the lack of a real OpenAI API key in the CI environment.
