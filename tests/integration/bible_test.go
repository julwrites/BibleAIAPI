//go:build integration

package integration

import (
	"bible-api-service/internal/bible/providers/biblegateway"
	"strings"
	"testing"
)

func TestBibleGatewayIntegration(t *testing.T) {
	scraper := biblegateway.NewScraper()

	// Test GetVerse
	verse, err := scraper.GetVerse("John", "3", "16", "ESV")
	if err != nil {
		t.Fatalf("Failed to fetch John 3:16: %v", err)
	}

	if !strings.Contains(verse, "For God so loved the world") {
		t.Errorf("Expected verse to contain 'For God so loved the world', got: %s", verse)
	}

	// Test SearchWords
	results, err := scraper.SearchWords("Jesus wept", "ESV")
	if err != nil {
		t.Fatalf("Failed to search 'Jesus wept': %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one search result for 'Jesus wept', got 0")
	}
}
