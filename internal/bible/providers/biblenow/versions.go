package biblenow

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"bible-api-service/internal/bible"

	"github.com/PuerkitoBio/goquery"
)

// GetVersions scrapes the available Bible versions from BibleNow.
// Currently, this fetches versions available on the English page.
// TODO: Iterate over other languages to get all versions.
func (s *Scraper) GetVersions() ([]bible.ProviderVersion, error) {
	// Fetch the English Bible page which lists versions
	url := s.baseURL + "/en/bible"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BibleAIAPI/1.0)")

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch versions, status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var versions []bible.ProviderVersion
	seen := make(map[string]bool)

	// Regex to match version links: /en/bible/{slug}
	// Note: We want to avoid links that go deeper like /en/bible/{slug}/{book}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// Check if it's a version link
		if !strings.HasPrefix(href, "/en/bible/") {
			return
		}

		// Remove prefix
		slug := strings.TrimPrefix(href, "/en/bible/")
		// If slug contains '/', it might be a book or chapter link (e.g. /en/bible/kjv/genesis)
		if strings.Contains(slug, "/") {
			return
		}

		// Also filter out 'books', 'old-testament', 'new-testament' if they appear as links
		if slug == "books" || slug == "old-testament" || slug == "new-testament" {
			return
		}

		if seen[slug] {
			return
		}

		text := strings.TrimSpace(s.Text())
		if text == "" {
			return
		}

		// Extract Code from Name if present, e.g. "American Standard Version (ASV)"
		name := text
		code := ""
		re := regexp.MustCompile(`^(.*)\s+\(([^)]+)\)$`)
		matches := re.FindStringSubmatch(text)
		if len(matches) == 3 {
			name = strings.TrimSpace(matches[1])
			code = matches[2]
		} else {
			if code == "" && len(text) < 10 && strings.ToUpper(text) == text {
				code = text
			}
		}

		versions = append(versions, bible.ProviderVersion{
			Name:     name,
			Value:    slug,
			Code:     code,
			Language: "English", // We are on the English page
		})
		seen[slug] = true
	})

	return versions, nil
}
