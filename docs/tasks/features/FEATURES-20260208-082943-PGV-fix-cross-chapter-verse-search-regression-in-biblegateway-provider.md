---
id: FEATURES-20260208-082943-PGV
status: review_requested
title: Fix cross-chapter verse search regression in BibleGateway provider
priority: medium
created: 2026-02-08 08:29:43
category: features
dependencies: 
type: task
---

# Fix cross-chapter verse search regression in BibleGateway provider

Regression test fails for cross-chapter verse search with BibleGateway provider. Need to implement cross-chapter support in biblegateway scraper.

## Implementation

- Updated `internal/bible/providers/biblegateway/scraper.go`:
  - Added import for `util` package
  - Modified `GetVerse` to parse verse ranges using `util.ParseVerseRange`
  - Added cross-chapter handling: iterate through chapters for cross-chapter ranges
  - Added helper method `getVersesFromChapter` to fetch whole chapter and extract verses within range (plain text)
  - Preserve existing behavior for single-chapter ranges (HTML output)
- Updated `internal/bible/providers/biblegateway/scraper_test.go`:
  - Modified `TestGetVerse_CrossChapterQuery` to expect two chapter requests instead of one cross-chapter range request
- All existing tests pass; unit test for cross-chapter passes

## Changes

- `internal/bible/providers/biblegateway/scraper.go`
- `internal/bible/providers/biblegateway/scraper_test.go`

## Branch

`fix/cross-chapter-biblegateway`

## PR

https://github.com/julwrites/BibleAIAPI/pull/new/fix/cross-chapter-biblegateway
