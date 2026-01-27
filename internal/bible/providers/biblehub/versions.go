package biblehub

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"bible-api-service/internal/bible"

	"github.com/PuerkitoBio/goquery"
)

// GetVersions scrapes the available Bible versions from BibleHub.
// It uses the Genesis 1:1 page which lists many translations.
func (s *Scraper) GetVersions() ([]bible.ProviderVersion, error) {
	// Fetch Genesis 1:1 page to find version links
	url := s.baseURL + "/genesis/1-1.htm"
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

	// Regex to match version links: /{code}/genesis/1.htm
	// Note: hrefs might be relative or absolute.
	re := regexp.MustCompile(`^/?([^/]+)/genesis/1\.htm$`)

	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists {
			return
		}

		// Normalize href
		href = strings.TrimPrefix(href, s.baseURL)

		matches := re.FindStringSubmatch(href)
		if len(matches) == 2 {
			code := matches[1]

			// But check for "study", "commentaries" etc just in case they mimic this pattern.
			// Actually "study" usually has /study/...
			// "text" -> /text/...
			if code == "study" || code == "commentaries" || code == "text" || code == "context" || code == "audio" {
				return
			}

			if seen[code] {
				return
			}

			name := strings.TrimSpace(sel.Text())
			if name == "" {
				return
			}

			versions = append(versions, bible.ProviderVersion{
				Name:     name,
				Value:    code,
				Code:     strings.ToUpper(code), // Unified code guess
				Language: "English", // Default assumption, maybe incorrect for some
			})
			seen[code] = true
		}
	})

	return versions, nil
}
