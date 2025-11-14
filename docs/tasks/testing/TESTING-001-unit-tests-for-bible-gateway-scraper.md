# Task: Unit Tests for Bible Gateway Scraper

## Task Information
- **Task ID**: TESTING-001
- **Status**: pending
- **Priority**: high
- **Phase**: 2
- **Estimated Effort**: 1 day
- **Dependencies**: INFRASTRUCTURE-001

## Task Details

### Description
This task involves writing unit tests for the Bible Gateway scraper located in `internal/biblegateway/scraper.go`. The tests should cover both the `GetVerse` and `SearchWords` functions and ensure that the HTML parsing logic is robust and correct.

### Acceptance Criteria
- [ ] Unit tests for the `GetVerse` function are implemented.
- [ ] Unit tests for the `SearchWords` function are implemented.
- [ ] Tests use mock HTML responses to avoid actual network calls.
- [ ] Tests cover successful parsing scenarios.
- [ ] Tests cover scenarios where the HTML structure is unexpected or missing elements.
- [ ] All tests pass.

### Implementation Notes
- The `net/http/httptest` package can be used to create a mock HTTP server.
- Sample HTML responses should be stored in a `testdata` directory within `internal/biblegateway`.

---

*Created: 2025-11-14*
*Status: pending*
