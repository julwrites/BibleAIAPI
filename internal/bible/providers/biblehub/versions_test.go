package biblehub

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bible-api-service/internal/bible"
)

func TestGetVersions(t *testing.T) {
	// Mock HTML response that simulates BibleHub versions list
	mockHTML := `
<!DOCTYPE html>
<html>
<body>
    <div class="versions">
        <a href="/kjv/genesis/1.htm">King James Version</a>
        <p>In the beginning God created the heaven and the earth.</p>

        <a href="/esv/genesis/1.htm">English Standard Version</a>
        <p>In the beginning, God created the heavens and the earth.</p>

        <a href="/niv/genesis/1.htm">New International Version</a>
        <p>In the beginning God created the heavens and the earth.</p>

        <!-- Some other links to ignore -->
        <a href="/genesis/1-2.htm">Next Verse</a>
        <a href="/study/genesis/1.htm">Study</a>
    </div>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/genesis/1-1.htm" {
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
		{Name: "King James Version", Code: "KJV", Value: "kjv", Language: "English"},
		{Name: "English Standard Version", Code: "ESV", Value: "esv", Language: "English"},
		{Name: "New International Version", Code: "NIV", Value: "niv", Language: "English"},
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
