package biblegateway

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Version represents a Bible version.
type Version struct {
	Name     string `json:"name" yaml:"name"`
	Value    string `json:"value" yaml:"value"`
	Language string `json:"language" yaml:"language"`
}

// GetVersions scrapes the available Bible versions from Bible Gateway.
func (s *Scraper) GetVersions() ([]Version, error) {
	url := s.baseURL + "/versions/"
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
		return nil, fmt.Errorf("failed to fetch versions, status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var versions []Version
	// Find the select element with class "search-dropdown" and name "version"
	sel := doc.Find("select.search-dropdown[name='version']")

	var currentLanguage string = "Unknown"

	// Iterate over all options directly
	sel.Find("option").Each(func(i int, opt *goquery.Selection) {
		val := opt.AttrOr("value", "")
		text := strings.TrimSpace(opt.Text())

		// Check if this option acts as a language header
		// Format seen: "---Language Name (Code)---"
		if strings.HasPrefix(text, "---") && strings.HasSuffix(text, "---") {
			currentLanguage = strings.TrimSuffix(strings.TrimPrefix(text, "---"), "---")
			return // Skip adding this as a version
		}

		// Skip empty options or non-breaking spaces
		if val == "" || text == "" || text == "\u00a0" {
			return
		}

		versions = append(versions, Version{
			Name:     text,
			Value:    val,
			Language: currentLanguage,
		})
	})

	return versions, nil
}
