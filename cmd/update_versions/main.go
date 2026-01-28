package main

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"bible-api-service/internal/bible"
	"bible-api-service/internal/bible/providers/biblecom"
	"bible-api-service/internal/bible/providers/biblegateway"
	"bible-api-service/internal/bible/providers/biblehub"
	"bible-api-service/internal/bible/providers/biblenow"

	"gopkg.in/yaml.v2"
)

func run(providers map[string]bible.Provider, outputPath string) error {
	log.Println("Fetching Bible versions from all providers...")

	// Unified map: Code -> Version
	versionMap := make(map[string]*bible.Version)

	for pName, provider := range providers {
		log.Printf("Fetching versions from %s...", pName)
		pVersions, err := provider.GetVersions()
		if err != nil {
			log.Printf("Error fetching versions from %s: %v", pName, err)
			return err
		}
		log.Printf("Found %d versions from %s", len(pVersions), pName)

		for _, v := range pVersions {
			code := strings.ToUpper(v.Code)
			if code == "" {
				continue
			}

			if _, exists := versionMap[code]; !exists {
				versionMap[code] = &bible.Version{
					Code:      code,
					Name:      v.Name,
					Language:  v.Language,
					Providers: make(map[string]string),
				}
			}

			// Update provider mapping
			versionMap[code].Providers[pName] = v.Value

			// If current provider is biblegateway, update metadata as it tends to be more accurate (e.g. language)
			if pName == "biblegateway" {
				versionMap[code].Name = v.Name
				versionMap[code].Language = v.Language
			} else if versionMap[code].Language == "" || versionMap[code].Language == "English" {
				// If we don't have language yet or it's just default "English", maybe update it
				if v.Language != "English" && v.Language != "" {
					versionMap[code].Language = v.Language
				}
				if versionMap[code].Name == "" {
					versionMap[code].Name = v.Name
				}
			}
		}
	}

	// Convert map to slice
	var unifiedVersions []bible.Version
	for _, v := range versionMap {
		unifiedVersions = append(unifiedVersions, *v)
	}

	// Sort by Language then Code to match existing format roughly
	sort.Slice(unifiedVersions, func(i, j int) bool {
		if unifiedVersions[i].Language != unifiedVersions[j].Language {
			return unifiedVersions[i].Language < unifiedVersions[j].Language
		}
		return unifiedVersions[i].Code < unifiedVersions[j].Code
	})

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
	providers := map[string]bible.Provider{
		"biblegateway": biblegateway.NewScraper(),
		"biblehub":     biblehub.NewScraper(),
		"biblenow":     biblenow.NewScraper(),
		"biblecom":     biblecom.NewScraper(),
	}

	if err := run(providers, "configs/versions.yaml"); err != nil {
		log.Fatalf("Failed to update versions: %v", err)
	}
}
