package biblehub

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVerse(t *testing.T) {
	// Mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve mocked HTML
		// URL: /esv/john/3.htm
		if r.URL.Path == "/esv/john/3.htm" {
			fmt.Fprintln(w, `
<html>
<body>
<p class="regular">
  <span class="reftext"><a href="...">16</a></span>
  For God so loved the world, that he gave his only Son, that whoever believes in him should not perish but have eternal life.
  <span class="reftext"><a href="...">17</a></span>
  For God did not send his Son into the world to condemn the world, but in order that the world might be saved through him.
</p>
</body>
</html>
			`)
		} else if r.URL.Path == "/esv/john/1.htm" {
			fmt.Fprintln(w, `
<html><body><p class="regular">
<span class="reftext">12</span>But to all who did receive him, who believed in his name, he gave the right to become children of God,
</p></body></html>`)
		} else if r.URL.Path == "/esv/john/2.htm" {
			fmt.Fprintln(w, `
<html><body><p class="regular">
<span class="reftext">1</span>On the third day there was a wedding at Cana in Galilee, and the mother of Jesus was there.
<span class="reftext">2</span>Jesus also was invited to the wedding with his disciples.
<span class="reftext">3</span>When the wine ran out, the mother of Jesus said to him, “They have no wine.”
<span class="reftext">4</span>And Jesus said to her, “Woman, what does this have to do with me? My hour has not yet come.”
</p></body></html>`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	scraper := NewScraper()
	scraper.baseURL = ts.URL
	scraper.client = ts.Client()

	// Test case 1: Single verse
	verse, err := scraper.GetVerse("John", "3", "16", "esv")
	assert.NoError(t, err)
	// The scraper might include whitespace/newlines depending on how nodes are processed.
	// We check for content presence.
	assert.Contains(t, verse, "For God so loved the world")
	assert.NotContains(t, verse, "For God did not send his Son")

	// Test case 2: Verse range
	verse, err = scraper.GetVerse("John", "3", "16-17", "esv")
	assert.NoError(t, err)
	assert.Contains(t, verse, "For God so loved the world")
	assert.Contains(t, verse, "For God did not send his Son")

	// Test case 3: Cross-chapter range
	verse, err = scraper.GetVerse("John", "1", "12-2:4", "esv")
	assert.NoError(t, err)
	assert.Contains(t, verse, "But to all who did receive him") // 1:12
	assert.Contains(t, verse, "On the third day")               // 2:1
	assert.Contains(t, verse, "And Jesus said to her")          // 2:4
}

func TestSearchWords(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search.php" && r.URL.Query().Get("q") == "love" {
			fmt.Fprintln(w, `
<html>
<body>
<div class="result_block">
    <div class="result_title"><a href="/john/3-16.htm">John 3:16</a></div>
    <div class="description">For God so loved the world...</div>
</div>
<div class="result_altblock">
    <div class="result_title"><a href="/1_john/4-8.htm">1 John 4:8</a></div>
    <div class="description">God is love.</div>
</div>
</body>
</html>
            `)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	scraper := NewScraper()
	scraper.baseURL = ts.URL
	scraper.client = ts.Client()

	results, err := scraper.SearchWords("love", "esv")
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "John 3:16", results[0].Verse)
	assert.Equal(t, "For God so loved the world...", results[0].Text)
	assert.Equal(t, ts.URL+"/john/3-16.htm", results[0].URL)

	assert.Equal(t, "1 John 4:8", results[1].Verse)
}
