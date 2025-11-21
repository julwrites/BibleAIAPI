package util

import (
	"fmt"
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

	chapterAndVerse := strings.Split(chapterAndVerseStr, ":")
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
