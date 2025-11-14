package biblegateway

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseURL        = "https://classic.biblegateway.com"
	passageURL     = baseURL + "/passage/?search=%s&version=%s&interface=print"
	quickSearchURL = baseURL + "/quicksearch/?quicksearch=%s&version=%s&interface=print"
)

type Verse struct {
	Text string `json:"text"`
}

type SearchResult struct {
	Verse string `json:"verse"`
	URL   string `json:"url"`
}

func GetVerse(reference, version string) (*Verse, error) {
	url := fmt.Sprintf(passageURL, reference, version)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch verse, status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var verseText strings.Builder
	doc.Find(".passage-text .text").Each(func(i int, s *goquery.Selection) {
		verseText.WriteString(s.Text())
	})

	return &Verse{Text: strings.TrimSpace(verseText.String())}, nil
}

func SearchWords(query, version string) ([]SearchResult, error) {
	url := fmt.Sprintf(quickSearchURL, query, version)
	res, err := http.Get(url)
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

	var results []SearchResult
	doc.Find(".search-result-list .search-result").Each(func(i int, s *goquery.Selection) {
		verse := s.Find(".bible-item-title a").Text()
		url, _ := s.Find(".bible-item-title a").Attr("href")
		results = append(results, SearchResult{
			Verse: verse,
			URL:   baseURL + url,
		})
	})

	return results, nil
}
