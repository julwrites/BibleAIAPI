package main

import (
	"log"
	"os"

	"bible-api-service/internal/biblegateway"

	"gopkg.in/yaml.v2"
)

func main() {
	scraper := biblegateway.NewScraper()
	log.Println("Fetching Bible versions...")
	versions, err := scraper.GetVersions()
	if err != nil {
		log.Fatalf("Failed to get versions: %v", err)
	}

	log.Printf("Found %d versions. Writing to configs/versions.yaml...", len(versions))

	data, err := yaml.Marshal(versions)
	if err != nil {
		log.Fatalf("Failed to marshal versions to YAML: %v", err)
	}

	// Ensure configs directory exists
	if err := os.MkdirAll("configs", 0755); err != nil {
		log.Fatalf("Failed to create configs directory: %v", err)
	}

	if err := os.WriteFile("configs/versions.yaml", data, 0644); err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	log.Println("Successfully updated configs/versions.yaml")
}
