---
id: INFRASTRUCTURE-004
status: completed
title: Implement Unified Version Management
priority: medium
created: 2026-01-26 18:00:00
category: infrastructure
dependencies: INFRASTRUCTURE-20260126-152022-TMY
type: task
---

# Implement Unified Version Management

## Task Information
- **Dependencies**: INFRASTRUCTURE-20260126-152022-TMY (Refactor to Bible Service Provider Pattern)

## Task Details
Redesign the `configs/versions.yaml` schema and the `cmd/update_versions` tool to support unified version codes mapped to provider-specific identifiers. This allows the application to handle multiple Bible providers (Bible Gateway, BibleHub, BibleNow) with consistent versioning.

### Acceptance Criteria
- [x] `configs/versions.yaml` schema is updated to include provider mappings.
- [x] `cmd/update_versions` is updated to generate the new schema.
- [x] `cmd/update_versions` populates `biblegateway` mappings from the scraper.
- [x] `cmd/update_versions` generates default mappings for `biblehub` (lowercase) and `biblenow` (slugs).
- [x] Verify that `versions.yaml` can be successfully generated.

## Implementation Plan
1.  Define new structs for the version config.
2.  Update `cmd/update_versions/main.go` to use the new structs.
3.  Add logic to generate provider-specific codes.
4.  Run `go run cmd/update_versions/main.go` to regenerate the file.
