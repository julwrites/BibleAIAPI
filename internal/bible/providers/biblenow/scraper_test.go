package biblenow

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScraper_GetVerse(t *testing.T) {
	// Mock HTML response
	mockHTML := `
<!DOCTYPE html>
<html>
<body>
	<div class="verse list-group chapter-content">
		<a href="#" class="list-group-item">
			<p class="verse"><span>1</span> In the beginning God created the heaven and the earth.</p>
		</a>
		<a href="#" class="list-group-item">
			<p class="verse"><span>2</span> And the earth was without form, and void; and darkness was upon the face of the deep. And the Spirit of God moved upon the face of the waters.</p>
		</a>
	</div>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL format
		// /en/bible/king-james-version/old-testament/genesis/1
		expectedPath := "/en/bible/king-james-version/old-testament/genesis/1"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}
		w.Write([]byte(mockHTML))
	}))
	defer server.Close()

	scraper := NewScraper()
	scraper.baseURL = server.URL // Override baseURL with mock server URL

	// Test case 1: Single verse
	verse, err := scraper.GetVerse("Genesis", "1", "1", "KJV")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedVerse := "In the beginning God created the heaven and the earth."
	if verse != expectedVerse {
		t.Errorf("expected '%s', got '%s'", expectedVerse, verse)
	}

	// Test case 2: Verse range
	verseRange, err := scraper.GetVerse("Genesis", "1", "1-2", "KJV")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedRange := "In the beginning God created the heaven and the earth. And the earth was without form, and void; and darkness was upon the face of the deep. And the Spirit of God moved upon the face of the waters."
	if verseRange != expectedRange {
		t.Errorf("expected '%s', got '%s'", expectedRange, verseRange)
	}
}

func TestScraper_SearchWords(t *testing.T) {
	scraper := NewScraper()
	_, err := scraper.SearchWords("love", "KJV")
	if err == nil {
		t.Error("expected error for SearchWords, got nil")
	}
}

func TestGetVersionSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"KJV", "king-james-version"},
		{"ESV", "english-standard-version"},
		{"NIV", "new-international-version"},
		{"NKJV", "new-king-james-version"},
		{"ASV", "american-standard-version"},
		{"NASB", "new-american-standard-bible"},
		{"NLT", "new-living-translation"},
		{"WEB", "web"}, // lowercase default
		{"Unknown Version", "unknown-version"}, // slugify
		{"UPPERCASE", "uppercase"},
		{"CamelCase", "camelcase"},
		{"Space Version", "space-version"},
	}

	for _, tt := range tests {
		got := GetVersionSlug(tt.input)
		if got != tt.expected {
			t.Errorf("GetVersionSlug(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
