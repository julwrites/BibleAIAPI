package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bible-api-service/internal/bible"
	"bible-api-service/internal/bible/providers/biblecom"
	"bible-api-service/internal/bible/providers/biblegateway"
	"bible-api-service/internal/bible/providers/biblehub"
	"bible-api-service/internal/bible/providers/biblenow"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestRun_Success(t *testing.T) {
	// Mock Bible Gateway Server
	bgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/versions/", r.URL.Path)
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
		w.Write([]byte(html))
	}))
	defer bgServer.Close()

	// Mock BibleHub Server
	bhServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/genesis/1-1.htm", r.URL.Path)
		html := `
	<html>
		<body>
			<a href="/esv/genesis/1.htm">English Standard Version</a>
			<a href="/kjv/genesis/1.htm">King James Version</a>
		</body>
	</html>
	`
		w.Write([]byte(html))
	}))
	defer bhServer.Close()

	// Mock BibleNow Server
	bnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/en/bible", r.URL.Path)
		html := `
	<html>
		<body>
			<a href="/en/bible/english-standard-version">English Standard Version (ESV)</a>
			<a href="/en/bible/new-international-version">New International Version (NIV)</a>
		</body>
	</html>
	`
		w.Write([]byte(html))
	}))
	defer bnServer.Close()

	// Mock Bible.com Server
	bcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/versions", r.URL.Path)
		html := `
	<html>
		<body>
			<a href="/versions/59-esv-english-standard-version">English Standard Version (ESV)</a>
			<a href="/versions/111-niv-new-international-version">New International Version (NIV)</a>
		</body>
	</html>
	`
		w.Write([]byte(html))
	}))
	defer bcServer.Close()

	// Setup Scrapers
	bgScraper := biblegateway.NewScraper()
	bgScraper.SetBaseURL(bgServer.URL)

	bhScraper := biblehub.NewScraper()
	bhScraper.SetBaseURL(bhServer.URL)

	bnScraper := biblenow.NewScraper()
	bnScraper.SetBaseURL(bnServer.URL)

	bcScraper := biblecom.NewScraper()
	bcScraper.SetBaseURL(bcServer.URL)

	providers := map[string]bible.Provider{
		"biblegateway": bgScraper,
		"biblehub":     bhScraper,
		"biblenow":     bnScraper,
		"biblecom":     bcScraper,
	}

	// Setup Temp Output File
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "versions.yaml")

	// Run
	err := run(providers, outputPath)
	require.NoError(t, err)

	// Verify File Content
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	var versions []bible.Version
	err = yaml.Unmarshal(data, &versions)
	require.NoError(t, err)

	// Check for ESV (present in all)
	esv := findVersion(versions, "ESV")
	require.NotNil(t, esv)
	assert.Equal(t, "English Standard Version", esv.Name)
	// BibleGateway sets language to Unknown because it was before ---Spanish---
	assert.Equal(t, "Unknown", esv.Language)
	assert.Equal(t, "ESV", esv.Providers["biblegateway"])
	assert.Equal(t, "esv", esv.Providers["biblehub"])
	assert.Equal(t, "en/bible/english-standard-version", esv.Providers["biblenow"])
	assert.Equal(t, "59", esv.Providers["biblecom"])

	// Check for RVR1960 (only in BibleGateway)
	rvr := findVersion(versions, "RVR1960")
	require.NotNil(t, rvr)
	assert.Equal(t, "Reina-Valera 1960", rvr.Name)
	assert.Equal(t, "Spanish", rvr.Language)
	assert.Equal(t, "RVR1960", rvr.Providers["biblegateway"])
	assert.NotContains(t, rvr.Providers, "biblehub")

	// Check for KJV (only in BibleHub)
	kjv := findVersion(versions, "KJV")
	require.NotNil(t, kjv)
	assert.Equal(t, "King James Version", kjv.Name)
	assert.Equal(t, "English", kjv.Language) // BibleHub default
	assert.Equal(t, "kjv", kjv.Providers["biblehub"])

	// Check for NIV (in BibleNow and Bible.com)
	niv := findVersion(versions, "NIV")
	require.NotNil(t, niv)
	assert.Equal(t, "english", strings.ToLower(niv.Language)) // BibleNow/BibleCom default
	assert.Contains(t, niv.Providers, "biblenow")
	assert.Contains(t, niv.Providers, "biblecom")
}

func findVersion(versions []bible.Version, code string) *bible.Version {
	for _, v := range versions {
		if v.Code == code {
			return &v
		}
	}
	return nil
}

func TestRun_ScraperError(t *testing.T) {
	// Mock Server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := biblegateway.NewScraper()
	scraper.SetBaseURL(server.URL)

	providers := map[string]bible.Provider{
		"biblegateway": scraper,
	}

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "versions.yaml")

	err := run(providers, outputPath)
	// Expect no error, just logging and skipping
	assert.NoError(t, err)

	// File should exist but contain no versions
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	var versions []bible.Version
	err = yaml.Unmarshal(data, &versions)
	require.NoError(t, err)
	assert.Empty(t, versions)
}
