---
id: INFRASTRUCTURE-20260127-081821-UIS
status: completed
title: Implement Provider Preference in Fallback Client
priority: medium
created: 2026-01-27 08:18:21
category: infrastructure
dependencies: [INFRASTRUCTURE-20260127-081821-PSR]
type: task
---

# Implement Provider Preference in Fallback Client

## Task Details
Refactor `internal/llm/client.go` (`FallbackClient`) to support dynamic provider selection.

### Requirements
- [ ] Store a map of initialized clients in `FallbackClient` in addition to the list.
- [ ] Update `Query` and `Stream` to accept an optional "preferred provider" name.
- [ ] Logic: Try preferred -> Try others in order.
- [ ] Return the name of the provider that actually succeeded.
