package biblenow

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"bible-api-service/internal/bible"

	"github.com/PuerkitoBio/goquery"
)

// Scraper is a client for scraping BibleNow.net.
type Scraper struct {
	client  *http.Client
	baseURL string
}

// NewScraper creates a new Scraper.
func NewScraper() *Scraper {
	return &Scraper{
		client:  &http.Client{},
		baseURL: "https://biblenow.net",
	}
}

// Standard order of books in Protestant bibles
var standardBookOrder = []string{
	"genesis", "exodus", "leviticus", "numbers", "deuteronomy",
	"joshua", "judges", "ruth", "1 samuel", "2 samuel",
	"1 kings", "2 kings", "1 chronicles", "2 chronicles", "ezra",
	"nehemiah", "esther", "job", "psalms", "proverbs",
	"ecclesiastes", "song of solomon", "isaiah", "jeremiah",
	"lamentations", "ezekiel", "daniel", "hosea", "joel",
	"amos", "obadiah", "jonah", "micah", "nahum",
	"habakkuk", "zephaniah", "haggai", "zechariah", "malachi",
	"matthew", "mark", "luke", "john", "acts",
	"romans", "1 corinthians", "2 corinthians", "galatians", "ephesians",
	"philippians", "colossians", "1 thessalonians", "2 thessalonians",
	"1 timothy", "2 timothy", "titus", "philemon", "hebrews",
	"james", "1 peter", "2 peter", "1 john", "2 john",
	"3 john", "jude", "revelation",
}

// GetVerse fetches a verse or range of verses from BibleNow.
func (s *Scraper) GetVerse(book, chapter, verse, version string) (string, error) {
	if version == "" {
		version = "KJV"
	}

	// Resolve the version path
	versionPath := version
	if !strings.Contains(version, "/") {
		// Legacy behavior: assume English code if no path separator
		slug := GetVersionSlug(version)
		versionPath = "en/bible/" + slug
	}
	// Ensure versionPath does not start with / (relative to baseURL)
	versionPath = strings.TrimPrefix(versionPath, "/")

	// 1. Identify the book index
	bookIndex := -1
	normalizedBook := strings.ToLower(strings.TrimSpace(book))
	for i, b := range standardBookOrder {
		if b == normalizedBook {
			bookIndex = i
			break
		}
	}
	if bookIndex == -1 {
		return "", fmt.Errorf("unknown book: %s", book)
	}

	// 2. Fetch the version page to find the book URL
	versionURL := fmt.Sprintf("%s/%s", s.baseURL, versionPath)

	// Optimization: For English KJV, we can guess the URL to save a request
	// But sticking to the plan for robustness across languages

	req, err := http.NewRequest("GET", versionURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BibleAIAPI/1.0)")

	res, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch version page, status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	// 3. Extract book links
	var bookLinks []string
	seen := make(map[string]bool)

	// Iterate over links and filter for book links
	// Book links are usually deeper than the version path
	// e.g. {versionPath}/{testament}/{bookSlug}

	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists {
			return
		}

		if strings.HasPrefix(href, s.baseURL) {
			href = strings.TrimPrefix(href, s.baseURL)
		}

		// Ensure it starts with /versionPath
		prefix := "/" + versionPath + "/"
		if !strings.HasPrefix(href, prefix) {
			return
		}

		// Check depth
		suffix := strings.TrimPrefix(href, prefix)
		parts := strings.Split(suffix, "/")

		// We expect "{testament}/{book}" -> 2 parts
		if len(parts) != 2 {
			return
		}

		// Filter out if any part is empty
		if parts[0] == "" || parts[1] == "" {
			return
		}

		// Filter out common non-book links
		slug := parts[1]
		if slug == "introduction" || slug == "preface" || slug == "index" || slug == "contents" {
			return
		}

		if seen[href] {
			return
		}

		bookLinks = append(bookLinks, href)
		seen[href] = true
	})

	if bookIndex >= len(bookLinks) {
		return "", fmt.Errorf("book index %d out of range (found %d books)", bookIndex, len(bookLinks))
	}

	bookURLPath := bookLinks[bookIndex]

	// 4. Construct Chapter URL
	chapterURL := fmt.Sprintf("%s%s/%s", s.baseURL, bookURLPath, chapter)

	// 5. Fetch Chapter and scrape verses
	req, err = http.NewRequest("GET", chapterURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BibleAIAPI/1.0)")

	res, err = s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch chapter, status code: %d", res.StatusCode)
	}

	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	startVerse := 1
	endVerse := 999

	if verse != "" {
		start, end, err := parseVerseRange(verse)
		if err != nil {
			return "", fmt.Errorf("invalid verse range: %v", err)
		}
		startVerse = start
		endVerse = end
	}

	var textBuilder strings.Builder

	// Find all verses in the chapter content
	doc.Find("div.chapter-content a.list-group-item p.verse").Each(func(i int, sel *goquery.Selection) {
		verseNumSpan := sel.Find("span")
		verseNumStr := strings.TrimSpace(verseNumSpan.Text())
		verseNum, err := strconv.Atoi(verseNumStr)
		if err != nil {
			return
		}

		if verseNum >= startVerse && verseNum <= endVerse {
			var verseTextBuilder strings.Builder
			sel.Contents().Each(func(j int, node *goquery.Selection) {
				if node.Is("span") {
					return
				}
				verseTextBuilder.WriteString(node.Text())
			})

			text := strings.TrimSpace(verseTextBuilder.String())
			if textBuilder.Len() > 0 {
				textBuilder.WriteString(" ")
			}
			textBuilder.WriteString(text)
		}
	})

	result := textBuilder.String()
	if result == "" {
		return "", fmt.Errorf("verses not found")
	}

	return result, nil
}

// SearchWords is not supported on BibleNow.
func (s *Scraper) SearchWords(query, version string) ([]bible.SearchResult, error) {
	return nil, fmt.Errorf("search not supported on BibleNow")
}

func parseVerseRange(verse string) (int, int, error) {
	parts := strings.Split(verse, "-")
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}

	end := start
	if len(parts) > 1 {
		end, err = strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return 0, 0, err
		}
	}

	return start, end, nil
}

// GetVersionSlug returns the slug for a given version code.
func GetVersionSlug(version string) string {
	v := strings.ToUpper(version)
	switch v {
	case "KJV":
		return "king-james-version"
	case "ESV":
		return "english-standard-version"
	case "NIV":
		return "new-international-version"
	case "NKJV":
		return "new-king-james-version"
	case "ASV":
		return "american-standard-version"
	case "NASB":
		return "new-american-standard-bible"
	case "NLT":
		return "new-living-translation"
	default:
		// Try to slugify the version if it's not a known code
		return strings.ToLower(strings.ReplaceAll(version, " ", "-"))
	}
}

