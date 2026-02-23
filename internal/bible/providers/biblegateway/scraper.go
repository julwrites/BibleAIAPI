package biblegateway

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"bible-api-service/internal/bible"
	"bible-api-service/internal/util"

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

func unwrapWoj(s *goquery.Selection) {
	s.Find(".woj").Each(func(i int, sel *goquery.Selection) {
		sel.ReplaceWithHtml(sel.Text())
	})
}

func removeAllAttributes(s *goquery.Selection) {
	s.Find("*").Each(func(_ int, sel *goquery.Selection) {
		sel.Get(0).Attr = []html.Attribute{}
	})
}

func unwrapRedundantSpans(s *goquery.Selection) {
	for {
		unwrapped := false
		s.Find("span").Each(func(i int, sel *goquery.Selection) {
			isRedundant := true
			sel.Contents().EachWithBreak(func(j int, content *goquery.Selection) bool {
				if goquery.NodeName(content) == "#text" {
					if strings.TrimSpace(content.Text()) != "" {
						isRedundant = false
						return false
					}
				} else if !content.Is("span") {
					isRedundant = false
					return false
				}
				return true
			})

			if isRedundant && sel.Children().Length() > 0 {
				html, _ := sel.Html()
				sel.ReplaceWithHtml(html)
				unwrapped = true
			}
		})

		if !unwrapped {
			break
		}
	}
}

func removeEmptyParagraphs(s *goquery.Selection) {
	s.Find("p").Each(func(i int, sel *goquery.Selection) {
		if strings.TrimSpace(sel.Text()) == "" && sel.Find("br").Length() == 0 {
			sel.Remove()
		}
	})
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

// GetVerse fetches a single Bible verse by reference and returns it as sanitized HTML.
func (s *Scraper) GetVerse(book, chapter, verse, version string) (string, error) {
	// Parse verse range
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

	// If cross-chapter range, iterate through chapters
	if startChapterVal != endChapter {
		if endChapter < startChapterVal {
			return "", nil
		}
		numChapters := endChapter - startChapterVal + 1
		chapterTexts := make([]string, numChapters)
		errChan := make(chan error, numChapters)
		var wg sync.WaitGroup

		for i := 0; i < numChapters; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				currentChap := startChapterVal + i
				currentStartV := 1
				currentEndV := 999

				if currentChap == startChapterVal {
					currentStartV = startVerse
				}
				if currentChap == endChapter {
					currentEndV = endVerse
				}

				text, err := s.getVersesFromChapter(book, strconv.Itoa(currentChap), currentStartV, currentEndV, version)
				if err != nil {
					errChan <- fmt.Errorf("failed to fetch chapter %d: %v", currentChap, err)
					return
				}
				chapterTexts[i] = text
			}(i)
		}

		wg.Wait()
		close(errChan)

		if len(errChan) > 0 {
			return "", <-errChan
		}

		var allTextBuilder strings.Builder
		for _, text := range chapterTexts {
			if allTextBuilder.Len() > 0 && text != "" {
				allTextBuilder.WriteString("\n")
			}
			allTextBuilder.WriteString(text)
		}
		return strings.TrimSpace(allTextBuilder.String()), nil
	}

	// Single chapter range (including whole chapter)
	// Use existing logic for efficiency and to preserve existing behavior
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

func (s *Scraper) getVersesFromChapter(book, chapter string, startVerse, endVerse int, version string) (string, error) {
    // Fetch whole chapter
    reference := fmt.Sprintf("%s %s", book, chapter)
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
        return "", fmt.Errorf("failed to fetch chapter, status code: %d", res.StatusCode)
    }

    doc, err := goquery.NewDocumentFromReader(res.Body)
    if err != nil {
        return "", err
    }

    passageSelection := doc.Find(".passage-text")
    if passageSelection.Length() == 0 || strings.Contains(passageSelection.Text(), "No results found") {
        return "", fmt.Errorf("chapter not found")
    }

    // Extract verses within range
    var textBuilder strings.Builder
    // Handle verse 1 if in range (verse 1 uses span.chapternum instead of sup.versenum)
    if startVerse == 1 && 1 <= endVerse {
        passageSelection.Find("span.chapternum").Each(func(i int, chapSel *goquery.Selection) {
            // Verse 1 container is parent span
            verseContainer := chapSel.Parent()
            if verseContainer.Is("span") {
                // Remove cross-references and footnotes before extracting text
                verseContainer.Find("sup.crossreference, sup.footnote").Remove()
                // Remove the chapter number span itself
                chapSel.Remove()
                // Get text
                verseText := strings.TrimSpace(verseContainer.Text())
                if textBuilder.Len() > 0 {
                    textBuilder.WriteString(" ")
                }
                textBuilder.WriteString(verseText)
            }
            // Only process first chapternum (should be only one)
            return
        })
    }
    passageSelection.Find("sup.versenum").Each(func(i int, supSel *goquery.Selection) {
        verseNumText := strings.TrimSpace(supSel.Text())
        verseNumText = strings.TrimRight(verseNumText, "\u00a0")
        verseNumText = strings.TrimSpace(verseNumText)
        verseNum, err := strconv.Atoi(verseNumText)
        if err != nil {
            return
        }
        if verseNum >= startVerse && verseNum <= endVerse {
            // Find the verse container: parent span or p.verse/p.line
            verseContainer := supSel.Parent()
            // If parent is span, maybe its parent is p.verse or p.line
            if verseContainer.Is("span") {
                if verseContainer.Parent().Is("p.verse") || verseContainer.Parent().Is("p.line") {
                    verseContainer = verseContainer.Parent()
                }
            }
            // Remove cross-references and footnotes before extracting text
            verseContainer.Find("sup.crossreference, sup.footnote").Remove()
            // Remove the verse number superscript
            supSel.Remove()
            // Get text excluding the superscript verse number
            verseText := strings.TrimSpace(verseContainer.Text())
            if textBuilder.Len() > 0 {
                textBuilder.WriteString(" ")
            }
            textBuilder.WriteString(verseText)
        }
    })

    result := strings.TrimSpace(textBuilder.String())
    if result == "" {
        return "", fmt.Errorf("no verses found in range %d-%d", startVerse, endVerse)
    }
    return result, nil
}

func sanitizeSelection(s *goquery.Selection) (string, error) {
	isPoetry := s.Find("div.poetry").Length() > 0

	removeUnwantedElements(s)
	unwrapSmallCaps(s)
	unwrapWoj(s)

	if isPoetry {
		s.Find("p.top-1").ReplaceWithHtml("<br/>")
		s.Find("div.poetry, p.line, span.indent-1").Each(func(i int, sel *goquery.Selection) {
			html, _ := sel.Html()
			sel.ReplaceWithHtml(html)
		})
	}

	strictSanitize(s)
	unwrapRedundantSpans(s)
	removeAllAttributes(s)
	removeEmptyParagraphs(s)

	html, err := s.Html()
	if err != nil {
		return "", err
	}

	html = strings.ReplaceAll(html, "\u00a0", " ")

	return strings.TrimSpace(html), nil
}

func sanitizeSnippet(s *goquery.Selection) (string, error) {
	s.Find(".bible-item-extras").Remove()
	s.Find("h1, h2, h3, h4, h5, h6").Remove()
	return sanitizeSelection(s)
}

// SearchWords searches for a word or phrase and returns a list of relevant verses.
func (s *Scraper) SearchWords(query, version string) ([]bible.SearchResult, error) {
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

	results := []bible.SearchResult{}
	selection := doc.Find(".search-result-list .bible-item")
	log.Printf("Found %d search results for query '%s'", selection.Length(), query)

	selection.Each(func(i int, sel *goquery.Selection) {
		titleLink := sel.Find(".bible-item-title")
		verse := titleLink.Text()
		url, _ := titleLink.Attr("href")

		textSel := sel.Find(".bible-item-text")
		text, _ := sanitizeSnippet(textSel)

		results = append(results, bible.SearchResult{
			Verse: verse,
			Text:  text,
			URL:   s.baseURL + url,
		})
	})

	return results, nil
}
