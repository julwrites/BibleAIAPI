---
id: INFRASTRUCTURE-20260127-040640-ZAB
status: completed
title: Update versions.yaml Generation Tool
priority: medium
created: 2026-01-27 04:06:40
category: infrastructure
dependencies: [FEATURES-20260127-040617-RIS, FEATURES-20260127-040629-COZ]
type: task
---

# Update versions.yaml Generation Tool

Refactor cmd/update_versions to fetch versions from all registered providers (Gateway, Hub, Now). It should aggregate them into a union set and populate the 'providers' map in versions.yaml accordingly.
