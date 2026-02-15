package biblenow

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"bible-api-service/internal/bible"

	"github.com/PuerkitoBio/goquery"
)

var (
	// languageRegex matches language links: https://biblenow.net/{code} or /{code}
	// Codes are usually 2 or 3 letters, or hyphenated (e.g. zh-Hant)
	languageRegex = regexp.MustCompile(`^/([a-z]{2,3}(-[A-Za-z]+)?)$`)

	// versionTextRegex heuristic: Text must look like "Name (CODE)"
	versionTextRegex = regexp.MustCompile(`^(.*)\s+\(([^)]+)\)$`)
)

// GetVersions scrapes the available Bible versions from BibleNow.
// It iterates over all available languages to find versions.
func (s *Scraper) GetVersions() ([]bible.ProviderVersion, error) {
	// 1. Fetch the English Bible page to find language links
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
		return nil, fmt.Errorf("failed to fetch English page, status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// 2. Extract languages
	languages, err := s.extractLanguages(doc)
	if err != nil {
		return nil, err
	}

	// Always include English manually if not found (it should be found)
	foundEn := false
	for _, l := range languages {
		if l.Code == "en" {
			foundEn = true
			break
		}
	}
	if !foundEn {
		languages = append(languages, languageInfo{Code: "en", Name: "English"})
	}

	// 3. Fetch versions from all languages concurrently
	var versions []bible.ProviderVersion
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit concurrency to 5
	sem := make(chan struct{}, 5)

	for _, lang := range languages {
		wg.Add(1)
		go func(l languageInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			langVersions, err := s.fetchVersionsForLanguage(l)
			if err != nil {
				// Log error but continue
				// log.Printf("Error fetching versions for %s: %v\n", l.Code, err)
				return
			}

			mu.Lock()
			versions = append(versions, langVersions...)
			mu.Unlock()
		}(lang)
	}

	wg.Wait()

	return versions, nil
}

type languageInfo struct {
	Code string
	Name string
}

func (s *Scraper) extractLanguages(doc *goquery.Document) ([]languageInfo, error) {
	var languages []languageInfo
	seen := make(map[string]bool)

	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists {
			return
		}

		if strings.HasPrefix(href, s.baseURL) {
			href = strings.TrimPrefix(href, s.baseURL)
		}

		matches := languageRegex.FindStringSubmatch(href)
		if len(matches) > 1 {
			code := matches[1]

			// Exclude common non-language root paths
			// (e.g. /css, /js, /img, /en/bible/..., /build)
			if code == "news" || code == "books" || code == "css" || code == "js" || code == "img" || code == "build" || code == "login" || code == "register" {
				return
			}

			if seen[code] {
				return
			}

			text := strings.TrimSpace(sel.Text())
			// Clean up text if it contains parens with code, e.g. "Afrikaans (AF)"
			name := text
			if idx := strings.Index(name, "("); idx != -1 {
				name = strings.TrimSpace(name[:idx])
			}
			if name == "" {
				name = code
			}

			languages = append(languages, languageInfo{
				Code: code,
				Name: name,
			})
			seen[code] = true
		}
	})

	return languages, nil
}

func (s *Scraper) fetchVersionsForLanguage(lang languageInfo) ([]bible.ProviderVersion, error) {
	url := fmt.Sprintf("%s/%s", s.baseURL, lang.Code)
	// Special case for English: versions are at /en/bible
	if lang.Code == "en" {
		url = fmt.Sprintf("%s/en/bible", s.baseURL)
	}

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
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var versions []bible.ProviderVersion
	seen := make(map[string]bool)

	// Look for version links.
	// Pattern: /{code}/{something}/{slug}
	// e.g. /es/biblia/reina-valera-1909

	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists {
			return
		}

		if strings.HasPrefix(href, s.baseURL) {
			href = strings.TrimPrefix(href, s.baseURL)
		}

		// Ensure it starts with /{lang.Code}/
		prefix := fmt.Sprintf("/%s/", lang.Code)
		if !strings.HasPrefix(href, prefix) {
			return
		}

		// Trim prefix
		path := strings.TrimPrefix(href, prefix)
		// Expected structure: {word}/{slug}
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return
		}

		slug := parts[1]
		// Exclude deeper links or empty slugs
		if slug == "" {
			return
		}

		// Filter out common non-version paths if any
		if slug == "antiguo-testamento" || slug == "nuevo-testamento" || slug == "old-testament" || slug == "new-testament" {
			return
		}

		fullPath := strings.TrimPrefix(href, "/") // store as "es/biblia/reina-valera-1909"

		if seen[fullPath] {
			return
		}

		text := strings.TrimSpace(sel.Text())
		if text == "" {
			return
		}

		// Heuristic: Text must look like "Name (CODE)"
		matches := versionTextRegex.FindStringSubmatch(text)

		name := text
		code := ""

		if len(matches) == 3 {
			name = strings.TrimSpace(matches[1])
			code = matches[2]
		} else {
			// Derive code from slug if not found
			code = slug
		}

		versions = append(versions, bible.ProviderVersion{
			Name:     name,
			Value:    fullPath, // Storing full path for GetVerse
			Code:     code,
			Language: lang.Name,
		})
		seen[fullPath] = true
	})

	return versions, nil
}
