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

// GetVerse fetches a single Bible verse by reference.
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

	passage := doc.Find(".passage-text")
	passage.Find(".footnotes").Remove()
	passage.Find(".crossrefs").Remove()

	verseText := passage.Text()

	if verseText == "" || strings.Contains(verseText, "No results found") {
		return "", fmt.Errorf("verse not found")
	}

	return strings.TrimSpace(verseText), nil
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
