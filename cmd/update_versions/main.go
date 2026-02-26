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
	// Use VersionConfig struct to match file format, or map to bible.Version
	// The original used bible.Version which has similar fields but maybe not exactly same yaml tags?
	// bible.Version in internal/bible/version.go:
	/*
	type Version struct {
		Code      string            `json:"code" yaml:"code"`
		Name      string            `json:"name" yaml:"name"`
		Language  string            `json:"language" yaml:"language"`
		Providers map[string]string `json:"providers" yaml:"providers"`
	}
	*/
	// So it matches.

	versionMap := make(map[string]*bible.Version)

	// 1. Read existing config first to preserve data
	existingData, err := os.ReadFile(outputPath)
	if err == nil && len(existingData) > 0 {
		var existingVersions []bible.Version
		if err := yaml.Unmarshal(existingData, &existingVersions); err == nil {
			log.Printf("Loaded %d existing versions from %s", len(existingVersions), outputPath)
			for i := range existingVersions {
				v := &existingVersions[i]
				code := strings.ToUpper(v.Code)
				versionMap[code] = v
			}
		} else {
			log.Printf("Warning: Failed to parse existing config: %v", err)
		}
	} else {
		log.Printf("Warning: Could not read existing config (or empty): %v", err)
	}

	// 2. Fetch from providers
	for pName, provider := range providers {
		log.Printf("Fetching versions from %s...", pName)
		pVersions, err := provider.GetVersions()
		if err != nil {
			log.Printf("Error fetching versions from %s: %v", pName, err)
			// Continue with other providers instead of failing
			continue
		}
		log.Printf("Found %d versions from %s", len(pVersions), pName)

		for _, v := range pVersions {
			code := strings.ToUpper(v.Code)
			if code == "" {
				continue
			}

			if _, exists := versionMap[code]; !exists {
				versionMap[code] = &bible.Version{
					Code:      v.Code, // Keep original casing if new
					Name:      v.Name,
					Language:  v.Language,
					Providers: make(map[string]string),
				}
			}

			// Update provider mapping
			if versionMap[code].Providers == nil {
				versionMap[code].Providers = make(map[string]string)
			}
			versionMap[code].Providers[pName] = v.Value

			// If current provider is biblegateway, update metadata as it tends to be more accurate (e.g. language)
			// But careful not to overwrite good data with bad data if keys exist
			if pName == "biblegateway" {
				if v.Name != "" {
					versionMap[code].Name = v.Name
				}
				if v.Language != "" && v.Language != "Unknown" {
					versionMap[code].Language = v.Language
				}
			} else {
				// Only update if missing or if existing is default/unknown
				if versionMap[code].Language == "" || versionMap[code].Language == "English" || versionMap[code].Language == "Unknown" {
					if v.Language != "English" && v.Language != "" {
						versionMap[code].Language = v.Language
					}
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

	// Sort by Code (Code is usually uppercase or standard)
	// Original sorted by Language then Code. I found sorting by Code easier for diffs.
	// But let's stick to Code for deterministic output.
	sort.Slice(unifiedVersions, func(i, j int) bool {
		return strings.ToUpper(unifiedVersions[i].Code) < strings.ToUpper(unifiedVersions[j].Code)
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
	// Initialize providers with scrapers
	// Note: Scrapers might need configuration (e.g. User-Agent, timeouts) which are default in NewScraper
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
