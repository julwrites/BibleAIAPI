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
			expected: `<h3>For God So Loved the World</h3> <p> <span> <sup>16 </sup>“For God so loved the world, that he gave his only Son, that whoever believes in him should not perish but have eternal life. </span> </p>`,
		},
		{
			name:     "Matthew 28:19-20",
			book:     "Matthew",
			chapter:  "28",
			verse:    "19-20",
			version:  "ESV",
			htmlFile: "testdata/get_verse_matthew.html",
			expected: `<p> <span> <sup>19 </sup>Go therefore and make disciples of all nations, baptizing them in the name of the Father and of the Son and of the Holy Spirit, </span> <span> <sup>20 </sup>teaching them to observe all that I have commanded you. And behold, I am with you always, to the end of the age.” </span> </p>`,
		},
		{
			name:     "Proverbs 3:5-6",
			book:     "Proverbs",
			chapter:  "3",
			verse:    "5-6",
			version:  "ESV",
			htmlFile: "testdata/get_verse_proverbs.html",
			expected: `<p><span><sup>5 </sup>Trust in the Lord with all your heart,</span><br/><span>and do not lean on your own understanding.</span></p> <p><span><sup>6 </sup>In all your ways acknowledge him,</span><br/><span>and he will make straight your paths.</span></p>`,
		},
		{
			name:     "Bug Reproduction (John 3:16 with extras)",
			book:     "John",
			chapter:  "3",
			verse:    "16",
			version:  "ESV",
			htmlFile: "testdata/bug_repro.html",
			expected: `<h3><span>For God So Loved the World</span></h3> <p> <span><sup>16 </sup>“For God so loved the world, that he gave his only Son, that whoever believes in him should not perish but have eternal life.</span> </p>`,
		},
		{
			name:     "Psalm 121",
			book:     "Psalm",
			chapter:  "121",
			verse:    "",
			version:  "NIV",
			htmlFile: "testdata/get_verse_psalm_121.html",
			expected: `<h3>My Help Comes from the Lord</h3> <h4>A Song of Ascents.</h4> <p><span><sup>1 </sup>I lift up my eyes to the hills.</span><br/><span>From where does my help come?</span></p> <p><span><sup>2 </sup>My help comes from the Lord,</span><br/><span>who made heaven and earth.</span></p> <br/> <p><span><sup>3 </sup>He will not let your foot be moved;</span><br/><span>he who keeps you will not slumber.</span></p> <p><span><sup>4 </sup>Behold, he who keeps Israel</span><br/><span>will neither slumber nor sleep.</span></p> <br/> <p><span><sup>5 </sup>The Lord is your keeper;</span><br/><span>the Lord is your shade on your right hand.</span></p> <p><span><sup>6 </sup>The sun shall not strike you by day,</span><br/><span>nor the moon by night.</span></p> <br/> <p><span><sup>7 </sup>The Lord will keep you from all evil;</span><br/><span>he will keep your life.</span></p> <p><span><sup>8 </sup>The Lord will keep</span><br/><span>your going out and your coming in</span><br/><span>from this time forth and forevermore.</span></p>`,
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

			if normalizeSpace(verse) != normalizeSpace(tc.expected) {
				t.Errorf("expected verse to be:\n%s\nbut got:\n%s", tc.expected, verse)
			}
		})
	}
}

func TestSearchWords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html, err := os.ReadFile("testdata/search_words_success.html")
		if err != nil {
			t.Fatalf("failed to read mock html file: %v", err)
		}
		fmt.Fprintln(w, string(html))
	}))
	defer server.Close()

	scraper := &Scraper{client: server.Client(), baseURL: server.URL}
	results, err := scraper.SearchWords("love", "ESV")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []SearchResult{
		{
			Verse: "John 3:16",
			Text:  "For God so loved the world, that he gave his only Son, that whoever believes in him should not perish but have eternal life.",
			URL:   server.URL + "/passage/?search=John%203%3A16&version=ESV",
		},
		{
			Verse: "1 John 4:8",
			Text:  "Anyone who does not love does not know God, because God is love.",
			URL:   server.URL + "/passage/?search=1%20John%204%3A8&version=ESV",
		},
	}

	if len(results) != len(expected) {
		t.Fatalf("expected %d results, but got %d", len(expected), len(results))
	}

	for i, result := range results {
		if result.Verse != expected[i].Verse {
			t.Errorf("expected verse to be %s, but got %s", expected[i].Verse, result.Verse)
		}
		if result.Text != expected[i].Text {
			t.Errorf("expected text to be %s, but got %s", expected[i].Text, result.Text)
		}
		if result.URL != expected[i].URL {
			t.Errorf("expected url to be %s, but got %s", expected[i].URL, result.URL)
		}
	}
}

func TestNewScraper(t *testing.T) {
	scraper := NewScraper()
	if scraper.client == nil {
		t.Error("expected client to be initialized")
	}
	if scraper.baseURL != "https://classic.biblegateway.com" {
		t.Errorf("expected baseURL to be https://classic.biblegateway.com, but got %s", scraper.baseURL)
	}
}

func TestGetVerse_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := &Scraper{client: server.Client(), baseURL: server.URL}
	_, err := scraper.GetVerse("John", "3", "16", "ESV")
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}

func TestSearchWords_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := &Scraper{client: server.Client(), baseURL: server.URL}
	_, err := scraper.SearchWords("love", "ESV")
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}
