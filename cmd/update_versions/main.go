package main

import (
	"log"
	"os"
	"path/filepath"

	"bible-api-service/internal/bible/providers/biblegateway"

	"gopkg.in/yaml.v2"
)

func run(scraper *biblegateway.Scraper, outputPath string) error {
	log.Println("Fetching Bible versions...")
	versions, err := scraper.GetVersions()
	if err != nil {
		return err
	}

	log.Printf("Found %d versions. Writing to %s...", len(versions), outputPath)

	data, err := yaml.Marshal(versions)
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

	log.Printf("Successfully updated %s", outputPath)
	return nil
}

func main() {
	scraper := biblegateway.NewScraper()
	if err := run(scraper, "configs/versions.yaml"); err != nil {
		log.Fatalf("Failed to update versions: %v", err)
	}
}
