package biblecom

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"bible-api-service/internal/bible"

	"github.com/PuerkitoBio/goquery"
)

// Scraper is a client for scraping Bible.com.
type Scraper struct {
	client  *http.Client
	baseURL string
}

// NewScraper creates a new Scraper.
func NewScraper() *Scraper {
	return &Scraper{
		client:  &http.Client{},
		baseURL: "https://www.bible.com",
	}
}

// GetVerse fetches a verse or range of verses from Bible.com.
func (s *Scraper) GetVerse(book, chapter, verse, version string) (string, error) {
	if version == "" {
		version = "111" // Default to NIV ID
	}

	usfmBook, err := mapBookToUSFM(book)
	if err != nil {
		return "", err
	}

	// URL format: https://www.bible.com/bible/{versionID}/{BookUSFM}.{Chapter}
	// Example: https://www.bible.com/bible/111/JHN.3
	url := fmt.Sprintf("%s/bible/%s/%s.%s", s.baseURL, version, usfmBook, chapter)

	req, err := http.NewRequest("GET", url, nil)
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
		return "", fmt.Errorf("failed to fetch chapter, status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
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

	// Iterate through verses in the range
	for v := startVerse; v <= endVerse; v++ {
		// Selector: span[data-usfm='BOOK.CHAPTER.VERSE']
		selector := fmt.Sprintf("span[data-usfm='%s.%s.%d']", usfmBook, chapter, v)
		selection := doc.Find(selector)

		if selection.Length() == 0 {
			// Stop if we can't find the verse (and it's not the first one we are looking for, or maybe we should error if first not found?)
			// If we are looking for a range, and we hit the end of chapter, we stop.
			if v > startVerse {
				break
			}
			// If even the first verse is not found, it might be an issue.
			// However, sometimes verses are merged or formatted differently.
			// But for now, strict check.
			continue
		}

		// Extract text. Usually text is inside span.content or just text nodes.
		// Bible.com structure inside the span might contain formatting tags.
		// We want plain text.

		// The span might contain multiple spans.
		// Example: <span data-usfm="JHN.3.1"><span class="content">For God so loved...</span></span>

		text := strings.TrimSpace(selection.Text())

		// Clean up: sometimes text contains verse numbers if not careful, but data-usfm usually wraps the content.
		// But in view_text_website we saw verse numbers. They might be outside or inside with a specific class.
		// If they are inside, we might capture them.
		// But let's assume simple text extraction first.

		if textBuilder.Len() > 0 {
			textBuilder.WriteString(" ")
		}
		textBuilder.WriteString(text)
	}

	result := textBuilder.String()
	if result == "" {
		return "", fmt.Errorf("verses not found")
	}

	return result, nil
}

// GetVersions fetches the list of available Bible versions from Bible.com.
func (s *Scraper) GetVersions() ([]bible.ProviderVersion, error) {
	url := s.baseURL + "/versions"

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

	// Regex to parse href: /versions/(\d+)-([a-zA-Z0-9]+)-(.*)
	// Example: /versions/111-niv-new-international-version
	re := regexp.MustCompile(`^/versions/(\d+)-([a-zA-Z0-9]+)-(.*)$`)

	doc.Find("a[href^='/versions/']").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		matches := re.FindStringSubmatch(href)
		if len(matches) == 4 {
			id := matches[1]
			code := strings.ToUpper(matches[2])
			// slug := matches[3] // Unused

			// Name usually is the text of the link
			name := strings.TrimSpace(s.Text())
			// Clean up name (sometimes contains (CODE))
			// e.g., "New International Version (NIV)" -> "New International Version"
			name = strings.TrimSuffix(name, fmt.Sprintf(" (%s)", code))

			versions = append(versions, bible.ProviderVersion{
				Name:     name,
				Value:    id,    // Provider specific ID
				Code:     code,  // Unified code (e.g. NIV)
				Language: "English", // We might want to infer language, but default to English or parse from section headers if needed.
			})
		}
	})

	return versions, nil
}

// SearchWords is not supported on Bible.com.
func (s *Scraper) SearchWords(query, version string) ([]bible.SearchResult, error) {
	return nil, fmt.Errorf("search not supported on Bible.com")
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

func mapBookToUSFM(book string) (string, error) {
	book = strings.ToLower(strings.TrimSpace(book))

	// USFM Mapping
	m := map[string]string{
		"genesis":         "GEN",
		"exodus":          "EXO",
		"leviticus":       "LEV",
		"numbers":         "NUM",
		"deuteronomy":     "DEU",
		"joshua":          "JOS",
		"judges":          "JDG",
		"ruth":            "RUT",
		"1 samuel":        "1SA",
		"2 samuel":        "2SA",
		"1 kings":         "1KI",
		"2 kings":         "2KI",
		"1 chronicles":    "1CH",
		"2 chronicles":    "2CH",
		"ezra":            "EZR",
		"nehemiah":        "NEH",
		"esther":          "EST",
		"job":             "JOB",
		"psalms":          "PSA",
		"psalm":           "PSA", // Alias
		"proverbs":        "PRO",
		"ecclesiastes":    "ECC",
		"song of solomon": "SNG",
		"song of songs":   "SNG", // Alias
		"isaiah":          "ISA",
		"jeremiah":        "JER",
		"lamentations":    "LAM",
		"ezekiel":         "EZK",
		"daniel":          "DAN",
		"hosea":           "HOS",
		"joel":            "JOL",
		"amos":            "AMO",
		"obadiah":         "OBA",
		"jonah":           "JON",
		"micah":           "MIC",
		"nahum":           "NAM",
		"habakkuk":        "HAB",
		"zephaniah":       "ZEP",
		"haggai":          "HAG",
		"zechariah":       "ZEC",
		"malachi":         "MAL",

		"matthew":         "MAT",
		"mark":            "MRK",
		"luke":            "LUK",
		"john":            "JHN",
		"acts":            "ACT",
		"romans":          "ROM",
		"1 corinthians":   "1CO",
		"2 corinthians":   "2CO",
		"galatians":       "GAL",
		"ephesians":       "EPH",
		"philippians":     "PHP",
		"colossians":      "COL",
		"1 thessalonians": "1TH",
		"2 thessalonians": "2TH",
		"1 timothy":       "1TI",
		"2 timothy":       "2TI",
		"titus":           "TIT",
		"philemon":        "PHM",
		"hebrews":         "HEB",
		"james":           "JAS",
		"1 peter":         "1PE",
		"2 peter":         "2PE",
		"1 john":          "1JN",
		"2 john":          "2JN",
		"3 john":          "3JN",
		"jude":            "JUD",
		"revelation":      "REV",
	}

	if val, ok := m[book]; ok {
		return val, nil
	}

	return "", fmt.Errorf("unknown book: %s", book)
}
