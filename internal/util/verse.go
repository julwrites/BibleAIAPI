package util

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseVerseReference parses a verse reference string into book, chapter, and verse components.
// It handles book names with spaces (e.g., "1 John").
// The expected format is "Book Name Chapter:Verse" or "Book Name Chapter".
func ParseVerseReference(ref string) (string, string, string, error) {
	lastSpaceIndex := strings.LastIndex(ref, " ")
	if lastSpaceIndex == -1 {
		return "", "", "", fmt.Errorf("invalid verse reference format: missing space between book and chapter")
	}

	book := ref[:lastSpaceIndex]
	chapterAndVerseStr := ref[lastSpaceIndex+1:]

	chapterAndVerse := strings.SplitN(chapterAndVerseStr, ":", 2)
	chapter := chapterAndVerse[0]
	var verse string

	if len(chapterAndVerse) > 1 {
		verse = chapterAndVerse[1]
	}

	// Basic validation to ensure chapter is not empty
	if chapter == "" {
		return "", "", "", fmt.Errorf("invalid verse reference format: missing chapter")
	}

	return book, chapter, verse, nil
}

// ParsedVerseRange contains the components of a verse range.
type ParsedVerseRange struct {
	StartVerse     int
	EndVerse       int
	EndChapter     int
	IsCrossChapter bool
}

// ParseVerseRange parses a verse string which can be a single number (16),
// a range within a chapter (16-20), or a cross-chapter range (12-2:4).
func ParseVerseRange(verseStr string) (ParsedVerseRange, error) {
	var result ParsedVerseRange

	if verseStr == "" {
		return result, nil
	}

	// Check for cross-chapter range (contains "-" and then ":")
	// Format: "12-2:4" -> Start: 12, EndChapter: 2, EndVerse: 4
	if strings.Contains(verseStr, "-") {
		parts := strings.Split(verseStr, "-")
		if len(parts) != 2 {
			return result, fmt.Errorf("invalid range format")
		}

		startStr := strings.TrimSpace(parts[0])
		endStr := strings.TrimSpace(parts[1])

		// Parse Start Verse
		start, err := strconv.Atoi(startStr)
		if err != nil {
			return result, fmt.Errorf("invalid start verse: %v", err)
		}
		result.StartVerse = start

		// Check if end part has chapter
		if strings.Contains(endStr, ":") {
			// Cross-chapter
			endParts := strings.Split(endStr, ":")
			if len(endParts) != 2 {
				return result, fmt.Errorf("invalid end reference in cross-chapter range")
			}

			endChap, err := strconv.Atoi(strings.TrimSpace(endParts[0]))
			if err != nil {
				return ParsedVerseRange{}, fmt.Errorf("invalid end chapter: %v", err)
			}
			endV, err := strconv.Atoi(strings.TrimSpace(endParts[1]))
			if err != nil {
				return ParsedVerseRange{}, fmt.Errorf("invalid end verse: %v", err)
			}

			result.EndChapter = endChap
			result.EndVerse = endV
			result.IsCrossChapter = true
		} else {
			// Same chapter range
			end, err := strconv.Atoi(endStr)
			if err != nil {
				return ParsedVerseRange{}, fmt.Errorf("invalid end verse: %v", err)
			}
			result.EndVerse = end
			result.IsCrossChapter = false
		}
	} else {
		// Single verse
		val, err := strconv.Atoi(strings.TrimSpace(verseStr))
		if err != nil {
			return ParsedVerseRange{}, fmt.Errorf("invalid verse number: %v", err)
		}
		result.StartVerse = val
		result.EndVerse = val
		result.IsCrossChapter = false
	}

	return result, nil
}
