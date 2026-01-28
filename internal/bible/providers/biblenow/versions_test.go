package biblenow

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"bible-api-service/internal/bible"
)

func TestGetVersions(t *testing.T) {
	// Mock English page with language links
	mockEnglishPage := `
<!DOCTYPE html>
<html>
<body>
    <div class="menu">
        <a href="%s/en/bible/king-james-version">King James Version (KJV)</a>
        <a href="%s/es">Español (ES)</a>
    </div>
</body>
</html>
`
	// Mock Spanish page with version links
	mockSpanishPage := `
<!DOCTYPE html>
<html>
<body>
    <div class="menu">
        <a href="%s/es/biblia/reina-valera-1909">Reina-Valera 1909 (RVR1909)</a>
        <!-- Should be ignored -->
        <a href="%s/es/biblia/reina-valera-1909/antiguo-testamento">Antiguo Testamento</a>
    </div>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baseURL := "http://" + r.Host
		if r.URL.Path == "/en/bible" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(mockEnglishPage, baseURL, baseURL)))
			return
		}
		if r.URL.Path == "/es" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(mockSpanishPage, baseURL, baseURL)))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	scraper := NewScraper()
	scraper.baseURL = server.URL // Override base URL to point to mock server

	versions, err := scraper.GetVersions()
	if err != nil {
		t.Fatalf("GetVersions failed: %v", err)
	}

	if len(versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(versions))
	}

	// Sort for stability
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Code < versions[j].Code
	})

	expected := []bible.ProviderVersion{
		{Name: "King James Version", Code: "KJV", Value: "en/bible/king-james-version", Language: "English"},
		{Name: "Reina-Valera 1909", Code: "RVR1909", Value: "es/biblia/reina-valera-1909", Language: "Español"},
	}

	for i, v := range versions {
		if v.Name != expected[i].Name {
			t.Errorf("Version %d: Expected Name %s, got %s", i, expected[i].Name, v.Name)
		}
		if v.Code != expected[i].Code {
			t.Errorf("Version %d: Expected Code %s, got %s", i, expected[i].Code, v.Code)
		}
		if v.Value != expected[i].Value {
			t.Errorf("Version %d: Expected Value %s, got %s", i, expected[i].Value, v.Value)
		}
		if v.Language != expected[i].Language {
			t.Errorf("Version %d: Expected Language %s, got %s", i, expected[i].Language, v.Language)
		}
	}
}
