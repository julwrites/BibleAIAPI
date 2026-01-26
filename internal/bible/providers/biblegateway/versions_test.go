package biblegateway

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScraper_GetVersions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		html := `
		<html>
			<body>
				<select class="search-dropdown" name="version">
					<option value="ESV">English Standard Version</option>
					<option>---Spanish---</option>
					<option value="RVR1960">Reina-Valera 1960</option>
					<option value="">Empty Value</option>
					<option value="SPACE">&nbsp;</option>
				</select>
			</body>
		</html>
		`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/versions/", r.URL.Path)
			w.Write([]byte(html))
		}))
		defer server.Close()

		scraper := NewScraper()
		scraper.SetBaseURL(server.URL)

		versions, err := scraper.GetVersions()
		require.NoError(t, err)

		// Expected: ESV (Unknown lang), RVR1960 (Spanish lang).
		// Empty value and non-breaking space options should be skipped.
		// "---Spanish---" should set the language but not be a version.
		assert.Len(t, versions, 2)

		assert.Equal(t, "ESV", versions[0].Value)
		assert.Equal(t, "English Standard Version", versions[0].Name)
		assert.Equal(t, "Unknown", versions[0].Language)

		assert.Equal(t, "RVR1960", versions[1].Value)
		assert.Equal(t, "Reina-Valera 1960", versions[1].Name)
		assert.Equal(t, "Spanish", versions[1].Language)
	})

	t.Run("HTTPError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		scraper := NewScraper()
		scraper.SetBaseURL(server.URL)

		_, err := scraper.GetVersions()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status code: 500")
	})

	t.Run("ParsingComplexStructure", func(t *testing.T) {
		html := `
		<select class="search-dropdown" name="version">
			<option>---English---</option>
			<option value="NIV">New International Version</option>
			<option>---Ancient Greek (Koine)---</option>
			<option value="SBLGNT">SBL Greek New Testament</option>
		</select>
		`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(html))
		}))
		defer server.Close()

		scraper := NewScraper()
		scraper.SetBaseURL(server.URL)

		versions, err := scraper.GetVersions()
		require.NoError(t, err)
		assert.Len(t, versions, 2)

		assert.Equal(t, "NIV", versions[0].Value)
		assert.Equal(t, "English", versions[0].Language)

		assert.Equal(t, "SBLGNT", versions[1].Value)
		// Assuming the logic trims "---" and "---"
		assert.Equal(t, "Ancient Greek (Koine)", versions[1].Language)
	})

	t.Run("ConnectionRefused", func(t *testing.T) {
		scraper := NewScraper()
		scraper.SetBaseURL("http://localhost:12345") // Nothing listening here

		_, err := scraper.GetVersions()
		assert.Error(t, err)
	})
}
