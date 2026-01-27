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

// GetVerse fetches a verse or range of verses from BibleNow.
func (s *Scraper) GetVerse(book, chapter, verse, version string) (string, error) {
	if version == "" {
		version = "KJV"
	}
	versionSlug := GetVersionSlug(version)

	bookSlug, testament, err := getBookSlugAndTestament(book)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/en/bible/%s/%s/%s/%s", s.baseURL, versionSlug, testament, bookSlug, chapter)

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

	// Find all verses in the chapter content
	doc.Find("div.chapter-content a.list-group-item p.verse").Each(func(i int, sel *goquery.Selection) {
		verseNumSpan := sel.Find("span")
		verseNumStr := strings.TrimSpace(verseNumSpan.Text())
		verseNum, err := strconv.Atoi(verseNumStr)
		if err != nil {
			return
		}

		if verseNum >= startVerse && verseNum <= endVerse {
			// Get text, excluding the span (verse number)
			// We can clone the selection or remove the span from the selection
			// But modifying the selection might be tricky.
			// Easier: Get the full text and remove the verse number prefix if it matches
			// Or iterate over child nodes.

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
		// Or return default KJV if unknown?
		// Let's assume KJV as fallback or just lowercase it
		return strings.ToLower(strings.ReplaceAll(version, " ", "-"))
	}
}

func getBookSlugAndTestament(book string) (string, string, error) {
	book = strings.ToLower(strings.TrimSpace(book))

	// Map of book names (or common variations) to slug and testament
	// This list needs to be comprehensive

	// Normalize book name (remove spaces to keys to handle variations better? no, slugs are standard)
	// We'll use a map lookup.

	type info struct {
		slug      string
		testament string
	}

	m := map[string]info{
		"genesis":         {"genesis", "old-testament"},
		"exodus":          {"exodus", "old-testament"},
		"leviticus":       {"leviticus", "old-testament"},
		"numbers":         {"numbers", "old-testament"},
		"deuteronomy":     {"deuteronomy", "old-testament"},
		"joshua":          {"joshua", "old-testament"},
		"judges":          {"judges", "old-testament"},
		"ruth":            {"ruth", "old-testament"},
		"1 samuel":        {"1-samuel", "old-testament"},
		"2 samuel":        {"2-samuel", "old-testament"},
		"1 kings":         {"1-kings", "old-testament"},
		"2 kings":         {"2-kings", "old-testament"},
		"1 chronicles":    {"1-chronicles", "old-testament"},
		"2 chronicles":    {"2-chronicles", "old-testament"},
		"ezra":            {"ezra", "old-testament"},
		"nehemiah":        {"nehemiah", "old-testament"},
		"esther":          {"esther", "old-testament"},
		"job":             {"job", "old-testament"},
		"psalms":          {"psalms", "old-testament"},
		"proverbs":        {"proverbs", "old-testament"},
		"ecclesiastes":    {"ecclesiastes", "old-testament"},
		"song of solomon": {"song-of-solomon", "old-testament"},
		"isaiah":          {"isaiah", "old-testament"},
		"jeremiah":        {"jeremiah", "old-testament"},
		"lamentations":    {"lamentations", "old-testament"},
		"ezekiel":         {"ezekiel", "old-testament"},
		"daniel":          {"daniel", "old-testament"},
		"hosea":           {"hosea", "old-testament"},
		"joel":            {"joel", "old-testament"},
		"amos":            {"amos", "old-testament"},
		"obadiah":         {"obadiah", "old-testament"},
		"jonah":           {"jonah", "old-testament"},
		"micah":           {"micah", "old-testament"},
		"nahum":           {"nahum", "old-testament"},
		"habakkuk":        {"habakkuk", "old-testament"},
		"zephaniah":       {"zephaniah", "old-testament"},
		"haggai":          {"haggai", "old-testament"},
		"zechariah":       {"zechariah", "old-testament"},
		"malachi":         {"malachi", "old-testament"},

		"matthew":         {"matthew", "new-testament"},
		"mark":            {"mark", "new-testament"},
		"luke":            {"luke", "new-testament"},
		"john":            {"john", "new-testament"},
		"acts":            {"acts", "new-testament"},
		"romans":          {"romans", "new-testament"},
		"1 corinthians":   {"1-corinthians", "new-testament"},
		"2 corinthians":   {"2-corinthians", "new-testament"},
		"galatians":       {"galatians", "new-testament"},
		"ephesians":       {"ephesians", "new-testament"},
		"philippians":     {"philippians", "new-testament"},
		"colossians":      {"colossians", "new-testament"},
		"1 thessalonians": {"1-thessalonians", "new-testament"},
		"2 thessalonians": {"2-thessalonians", "new-testament"},
		"1 timothy":       {"1-timothy", "new-testament"},
		"2 timothy":       {"2-timothy", "new-testament"},
		"titus":           {"titus", "new-testament"},
		"philemon":        {"philemon", "new-testament"},
		"hebrews":         {"hebrews", "new-testament"},
		"james":           {"james", "new-testament"},
		"1 peter":         {"1-peter", "new-testament"},
		"2 peter":         {"2-peter", "new-testament"},
		"1 john":          {"1-john", "new-testament"},
		"2 john":          {"2-john", "new-testament"},
		"3 john":          {"3-john", "new-testament"},
		"jude":            {"jude", "new-testament"},
		"revelation":      {"revelation", "new-testament"},
	}

	if val, ok := m[book]; ok {
		return val.slug, val.testament, nil
	}

	return "", "", fmt.Errorf("unknown book: %s", book)
}
