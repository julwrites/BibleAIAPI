package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"bible-api-service/internal/bible"
	"bible-api-service/internal/bible/providers/biblegateway"
	"bible-api-service/internal/bible/providers/biblenow"

	"gopkg.in/yaml.v2"
)

func run(scraper *biblegateway.Scraper, outputPath string) error {
	log.Println("Fetching Bible versions...")
	bgVersions, err := scraper.GetVersions()
	if err != nil {
		return err
	}

	log.Printf("Found %d versions from Bible Gateway. Generating unified list...", len(bgVersions))

	var unifiedVersions []bible.Version

	for _, v := range bgVersions {
		uv := bible.Version{
			Code:     v.Value,
			Name:     v.Name,
			Language: v.Language,
			Providers: map[string]string{
				"biblegateway": v.Value,
				"biblehub":     strings.ToLower(v.Value),
				"biblenow":     biblenow.GetVersionSlug(v.Value),
			},
		}
		unifiedVersions = append(unifiedVersions, uv)
	}

	data, err := yaml.Marshal(unifiedVersions)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return err
	}

	log.Printf("Successfully updated %s with %d versions", outputPath, len(unifiedVersions))
	return nil
}

func main() {
	scraper := biblegateway.NewScraper()
	if err := run(scraper, "configs/versions.yaml"); err != nil {
		log.Fatalf("Failed to update versions: %v", err)
	}
}
