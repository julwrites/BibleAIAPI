---
id: FOUNDATION-20260127-040609-YJP
status: pending
title: Dynamic Provider Selection in QueryHandler
priority: high
created: 2026-01-27 04:06:09
category: foundation
dependencies: [FOUNDATION-20260127-040600-ETB]
type: task
---

# Dynamic Provider Selection in QueryHandler

Update QueryHandler to select the Bible provider dynamically per request using the VersionManager. Remove the global BIBLE_PROVIDER env var dependency for request handling.
