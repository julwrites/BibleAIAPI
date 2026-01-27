package bible

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewVersionManager(t *testing.T) {
	// Setup Temp Config File
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "versions.yaml")
	content := []byte(`
- code: KJV
  name: King James Version
  language: English
  providers:
    biblegateway: KJV
    biblehub: kjv
    biblenow: king-james-version
- code: ESV
  name: English Standard Version
  language: English
  providers:
    biblegateway: ESV
    biblehub: esv
    biblenow: english-standard-version
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Test Success
	vm, err := NewVersionManager(configPath)
	if err != nil {
		t.Fatalf("NewVersionManager failed: %v", err)
	}

	if len(vm.GetAll()) != 2 {
		t.Errorf("expected 2 versions, got %d", len(vm.GetAll()))
	}
}

func TestVersionManager_GetProviderCode(t *testing.T) {
	// Setup Temp Config File
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "versions.yaml")
	content := []byte(`
- code: TEST
  name: Test Version
  language: TestLang
  providers:
    biblegateway: BG-TEST
    biblehub:
    # biblenow missing
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	vm, err := NewVersionManager(configPath)
	if err != nil {
		t.Fatalf("NewVersionManager failed: %v", err)
	}

	// Case 1: Existing Mapping
	code, err := vm.GetProviderCode("TEST", "biblegateway")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if code != "BG-TEST" {
		t.Errorf("expected 'BG-TEST', got '%s'", code)
	}

	// Case 2: Empty Mapping -> Fallback to Unified Code
	// "biblehub" is explicitly empty string in YAML
	code, err = vm.GetProviderCode("TEST", "biblehub")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if code != "TEST" {
		t.Errorf("expected fallback to 'TEST', got '%s'", code)
	}

	// Case 3: Missing Mapping -> Fallback to Unified Code
	// "biblenow" is missing from providers map
	code, err = vm.GetProviderCode("TEST", "biblenow")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if code != "TEST" {
		t.Errorf("expected fallback to 'TEST', got '%s'", code)
	}

	// Case 4: Unknown Version -> Fallback to Unified Code
	code, err = vm.GetProviderCode("UNKNOWN", "anyprovider")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if code != "UNKNOWN" {
		t.Errorf("expected fallback to 'UNKNOWN', got '%s'", code)
	}
}

func TestNewVersionManager_FileNotFound(t *testing.T) {
	_, err := NewVersionManager("non_existent_file.yaml")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestNewVersionManager_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	_, err := NewVersionManager(configPath)
	if err == nil {
		t.Error("expected error for invalid yaml, got nil")
	}
}

func TestVersionManager_SelectProvider(t *testing.T) {
	// Setup Temp Config File
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "versions.yaml")
	content := []byte(`
- code: KJV
  name: King James Version
  language: English
  providers:
    biblegateway: KJV
    biblehub: kjv
    biblenow: king-james-version
- code: ESV
  name: English Standard Version
  language: English
  providers:
    # biblegateway missing
    biblehub: esv
    biblenow: english-standard-version
- code: ONLY-NOW
  name: Only Now
  language: English
  providers:
    biblenow: only-now
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	vm, err := NewVersionManager(configPath)
	if err != nil {
		t.Fatalf("NewVersionManager failed: %v", err)
	}

	tests := []struct {
		name               string
		unifiedCode        string
		preferredProviders []string
		wantProvider       string
		wantCode           string
		wantErr            bool
	}{
		{
			name:               "KJV default priority (Gateway)",
			unifiedCode:        "KJV",
			preferredProviders: nil, // Should default to [biblegateway, biblehub, biblenow]
			wantProvider:       "biblegateway",
			wantCode:           "KJV",
			wantErr:            false,
		},
		{
			name:               "ESV default priority (Hub fallback)",
			unifiedCode:        "ESV",
			preferredProviders: nil,
			wantProvider:       "biblehub",
			wantCode:           "esv",
			wantErr:            false,
		},
		{
			name:               "ONLY-NOW default priority (Now fallback)",
			unifiedCode:        "ONLY-NOW",
			preferredProviders: nil,
			wantProvider:       "biblenow",
			wantCode:           "only-now",
			wantErr:            false,
		},
		{
			name:               "Explicit priority (Now > Gateway)",
			unifiedCode:        "KJV",
			preferredProviders: []string{"biblenow", "biblegateway"},
			wantProvider:       "biblenow",
			wantCode:           "king-james-version",
			wantErr:            false,
		},
		{
			name:               "Unknown version",
			unifiedCode:        "UNKNOWN",
			preferredProviders: nil,
			wantProvider:       "",
			wantCode:           "",
			wantErr:            true,
		},
		{
			name:               "No suitable provider found",
			unifiedCode:        "ONLY-NOW",
			preferredProviders: []string{"biblegateway", "biblehub"}, // biblenow excluded
			wantProvider:       "",
			wantCode:           "",
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, code, err := vm.SelectProvider(tt.unifiedCode, tt.preferredProviders)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if provider != tt.wantProvider {
				t.Errorf("SelectProvider() provider = %v, want %v", provider, tt.wantProvider)
			}
			if code != tt.wantCode {
				t.Errorf("SelectProvider() code = %v, want %v", code, tt.wantCode)
			}
		})
	}
}
