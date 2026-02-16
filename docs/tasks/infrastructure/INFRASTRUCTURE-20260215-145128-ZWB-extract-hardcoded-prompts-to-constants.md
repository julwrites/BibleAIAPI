---
id: INFRASTRUCTURE-20260215-145128-ZWB
status: in_progress
title: Extract hardcoded prompts to constants
priority: medium
created: 2026-02-15 14:51:28
category: infrastructure
dependencies:
type: task
---

# Extract hardcoded prompts to constants

## Overview
Extract hardcoded prompt strings in `internal/chat/chat.go` to constants to improve maintainability and readability.

## Acceptance Criteria
- [x] Identify hardcoded prompt strings in `internal/chat/chat.go`.
- [x] Define constants for these strings.
- [x] Replace hardcoded strings with constants in `Process` and `formatHistory` functions.
- [x] Ensure behavior remains unchanged.

## Progress
- [x] Explored `internal/chat/chat.go` to find hardcoded strings.
- [x] Defined constants at the top of the file.
- [x] Refactored `Process` and `formatHistory`.
- [x] Verified changes using `grep` and `gofmt`.
