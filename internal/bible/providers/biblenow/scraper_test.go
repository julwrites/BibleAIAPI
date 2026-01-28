package biblenow

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestScraper_GetVerse(t *testing.T) {
	// Mock HTML for Chapter Page
	mockChapterHTML := `
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

	// Mock HTML for Version Page (listing books)
	mockVersionHTML := `
<!DOCTYPE html>
<html>
<body>
	<a href="%s/en/bible/king-james-version/old-testament">Old Testament</a>
	<a href="%s/en/bible/king-james-version/old-testament/genesis">Genesis</a>
	<a href="%s/en/bible/king-james-version/old-testament/exodus">Exodus</a>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baseURL := "http://" + r.Host

		// Normalize paths to ignore host
		if r.URL.Path == "/en/bible/king-james-version" {
			w.Write([]byte(fmt.Sprintf(mockVersionHTML, baseURL, baseURL, baseURL)))
			return
		}

		if r.URL.Path == "/en/bible/king-james-version/old-testament/genesis/1" {
			w.Write([]byte(mockChapterHTML))
			return
		}

		t.Errorf("Unexpected request: %s", r.URL.Path)
		http.NotFound(w, r)
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

func TestScraper_GetVerse_Spanish(t *testing.T) {
	// Mock HTML for Spanish Chapter Page
	mockChapterHTML := `
<!DOCTYPE html>
<html>
<body>
	<div class="verse list-group chapter-content">
		<a href="#" class="list-group-item">
			<p class="verse"><span>1</span> En el principio creó Dios los cielos y la tierra.</p>
		</a>
	</div>
</body>
</html>
`
	// Mock HTML for Version Page
	mockVersionHTML := `
<!DOCTYPE html>
<html>
<body>
	<a href="%s/es/biblia/reina-valera-1909/antiguo-testamento">Antiguo Testamento</a>
	<a href="%s/es/biblia/reina-valera-1909/antiguo-testamento/genesis">Génesis</a>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baseURL := "http://" + r.Host

		if r.URL.Path == "/es/biblia/reina-valera-1909" {
			w.Write([]byte(fmt.Sprintf(mockVersionHTML, baseURL, baseURL)))
			return
		}

		if r.URL.Path == "/es/biblia/reina-valera-1909/antiguo-testamento/genesis/1" {
			w.Write([]byte(mockChapterHTML))
			return
		}

		t.Errorf("Unexpected request: %s", r.URL.Path)
		http.NotFound(w, r)
	}))
	defer server.Close()

	scraper := NewScraper()
	scraper.baseURL = server.URL

	// Pass full path as version
	verse, err := scraper.GetVerse("Genesis", "1", "1", "es/biblia/reina-valera-1909")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedVerse := "En el principio creó Dios los cielos y la tierra."
	if verse != expectedVerse {
		t.Errorf("expected '%s', got '%s'", expectedVerse, verse)
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

func TestScraper_GetVerse_WithExtraLinks(t *testing.T) {
	// Mock HTML for Chapter Page
	mockChapterHTML := `
<!DOCTYPE html>
<html>
<body>
	<div class="verse list-group chapter-content">
		<a href="#" class="list-group-item">
			<p class="verse"><span>1</span> In the beginning God created the heaven and the earth.</p>
		</a>
	</div>
</body>
</html>
`

	// Mock HTML for Version Page (listing books)
	// Inject an "Introduction" link that mimics the structure of a book link
	mockVersionHTML := `
<!DOCTYPE html>
<html>
<body>
	<a href="%s/en/bible/king-james-version/old-testament">Old Testament</a>
    <!-- Extra link that looks like a book but isn't -->
	<a href="%s/en/bible/king-james-version/old-testament/introduction">Introduction</a>
	<a href="%s/en/bible/king-james-version/old-testament/genesis">Genesis</a>
	<a href="%s/en/bible/king-james-version/old-testament/exodus">Exodus</a>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baseURL := "http://" + r.Host

		// Normalize paths to ignore host
		if r.URL.Path == "/en/bible/king-james-version" {
			w.Write([]byte(fmt.Sprintf(mockVersionHTML, baseURL, baseURL, baseURL, baseURL)))
			return
		}

		if r.URL.Path == "/en/bible/king-james-version/old-testament/genesis/1" {
			w.Write([]byte(mockChapterHTML))
			return
		}

		if strings.Contains(r.URL.Path, "introduction") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Introduction content"))
			return
		}

		t.Errorf("Unexpected request: %s", r.URL.Path)
		http.NotFound(w, r)
	}))
	defer server.Close()

	scraper := NewScraper()
	scraper.baseURL = server.URL

	// We expect Genesis (index 0) to be found.
	// If "Introduction" is NOT filtered out, it will be index 0, and Genesis will be index 1.
	_, err := scraper.GetVerse("Genesis", "1", "1", "KJV")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
