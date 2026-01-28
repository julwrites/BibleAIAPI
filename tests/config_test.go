package tests

import (
	"os"
	"testing"

	"bible-api-service/internal/bible"
	"gopkg.in/yaml.v2"
)

func TestVersionsConfigValidation(t *testing.T) {
	// Locate the versions.yaml file relative to the test file
	// Assuming tests are run from the repo root or tests directory
	// We'll try to find it in "configs/versions.yaml" or "../configs/versions.yaml"

	configPath := "configs/versions.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "../configs/versions.yaml"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatalf("versions.yaml not found at configs/versions.yaml or ../configs/versions.yaml")
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read versions.yaml: %v", err)
	}

	var versions []bible.Version
	if err := yaml.Unmarshal(data, &versions); err != nil {
		t.Fatalf("failed to unmarshal versions.yaml: %v", err)
	}

	knownProviders := map[string]bool{
		"biblegateway": true,
		"biblehub":     true,
		"biblenow":     true,
		"biblecom":     true,
	}

	seenCodes := make(map[string]bool)

	for i, v := range versions {
		if v.Code == "" {
			t.Errorf("version at index %d has empty Code", i)
		}
		if seenCodes[v.Code] {
			t.Errorf("duplicate version code found: %s", v.Code)
		}
		seenCodes[v.Code] = true

		if v.Name == "" {
			t.Errorf("version %s has empty Name", v.Code)
		}
		if v.Language == "" {
			t.Errorf("version %s has empty Language", v.Code)
		}

		if len(v.Providers) == 0 {
			t.Errorf("version %s has no providers", v.Code)
		}

		for providerName, providerCode := range v.Providers {
			if !knownProviders[providerName] {
				t.Errorf("version %s has unknown provider: %s", v.Code, providerName)
			}
			if providerCode == "" {
				t.Errorf("version %s has empty code for provider %s", v.Code, providerName)
			}
		}
	}
}

func TestFlagsConfigValidation(t *testing.T) {
	configPath := "configs/flags.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "../configs/flags.yaml"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// flags.yaml might be optional or generated, but if it exists it should be valid.
			// If it doesn't exist, we can skip or log.
			// Assuming it exists for now based on file listing.
			t.Skipf("flags.yaml not found at configs/flags.yaml or ../configs/flags.yaml")
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read flags.yaml: %v", err)
	}

	// Just check if it's valid YAML map
	var flags map[string]interface{}
	if err := yaml.Unmarshal(data, &flags); err != nil {
		t.Fatalf("failed to unmarshal flags.yaml: %v", err)
	}
}
