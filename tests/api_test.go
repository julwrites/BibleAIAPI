package tests

import (
	"bible-api-service/internal/bible"
	"bible-api-service/internal/bible/providers/biblegateway"
	"bible-api-service/internal/handlers"
	"bible-api-service/internal/secrets"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestAPI_WordSearch_CleanOutput(t *testing.T) {
	// 1. Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "quicksearch") && r.URL.Query().Get("quicksearch") == "love" {
			content, err := os.ReadFile("testdata/search_love.html")
			if err != nil {
				t.Logf("Error reading file: %v", err)
				http.Error(w, "file not found", http.StatusInternalServerError)
				return
			}
			w.Write(content)
			return
		}
		if strings.Contains(r.URL.Path, "passage") && r.URL.Query().Get("search") == "John 3:16" {
			content, err := os.ReadFile("testdata/verse_john316.html")
			if err != nil {
				t.Logf("Error reading file: %v", err)
				http.Error(w, "file not found", http.StatusInternalServerError)
				return
			}
			w.Write(content)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// 2. Setup Handler with Mock Scraper
	scraper := biblegateway.NewScraper()
	scraper.SetBaseURL(server.URL)

	handler := handlers.NewQueryHandler(&secrets.EnvClient{})
	handler.BibleGatewayClient = scraper

	// 3. Test Word Search
	t.Run("Word Search - Love", func(t *testing.T) {
		reqBody := `{"query": {"words": ["love"]}}`
		req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var results []bible.SearchResult
		if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("expected results, got none")
		}

		// Check ALL results for unwanted content
		unwanted := []string{"<h3>", "</h3>", "<b>", "</b>", "In Context", "Full Chapter", "Other Translations"}

		for i, res := range results {
			for _, u := range unwanted {
				if strings.Contains(res.Text, u) {
					t.Errorf("result %d (%s) contains unwanted content '%s'. \nGot: %s", i, res.Verse, u, res.Text)
				}
			}
		}

		// Verify first result (Genesis 22:2)
		firstResult := results[0]
		if !strings.Contains(firstResult.Text, "Take your son, your only son Isaac") {
			t.Errorf("expected text missing in first result. \nGot: %s", firstResult.Text)
		}
	})

	// 4. Test Verse Query
	t.Run("Verse Query - John 3:16", func(t *testing.T) {
		reqBody := `{"query": {"verses": ["John 3:16"]}}`
		req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var response map[string]string
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		verse := response["verse"]

		// Verse query SHOULD preserve some formatting, but definitely not metadata
		unwanted := []string{"In Context", "Full Chapter", "Other Translations"}
		for _, u := range unwanted {
			if strings.Contains(verse, u) {
				t.Errorf("response verse contains unwanted content '%s'. \nGot: %s", u, verse)
			}
		}

		if !strings.Contains(verse, "For God so loved the world") {
			t.Errorf("expected verse text missing. \nGot: %s", verse)
		}
	})
}
