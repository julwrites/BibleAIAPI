package biblecom

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"bible-api-service/internal/bible"
	"bible-api-service/internal/util"

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

	startVerse := 1
	endVerse := 999
	endChapter := 0

	startChapterVal, err := strconv.Atoi(chapter)
	if err != nil {
		return "", fmt.Errorf("invalid chapter format: %v", err)
	}

	if verse != "" {
		parsed, err := util.ParseVerseRange(verse)
		if err != nil {
			return "", fmt.Errorf("invalid verse range: %v", err)
		}
		startVerse = parsed.StartVerse
		endVerse = parsed.EndVerse
		if parsed.IsCrossChapter {
			endChapter = parsed.EndChapter
		} else {
			endChapter = startChapterVal
		}
	} else {
		endChapter = startChapterVal
	}

	var allTextBuilder strings.Builder

	for currentChap := startChapterVal; currentChap <= endChapter; currentChap++ {
		currentStartV := 1
		currentEndV := 999

		if currentChap == startChapterVal {
			currentStartV = startVerse
		}
		if currentChap == endChapter {
			currentEndV = endVerse
		}

		// URL format: https://www.bible.com/bible/{versionID}/{BookUSFM}.{Chapter}
		url := fmt.Sprintf("%s/bible/%s/%s.%d", s.baseURL, version, usfmBook, currentChap)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BibleAIAPI/1.0)")

		res, err := s.client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to fetch chapter %d: %v", currentChap, err)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			return "", fmt.Errorf("failed to fetch chapter %d, status code: %d", currentChap, res.StatusCode)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return "", err
		}

		var chapterTextBuilder strings.Builder

		for v := currentStartV; v <= currentEndV; v++ {
			// Selector: span[data-usfm='BOOK.CHAPTER.VERSE']
			selector := fmt.Sprintf("span[data-usfm='%s.%d.%d']", usfmBook, currentChap, v)
			selection := doc.Find(selector)

			if selection.Length() == 0 {
				if v > startVerse && verse != "" {
					// We might have reached end of chapter.
				}
				// Optimization: break if we've looked far past reasonable verse count with no hits?
				if v > 200 && chapterTextBuilder.Len() == 0 {
					// Stop trying if we haven't found anything by verse 200
					break
				} else if v > 200 && selection.Length() == 0 {
					// Stop trying if we are past verse 200 and stop finding verses
					// (Assuming chapter isn't > 200 verses long, longest is Ps 119 with 176)
					break
				}
				continue
			}

			text := strings.TrimSpace(selection.Text())
			if chapterTextBuilder.Len() > 0 {
				chapterTextBuilder.WriteString(" ")
			}
			chapterTextBuilder.WriteString(text)
		}

		chapterText := chapterTextBuilder.String()

		if allTextBuilder.Len() > 0 && chapterText != "" {
			allTextBuilder.WriteString("\n")
		}
		allTextBuilder.WriteString(chapterText)
	}

	finalResult := strings.TrimSpace(allTextBuilder.String())
	if finalResult == "" {
		return "", fmt.Errorf("verses not found")
	}

	return finalResult, nil
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
				Value:    id,        // Provider specific ID
				Code:     code,      // Unified code (e.g. NIV)
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
