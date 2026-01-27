---
id: FOUNDATION-20260127-040600-ETB
status: pending
title: Refactor VersionManager for Multi-Provider Support
priority: high
created: 2026-01-27 04:06:00
category: foundation
dependencies:
type: task
---

# Refactor VersionManager for Multi-Provider Support

Refactor the VersionManager to support dynamic provider lookup based on a priority list (Gateway > Hub > Now). Currently, it relies on a single provider env var. It should check versions.yaml mappings to select the best provider.
