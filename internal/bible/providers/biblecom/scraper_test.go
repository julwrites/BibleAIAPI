package biblecom

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVerse(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check URL
		if r.URL.Path == "/bible/111/JHN.3.1" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
				<html>
				<body>
					<div class="chapter">
						<span data-usfm="JHN.3.1" class="verse v1">For God so loved the world</span>
						<span data-usfm="JHN.3.2" class="verse v2">that he gave his one and only Son</span>
					</div>
				</body>
				</html>
			`))
			return
		}
		// My implementation constructs /bible/111/JHN.3 for GetVerse("John", "3", "1", ...)
		// Let's check what I implemented.
		// url := fmt.Sprintf("%s/bible/%s/%s.%s", s.baseURL, version, usfmBook, chapter)
		// So path is /bible/111/JHN.3
		if r.URL.Path == "/bible/111/JHN.3" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
				<html>
				<body>
					<div class="chapter">
						<span data-usfm="JHN.3.1" class="verse v1">For God so loved the world</span>
						<span data-usfm="JHN.3.2" class="verse v2">that he gave his one and only Son</span>
					</div>
				</body>
				</html>
			`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	scraper := NewScraper()
	scraper.baseURL = server.URL

	// Test case 1: Single verse
	text, err := scraper.GetVerse("John", "3", "1", "111")
	assert.NoError(t, err)
	assert.Equal(t, "For God so loved the world", text)

	// Test case 2: Verse range
	text, err = scraper.GetVerse("John", "3", "1-2", "111")
	assert.NoError(t, err)
	assert.Equal(t, "For God so loved the world that he gave his one and only Son", text)
}

func TestGetVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/versions" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
				<html>
				<body>
					<div class="versions-list">
						<a href="/versions/111-niv-new-international-version">New International Version (NIV)</a>
						<a href="/versions/1-kjv-king-james-version">King James Version (KJV)</a>
						<a href="/other">Irrelevant Link</a>
					</div>
				</body>
				</html>
			`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	scraper := NewScraper()
	scraper.baseURL = server.URL

	versions, err := scraper.GetVersions()
	assert.NoError(t, err)
	assert.Len(t, versions, 2)

	assert.Equal(t, "New International Version", versions[0].Name)
	assert.Equal(t, "111", versions[0].Value) // ID
	assert.Equal(t, "NIV", versions[0].Code)

	assert.Equal(t, "King James Version", versions[1].Name)
	assert.Equal(t, "1", versions[1].Value)
	assert.Equal(t, "KJV", versions[1].Code)
}

func TestSearchWords(t *testing.T) {
	scraper := NewScraper()
	results, err := scraper.SearchWords("test", "NIV")
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "not supported")
}
