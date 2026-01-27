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
