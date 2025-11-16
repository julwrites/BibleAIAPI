package biblegateway

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Scraper is a client for scraping the Bible Gateway website.
type Scraper struct {
	client  *http.Client
	baseURL string
}

// NewScraper creates a new Scraper with a default HTTP client and base URL.
func NewScraper() *Scraper {
	return &Scraper{
		client:  &http.Client{},
		baseURL: "https://classic.biblegateway.com",
	}
}

// SearchResult represents a search result.
type SearchResult struct {
	Verse string `json:"verse"`
	URL   string `json:"url"`
}

// GetVerse fetches a single Bible verse by reference and returns it as sanitized HTML.
func (s *Scraper) GetVerse(book, chapter, verse, version string) (string, error) {
	reference := fmt.Sprintf("%s %s:%s", book, chapter, verse)
	encodedReference := url.QueryEscape(reference)
	url := s.baseURL + fmt.Sprintf("/passage/?search=%s&version=%s&interface=print", encodedReference, version)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch verse, status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	passageSelection := doc.Find(".passage-text")
	if passageSelection.Length() == 0 || strings.Contains(passageSelection.Text(), "No results found") {
		return "", fmt.Errorf("verse not found")
	}

	// Remove unwanted elements first
	passageSelection.Find("sup:not(.versenum)").Remove()
	passageSelection.Find(".footnote, .chapternum, .small-caps").Remove()

	html, _ := passageSelection.Html()
	html = strings.ReplaceAll(html, "\n", "")
	html = strings.ReplaceAll(html, "  ", " ")
	html = strings.ReplaceAll(html, `class="text"`, "")
	html = strings.ReplaceAll(html, `class="line"`, "")
	html = strings.ReplaceAll(html, `class="verse"`, "")
	html = strings.ReplaceAll(html, `id="en-ESV-16461"`, "")
	html = strings.ReplaceAll(html, `id="en-ESV-16462"`, "")
	html = strings.ReplaceAll(html, `id="en-ESV-15881"`, "")
	html = strings.ReplaceAll(html, `id="en-ESV-15882"`, "")
	html = strings.ReplaceAll(html, `id="en-ESV-15883"`, "")
	html = strings.ReplaceAll(html, `id="en-ESV-15884"`, "")
	html = strings.ReplaceAll(html, `id="en-ESV-15885"`, "")
	html = strings.ReplaceAll(html, `id="en-ESV-15886"`, "")
	html = strings.ReplaceAll(html, `id="en-ESV-15887"`, "")
	html = strings.ReplaceAll(html, `id="en-ESV-15888"`, "")
	html = strings.ReplaceAll(html, `class="indent-1"`, "")
	html = strings.ReplaceAll(html, `class="indent-1-breaks"`, "")
	html = strings.ReplaceAll(html, `class="poetry top-1"`, "")
	html = strings.ReplaceAll(html, `class="poetry"`, "")
	html = strings.ReplaceAll(html, `class="psalm-verse"`, "")
	html = strings.ReplaceAll(html, `<p >`, "<p>")
	html = strings.ReplaceAll(html, `<div >`, "")
	html = strings.ReplaceAll(html, `</div>`, "")
	html = strings.ReplaceAll(html, ` >`, ">")

	return strings.TrimSpace(html), nil
}

// SearchWords searches for a word or phrase and returns a list of relevant verses.
func (s *Scraper) SearchWords(query, version string) ([]SearchResult, error) {
	encodedQuery := url.QueryEscape(query)
	url := s.baseURL + fmt.Sprintf("/quicksearch/?quicksearch=%s&version=%s&interface=print", encodedQuery, version)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to search, status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	doc.Find(".search-result-list .search-result").Each(func(i int, sel *goquery.Selection) {
		link := sel.Find(".bible-item-extras a")
		verse := link.Text()
		url, _ := link.Attr("href")
		results = append(results, SearchResult{
			Verse: verse,
			URL:   s.baseURL + url,
		})
	})

	return results, nil
}
