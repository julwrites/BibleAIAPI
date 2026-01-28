---
id: FEATURES-20260127-081822-HWS
status: completed
title: Implement AI API V2 with Streaming and Preferences
priority: medium
created: 2026-01-27 08:18:22
category: features
dependencies: [INFRASTRUCTURE-20260127-081821-UIS]
type: task
---

# Implement AI API V2 with Streaming and Preferences

## Task Details
Update the API layer to expose the new functionality.

### Requirements
- [x] Update `ChatService` to handle `Stream` option and provider preference.
- [x] Update `QueryHandler` to parse `ai_provider` and `stream` from request.
- [x] Implement SSE (Server-Sent Events) logic in `QueryHandler` for streaming responses.
- [x] Implement JSON metadata wrapping for blocking responses.
- [x] Verify validation rules (e.g. Stream mutually exclusive with Schema).
