package biblegateway

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var allowedTags = map[string]bool{
	"h1":   true,
	"h2":   true,
	"h3":   true,
	"h4":   true,
	"p":    true,
	"span": true,
	"i":    true,
	"br":   true,
	"sup":  true, // Preserved for verse numbers
}

func sanitizeNodes(n *html.Node) {
	// Iterate backwards to handle removals safely
	for c := n.LastChild; c != nil; {
		prev := c.PrevSibling
		sanitizeNodes(c)
		c = prev
	}

	if n.Type == html.CommentNode {
		n.Parent.RemoveChild(n)
		return
	}

	if n.Type == html.ElementNode {
		if n.Data == "script" || n.Data == "style" {
			n.Parent.RemoveChild(n)
			return
		}

		if !allowedTags[n.Data] {
			// Unwrap: Move children to before n, then remove n
			parent := n.Parent
			if parent != nil {
				for child := n.FirstChild; child != nil; {
					next := child.NextSibling
					n.RemoveChild(child)
					parent.InsertBefore(child, n)
					child = next
				}
				parent.RemoveChild(n)
			}
		}
	}
}

func strictSanitize(s *goquery.Selection) {
	for _, n := range s.Nodes {
		// Sanitize children of the selected nodes (to preserve the container if needed, though unwrap handles it)
		// We iterate backwards on children
		for c := n.LastChild; c != nil; {
			prev := c.PrevSibling
			sanitizeNodes(c)
			c = prev
		}
	}
}

func removeUnwantedElements(s *goquery.Selection) {
	s.Find(".footnote, .footnotes, .chapternum, .crossreference, .crossrefs, .publisher-info-bottom, .dropdown-version-switcher, .passage-scroller, .full-chap-link, .other-translations").Remove()
	s.Find("sup:not(.versenum)").Remove()
	s.Find("a").FilterFunction(func(i int, sel *goquery.Selection) bool {
		return strings.Contains(sel.Text(), "in all English translations")
	}).Remove()
}

func unwrapSmallCaps(s *goquery.Selection) {
	s.Find(".small-caps").Each(func(i int, sel *goquery.Selection) {
		sel.ReplaceWithHtml(sel.Text())
	})
}

func removeAttributes(s *goquery.Selection) {
	s.Find("*").RemoveAttr("class").RemoveAttr("id").RemoveAttr("style")
}

func removeEmptyParagraphs(s *goquery.Selection) {
	s.Find("p").Each(func(i int, sel *goquery.Selection) {
		if strings.TrimSpace(sel.Text()) == "" && sel.Find("br").Length() == 0 {
			sel.Remove()
		}
	})
}

func scrapeProse(s *goquery.Selection) (string, error) {
	removeUnwantedElements(s)
	unwrapSmallCaps(s)
	strictSanitize(s)
	removeAttributes(s)
	removeEmptyParagraphs(s)

	html, err := s.Html()
	if err != nil {
		return "", err
	}

	html = strings.ReplaceAll(html, "\u00a0", " ")
	html = strings.ReplaceAll(html, "\n", "")
	html = strings.ReplaceAll(html, "\r", "")
	html = strings.ReplaceAll(html, "<br/> ", "<br/>")
	re := regexp.MustCompile(`>\s+<`)
	html = re.ReplaceAllString(html, "><")
	re = regexp.MustCompile(`\s+`)
	html = re.ReplaceAllString(html, " ")
	html = strings.ReplaceAll(html, " >", ">")
	html = strings.ReplaceAll(html, " </span>", "</span>")
	html = strings.ReplaceAll(html, " </p>", "</p>")
	html = strings.ReplaceAll(html, " </h4>", "</h4>")
	html = strings.ReplaceAll(html, " </h3>", "</h3>")

	return html, nil
}

func scrapePoetry(s *goquery.Selection) (string, error) {
	removeUnwantedElements(s)
	unwrapSmallCaps(s)

	s.Find("div.poetry.top-1 br").Remove()
	s.Find("p.top-1").ReplaceWithHtml("<br/>")
	s.Find("div.poetry, p.line, span.indent-1").Each(func(i int, sel *goquery.Selection) {
		html, _ := sel.Html()
		sel.ReplaceWithHtml(html)
	})

	strictSanitize(s)
	removeAttributes(s)
	removeEmptyParagraphs(s)

	html, err := s.Html()
	if err != nil {
		return "", err
	}

	html = strings.ReplaceAll(html, "\u00a0", " ")
	html = strings.ReplaceAll(html, "\n", "")
	html = strings.ReplaceAll(html, "\r", "")
	html = strings.ReplaceAll(html, "<br/> ", "<br/>")

	return html, nil
}

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
	Text  string `json:"text"`
	URL   string `json:"url"`
}

// GetVerse fetches a single Bible verse by reference and returns it as sanitized HTML.
func (s *Scraper) GetVerse(book, chapter, verse, version string) (string, error) {
	reference := fmt.Sprintf("%s %s:%s", book, chapter, verse)
	if verse == "" {
		reference = fmt.Sprintf("%s %s", book, chapter)
	}
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

	return sanitizeSelection(passageSelection)
}

func sanitizeSelection(s *goquery.Selection) (string, error) {
	isPoetry := s.Find("div.poetry").Length() > 0

	var html string
	var err error

	if isPoetry {
		html, err = scrapePoetry(s)
	} else {
		html, err = scrapeProse(s)
	}

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(html), nil
}

func sanitizeSnippet(s *goquery.Selection) (string, error) {
	s.Find(".bible-item-extras").Remove()
	s.Find("h1, h2, h3, h4, h5, h6").Remove()
	return sanitizeSelection(s)
}

// SearchWords searches for a word or phrase and returns a list of relevant verses.
func (s *Scraper) SearchWords(query, version string) ([]SearchResult, error) {
	encodedQuery := url.QueryEscape(query)
	url := s.baseURL + fmt.Sprintf("/quicksearch/?quicksearch=%s&version=%s&interface=print", encodedQuery, version)
	log.Printf("Scraping URL: %s", url)

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
		log.Printf("Failed to search, status code: %d", res.StatusCode)
		return nil, fmt.Errorf("failed to search, status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	results := []SearchResult{}
	selection := doc.Find(".search-result-list .bible-item")
	log.Printf("Found %d search results for query '%s'", selection.Length(), query)

	selection.Each(func(i int, sel *goquery.Selection) {
		titleLink := sel.Find(".bible-item-title")
		verse := titleLink.Text()
		url, _ := titleLink.Attr("href")

		textSel := sel.Find(".bible-item-text")
		text, _ := sanitizeSnippet(textSel)

		results = append(results, SearchResult{
			Verse: verse,
			Text:  text,
			URL:   s.baseURL + url,
		})
	})

	return results, nil
}
