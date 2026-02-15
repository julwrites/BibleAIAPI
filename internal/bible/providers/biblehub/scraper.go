package biblehub

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"bible-api-service/internal/bible"
	"bible-api-service/internal/util"

	"github.com/PuerkitoBio/goquery"
)

// Scraper is a client for scraping BibleHub.
type Scraper struct {
	client  *http.Client
	baseURL string
}

// NewScraper creates a new Scraper.
func NewScraper() *Scraper {
	return &Scraper{
		client:  &http.Client{},
		baseURL: "https://biblehub.com",
	}
}

// GetVerse fetches a verse or range of verses from BibleHub.
// GetVerse fetches a verse or range of verses from BibleHub.
func (s *Scraper) GetVerse(book, chapter, verse, version string) (string, error) {
	if version == "" {
		version = "esv"
	}
	version = strings.ToLower(version)
	bookSlug := strings.ToLower(strings.ReplaceAll(book, " ", "_"))

	// Default to full chapter if verse is empty
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
		// Determine verse range for this chapter
		currentStartV := 1
		currentEndV := 999

		if currentChap == startChapterVal {
			currentStartV = startVerse
		}
		if currentChap == endChapter {
			currentEndV = endVerse
		}

		url := fmt.Sprintf("%s/%s/%s/%d.htm", s.baseURL, version, bookSlug, currentChap)

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
		var inRange bool

		doc.Find("p.regular, p.text").Each(func(i int, s *goquery.Selection) {
			s.Contents().Each(func(j int, node *goquery.Selection) {
				if node.HasClass("reftext") {
					vNumStr := strings.TrimSpace(node.Text())
					// Handle "16." -> "16"
					vNumStr = strings.TrimRight(vNumStr, ".")
					vNum, err := strconv.Atoi(vNumStr)
					if err == nil {
						if vNum >= currentStartV && vNum <= currentEndV {
							inRange = true
						} else {
							inRange = false
						}
					}
				}

				if inRange {
					if node.Is(".reftext") || node.Is(".footnote") || node.Is("sup") {
						return
					}

					// Handle text nodes and other elements (like .woc)
					text := ""
					if goquery.NodeName(node) == "#text" {
						text = node.Text()
					} else {
						text = node.Text()
					}

					chapterTextBuilder.WriteString(text)
				}
			})
		})

		if allTextBuilder.Len() > 0 {
			allTextBuilder.WriteString("\n")
		}
		allTextBuilder.WriteString(strings.TrimSpace(chapterTextBuilder.String()))
	}

	return strings.TrimSpace(allTextBuilder.String()), nil
}

// SearchWords searches for a word or phrase and returns a list of relevant verses.
func (s *Scraper) SearchWords(query, version string) ([]bible.SearchResult, error) {
	if version == "" {
		version = "esv"
	}

	searchURL := fmt.Sprintf("%s/search.php?q=%s", s.baseURL, url.QueryEscape(query))

	req, err := http.NewRequest("GET", searchURL, nil)
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
		return nil, fmt.Errorf("failed to search, status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var results []bible.SearchResult

	doc.Find(".result_block, .result_altblock").Each(func(i int, sel *goquery.Selection) {
		titleSel := sel.Find(".result_title a")
		verseRef := titleSel.Text()
		verseURL := titleSel.AttrOr("href", "")

		if strings.HasPrefix(verseURL, "/") {
			verseURL = s.baseURL + verseURL
		}

		textSel := sel.Find(".description")
		text := strings.TrimSpace(textSel.Text())

		if verseRef != "" {
			results = append(results, bible.SearchResult{
				Verse: verseRef,
				Text:  text,
				URL:   verseURL,
			})
		}
	})

	return results, nil
}
