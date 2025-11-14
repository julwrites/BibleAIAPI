package biblegateway

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetVerse(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			html, err := os.ReadFile("testdata/get_verse_success.html")
			if err != nil {
				t.Fatalf("failed to read mock html file: %v", err)
			}
			fmt.Fprintln(w, string(html))
		}))
		defer server.Close()

		scraper := &Scraper{client: server.Client(), baseURL: server.URL}

		verse, err := scraper.GetVerse("John", "3", "16", "ESV")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "For God so loved the world, that he gave his only Son, that whoever believes in him should not perish but have eternal life."
		if !strings.Contains(verse, expected) {
			t.Errorf("expected verse to contain %q, but got %q", expected, verse)
		}
	})

	t.Run("verse not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			html, err := os.ReadFile("testdata/get_verse_not_found.html")
			if err != nil {
				t.Fatalf("failed to read mock html file: %v", err)
			}
			fmt.Fprintln(w, string(html))
		}))
		defer server.Close()

		scraper := &Scraper{client: server.Client(), baseURL: server.URL}

		_, err := scraper.GetVerse("Invalid", "1", "1", "ESV")
		if err == nil {
			t.Fatal("expected an error, but got nil")
		}
	})
}

func TestSearchWords(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			html, err := os.ReadFile("testdata/search_words_success.html")
			if err != nil {
				t.Fatalf("failed to read mock html file: %v", err)
			}
			fmt.Fprintln(w, string(html))
		}))
		defer server.Close()

		scraper := &Scraper{client: server.Client(), baseURL: server.URL}

		results, err := scraper.SearchWords("grace", "ESV")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 results, but got %d", len(results))
		}

		expected := []struct {
			verse string
			url   string
		}{
			{verse: "Romans 3:24", url: "/passage/?search=Romans+3%3A24&version=ESV"},
			{verse: "Ephesians 2:8", url: "/passage/?search=Ephesians+2%3A8&version=ESV"},
		}

		for i, res := range results {
			if res.Verse != expected[i].verse {
				t.Errorf("expected verse %q, but got %q", expected[i].verse, res.Verse)
			}
			if res.URL != server.URL+expected[i].url {
				t.Errorf("expected url %q, but got %q", server.URL+expected[i].url, res.URL)
			}
		}
	})

	t.Run("no results", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			html, err := os.ReadFile("testdata/search_words_no_results.html")
			if err != nil {
				t.Fatalf("failed to read mock html file: %v", err)
			}
			fmt.Fprintln(w, string(html))
		}))
		defer server.Close()

		scraper := &Scraper{client: server.Client(), baseURL: server.URL}

		results, err := scraper.SearchWords("nonexistentword", "ESV")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 results, but got %d", len(results))
		}
	})
}
