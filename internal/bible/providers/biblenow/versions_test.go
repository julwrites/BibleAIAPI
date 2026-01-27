package biblenow

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bible-api-service/internal/bible"
)

func TestGetVersions(t *testing.T) {
	// Mock HTML response that simulates the BibleNow versions page
	mockHTML := `
<!DOCTYPE html>
<html>
<body>
    <div class="menu">
        <ul>
            <li><a href="/en/bible/king-james-version">King James Version (KJV)</a></li>
            <li><a href="/en/bible/american-standard-version">American Standard Version (ASV)</a></li>
            <li><a href="/en/bible/new-international-version">New International Version (NIV)</a></li>
            <!-- Some other links to ignore -->
            <li><a href="/en/bible/king-james-version/genesis">Genesis</a></li>
            <li><a href="/en/about">About</a></li>
        </ul>
    </div>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/en/bible" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockHTML))
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

	if len(versions) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(versions))
	}

	expected := []bible.ProviderVersion{
		{Name: "King James Version", Code: "KJV", Value: "king-james-version", Language: "English"},
		{Name: "American Standard Version", Code: "ASV", Value: "american-standard-version", Language: "English"},
		{Name: "New International Version", Code: "NIV", Value: "new-international-version", Language: "English"},
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
	}
}
