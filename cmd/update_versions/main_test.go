package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"bible-api-service/internal/biblegateway"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestRun_Success(t *testing.T) {
	// Mock Bible Gateway Server
	html := `
	<html>
		<body>
			<select class="search-dropdown" name="version">
				<option value="ESV">English Standard Version</option>
				<option>---Spanish---</option>
				<option value="RVR1960">Reina-Valera 1960</option>
			</select>
		</body>
	</html>
	`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/versions/", r.URL.Path)
		w.Write([]byte(html))
	}))
	defer server.Close()

	// Setup Scraper
	scraper := biblegateway.NewScraper()
	scraper.SetBaseURL(server.URL)

	// Setup Temp Output File
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "versions.yaml")

	// Run
	err := run(scraper, outputPath)
	require.NoError(t, err)

	// Verify File Content
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	var versions []biblegateway.Version
	err = yaml.Unmarshal(data, &versions)
	require.NoError(t, err)

	assert.Len(t, versions, 2)
	assert.Equal(t, "ESV", versions[0].Value)
	assert.Equal(t, "English Standard Version", versions[0].Name)
	assert.Equal(t, "Unknown", versions[0].Language)

	assert.Equal(t, "RVR1960", versions[1].Value)
	assert.Equal(t, "Reina-Valera 1960", versions[1].Name)
	assert.Equal(t, "Spanish", versions[1].Language)
}

func TestRun_ScraperError(t *testing.T) {
	// Mock Server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := biblegateway.NewScraper()
	scraper.SetBaseURL(server.URL)

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "versions.yaml")

	err := run(scraper, outputPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status code: 500")
}

func TestRun_FileWriteError(t *testing.T) {
	// Mock Successful Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html></html>"))
	}))
	defer server.Close()

	scraper := biblegateway.NewScraper()
	scraper.SetBaseURL(server.URL)

	// Invalid path (directory that cannot be written to or invalid path)
	// Using a path that implies a directory where a file exists
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "file")
	os.WriteFile(existingFile, []byte(""), 0644)
	outputPath := filepath.Join(existingFile, "versions.yaml") // Can't create file under a file

	err := run(scraper, outputPath)
	assert.Error(t, err)
}
