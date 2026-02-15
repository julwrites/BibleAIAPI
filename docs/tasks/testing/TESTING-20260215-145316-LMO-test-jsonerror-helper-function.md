---
id: TESTING-20260215-145316-LMO
status: completed
title: Test JSONError helper function
priority: medium
created: 2026-02-15 14:53:16
category: testing
dependencies:
type: task
---

# Test JSONError helper function

## Description
The `JSONError` helper function in `internal/util/error.go` needs comprehensive unit tests to ensure it correctly sets HTTP headers, status codes, and JSON response bodies.

## Acceptance Criteria
- [x] Verify `Content-Type` header is set to `application/json`.
- [x] Verify correct HTTP status code is returned.
- [x] Verify JSON response body matches the `ErrorResponse` structure.
- [x] Handle edge cases such as empty messages and special characters.
- [x] Ensure 100% statement coverage for `JSONError`.

## Sub-tasks
- [x] Analyze existing tests in `internal/util/error_test.go`.
- [x] Implement additional test cases.
- [x] Add header validation.
- [x] Verify coverage and passing tests.
