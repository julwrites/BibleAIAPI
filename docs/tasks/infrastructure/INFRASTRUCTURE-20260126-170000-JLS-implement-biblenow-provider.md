---
id: INFRASTRUCTURE-20260126-170000-JLS
status: completed
title: Implement BibleNow Provider
priority: medium
created: 2026-01-26 17:00:00
category: infrastructure
dependencies: INFRASTRUCTURE-20260126-152022-TMY
type: task
---

# Implement BibleNow Provider

## Task Information
- **Dependencies**: INFRASTRUCTURE-20260126-152022-TMY (Refactor to Bible Service Provider Pattern)

## Task Details
Implement a new Bible provider for `BibleNow.net`. This will allow the application to fetch verses from BibleNow.

### Acceptance Criteria
- [x] `internal/bible/providers/biblenow` package is created.
- [x] `Scraper` struct is implemented with `GetVerse` and `SearchWords` methods.
- [x] `GetVerse` correctly scrapes verses from BibleNow URLs (handling testament logic).
- [x] `SearchWords` is implemented (returning not supported if unavailable).
- [x] Unit tests are implemented for the new provider.
- [x] The provider adheres to the `internal/bible.Provider` interface.
- [x] The provider is integrated into the application (via `BIBLE_PROVIDER` env var).

## Implementation Status
### Completed Work
- Task created.
- Implemented `Scraper` in `internal/bible/providers/biblenow/scraper.go`.
- Added unit tests in `internal/bible/providers/biblenow/scraper_test.go`.
- Integrated into `internal/handlers/query.go`.
- Updated `docs/deployment.md`.

### Blockers
None.
