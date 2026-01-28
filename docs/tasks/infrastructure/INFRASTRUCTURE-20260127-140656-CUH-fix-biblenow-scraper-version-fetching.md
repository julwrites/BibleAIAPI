---
id: INFRASTRUCTURE-20260127-140656-CUH
status: in_progress
title: Fix BibleNow Scraper Version Fetching
priority: medium
created: 2026-01-27 14:06:56
category: infrastructure
dependencies:
type: task
---

# Fix BibleNow Scraper Version Fetching

## Task Details
The current BibleNow scraper only fetches versions from the English page. However, BibleNow supports multiple languages, each with its own list of versions. Additionally, the `GetVerse` implementation relies on hardcoded English URL paths, which fail for non-English versions.

This task involves refactoring the scraper to support multi-language version fetching and updating `GetVerse` to handle localized URLs dynamically.

### Requirements
- [ ] Update `GetVersions` to iterate over all available languages on BibleNow.
- [ ] Parse versions from each language page and store the full relative path (e.g., `es/biblia/reina-valera-1909`) as the `Value`.
- [ ] Update `GetVerse` to use the stored `Value` path to locate the version.
- [ ] Implement logic in `GetVerse` to find the correct book link by index (since book names are localized) rather than constructing the URL string manually.
- [ ] Ensure unit tests cover the new multi-language fetching and dynamic verse URL construction.

## Implementation Plan
1.  **GetVersions**:
    *   Fetch `/en/bible` to discover language links.
    *   Concurrently fetch each language page (e.g., `/es`, `/af`).
    *   Extract version links from each language page.
2.  **GetVerse**:
    *   Fetch the version page (e.g., `/es/biblia/reina-valera-1909`).
    *   Parse all book links, filtering out "Old/New Testament" links.
    *   Use a standard list of books to map the requested book name to an index.
    *   Select the Nth book link from the parsed list.
    *   Append `/{chapter}` to the book URL to get the verse page.
