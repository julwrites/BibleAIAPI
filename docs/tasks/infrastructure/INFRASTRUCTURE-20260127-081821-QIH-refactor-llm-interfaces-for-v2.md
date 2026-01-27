---
id: INFRASTRUCTURE-20260127-081821-QIH
status: pending
title: Refactor LLM Interfaces for V2
priority: medium
created: 2026-01-27 08:18:21
category: infrastructure
dependencies:
type: task
---

# Refactor LLM Interfaces for V2

## Task Details
Update `internal/llm/provider/provider.go` to support V2 features.

### Requirements
- [ ] Add `Stream(ctx context.Context, prompt string) (<-chan string, string, error)` to `LLMClient`.
- [ ] Add `Name() string` to `LLMClient`.
- [ ] Update `Query` signature to `(string, string, error)` to return the provider name.
- [ ] Reference `docs/api/ai_v2_design.md` for exact specifications.
