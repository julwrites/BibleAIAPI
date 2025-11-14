# Task: Core API Functionality

## Task Information
- **Task ID**: FEATURE-001
- **Status**: completed
- **Priority**: critical
- **Phase**: 1
- **Estimated Effort**: 1 day
- **Actual Effort**: 1 day
- **Completed**: 2025-11-14
- **Dependencies**: FOUNDATION-001, INFRASTRUCTURE-001

## Task Details

### Description
This task covers the implementation of the core API functionality for the `/query` endpoint. This includes handling both direct queries (verse, word, and open-ended) and instruction-based queries.

### Acceptance Criteria
- [x] The `/query` endpoint is implemented.
- [x] The handler can distinguish between requests with and without instructions.
- [x] Direct queries for verses are handled by the Bible Gateway scraper.
- [x] Direct queries for words/phrases are handled by the Bible Gateway scraper.
- [x] Direct open-ended queries are handled by the LLM client.
- [x] Instruction-based queries are handled by the LLM client with structured output.
- [x] Middleware for API key authentication and logging is implemented and applied to the endpoint.

## Implementation Status

### Completed Work
- ✅ Implemented the main `QueryHandler` in `internal/handlers/query.go`.
- ✅ Added logic to parse the request body and determine the query type.
- ✅ Integrated the Bible Gateway scraper for direct verse and word queries.
- ✅ Implemented a modular LLM client in `internal/llm/` with an OpenAI implementation.
- ✅ Implemented the logic for handling open-ended queries by calling the LLM client.
- ✅ Implemented the logic for handling instruction-based queries, including collating context, formatting the prompt, and calling the LLM client with a schema to enforce structured output.
- ✅ Implemented and applied middleware for API key authentication (`internal/middleware/auth.go`) and logging (`internal/middleware/logging.go`).
- ✅ Implemented a standardized JSON error response structure (`internal/util/error.go`).

---

*Created: 2025-11-14*
*Last updated: 2025-11-14*
*Status: completed - Core API functionality is implemented.*
