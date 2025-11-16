package biblegateway

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func TestGetVerse(t *testing.T) {
	testCases := []struct {
		name       string
		book       string
		chapter    string
		verse      string
		version    string
		htmlFile   string
		expected   string
		expectFail bool
	}{
		{
			name:     "John 3:16",
			book:     "John",
			chapter:  "3",
			verse:    "16",
			version:  "ESV",
			htmlFile: "testdata/get_verse_success.html",
			expected: "16 For God so loved the world, that he gave his only Son, that whoever believes in him should not perish but have eternal life.",
		},
		{
			name:     "Proverbs 1:7",
			book:     "Proverbs",
			chapter:  "1",
			verse:    "7",
			version:  "ESV",
			htmlFile: "testdata/get_verse_proverbs.html",
			expected: "7 The fear of the Lord is the beginning of knowledge; fools despise wisdom and instruction.",
		},
		{
			name:     "Psalm 23:1",
			book:     "Psalm",
			chapter:  "23",
			verse:    "1",
			version:  "NIV",
			htmlFile: "testdata/get_verse_psalms.html",
			expected: "1 The Lord is my shepherd, I shall not want.",
		},
		{
			name:     "Genesis 1:1",
			book:     "Genesis",
			chapter:  "1",
			verse:    "1",
			version:  "ESV",
			htmlFile: "testdata/get_verse_genesis.html",
			expected: "1 In the beginning, God created the heavens and the earth.",
		},
		{
			name:     "Romans 8:28-30",
			book:     "Romans",
			chapter:  "8",
			verse:    "28-30",
			version:  "ESV",
			htmlFile: "testdata/get_verse_romans.html",
			expected: "28 And we know that for those who love God all things work together for good, for those who are called according to his purpose. 29 For those whom he foreknew he also predestined to be conformed to the image of his Son, in order that he might be the firstborn among many brothers. 30 And those whom he predestined he also called, and those whom he called he also justified, and those whom he justified he also glorified.",
		},
		{
			name:       "verse not found",
			book:       "Invalid",
			chapter:    "1",
			verse:      "1",
			version:    "ESV",
			htmlFile:   "testdata/get_verse_not_found.html",
			expectFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				html, err := os.ReadFile(tc.htmlFile)
				if err != nil {
					t.Fatalf("failed to read mock html file: %v", err)
				}
				fmt.Fprintln(w, string(html))
			}))
			defer server.Close()

			scraper := &Scraper{client: server.Client(), baseURL: server.URL}

			verse, err := scraper.GetVerse(tc.book, tc.chapter, tc.verse, tc.version)
			if tc.expectFail {
				if err == nil {
					t.Fatal("expected an error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			normalizedVerse := normalizeSpace(verse)
			normalizedExpected := normalizeSpace(tc.expected)

			if !strings.Contains(normalizedVerse, normalizedExpected) {
				t.Errorf("expected verse to contain %q, but got %q", normalizedExpected, normalizedVerse)
			}
		})
	}
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
