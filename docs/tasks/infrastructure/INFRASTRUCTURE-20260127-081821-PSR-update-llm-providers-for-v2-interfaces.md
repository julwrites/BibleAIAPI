---
id: INFRASTRUCTURE-20260127-081821-PSR
status: pending
title: Update LLM Providers for V2 Interfaces
priority: medium
created: 2026-01-27 08:18:21
category: infrastructure
dependencies: [INFRASTRUCTURE-20260127-081821-QIH]
type: task
---

# Update LLM Providers for V2 Interfaces

## Task Details
Update all LLM provider implementations to satisfy the new `LLMClient` interface.

### Requirements
- [ ] Update `openai` provider (implement `Stream`, `Name`, update `Query`).
- [ ] Update `deepseek` provider.
- [ ] Update `gemini` provider.
- [ ] Update `openapicustom` provider.
- [ ] Verify `Stream` implementations use `langchaingo` streaming capabilities where available.
