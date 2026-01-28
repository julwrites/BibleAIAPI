---
id: FEATURE-003
status: completed
title: Refine LLM Schema Validation and Error Handling
priority: medium
created: 2025-12-11 06:17:08
category: features
type: task
---

# Refine LLM Schema Validation and Error Handling

### Description
This task involves refining the LLM schema validation and error handling. The current implementation relies on the LLM to always return a valid JSON object that conforms to the schema. This task will involve adding a validation step to the `ChatService` to ensure that the LLM's response is valid before returning it to the client. It will also involve implementing a more robust error handling strategy for when the LLM's response is invalid.

### Acceptance Criteria
- [x] A validation library (e.g., `gojsonschema`) is added to the project to validate the LLM's response against the JSON schema.
- [x] The `ChatService.Process` method is updated to validate the LLM's response against the provided schema.
- [x] A clear and informative error message is returned when the LLM's response is invalid, detailing the schema violations.
- [x] The feature is covered by unit tests in `internal/chat/chat_test.go`.
