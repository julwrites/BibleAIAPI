---
id: INFRASTRUCTURE-20251212-015124-HYC
status: verified
title: Fix feature flag rate limiting
priority: medium
created: 2025-12-12 01:51:24
category: infrastructure
dependencies:
type: task
---

# Fix feature flag rate limiting

## Task Information
- **Dependencies**: None

## Task Details
The application is currently retrieving feature flags very frequently (default 10s), which is causing rate limiting issues in the Cloud Run environment when interacting with the GitHub API.

The goal is to:
1.  Increase the default polling interval to a safer value (e.g., 5 minutes).
2.  Make the polling interval configurable via an environment variable (`FEATURE_FLAG_POLLING_INTERVAL`) to allow for adjustment without code changes.

### Acceptance Criteria
- [x] Default polling interval is increased to 5 minutes (300 seconds).
- [x] `FEATURE_FLAG_POLLING_INTERVAL` environment variable is read and used if present.
- [x] Application handles invalid duration strings in the environment variable gracefully (falls back to default).
- [x] Unit tests verify the parsing logic and default behavior.

## Implementation Status
### Completed Work
- ✅ Implemented `getPollingInterval` helper in `internal/config/featureflags.go`.
- ✅ Updated `InitFeatureFlags` to use the dynamic polling interval.
- ✅ Added unit tests for configuration logic in `internal/config/featureflags_test.go`.
- ✅ Verified tests pass with `go test ./internal/config/...`.

### Blockers
None.
