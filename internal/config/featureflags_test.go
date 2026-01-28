package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetPollingInterval(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		expected time.Duration
	}{
		{
			name:     "Default value",
			envVal:   "",
			expected: 300 * time.Second,
		},
		{
			name:     "Valid duration",
			envVal:   "10m",
			expected: 10 * time.Minute,
		},
		{
			name:     "Invalid duration",
			envVal:   "invalid",
			expected: 300 * time.Second,
		},
		{
			name:     "Short duration",
			envVal:   "1s",
			expected: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				os.Setenv("FEATURE_FLAG_POLLING_INTERVAL", tt.envVal)
			} else {
				os.Unsetenv("FEATURE_FLAG_POLLING_INTERVAL")
			}
			// Ensure cleanup
			defer os.Unsetenv("FEATURE_FLAG_POLLING_INTERVAL")

			assert.Equal(t, tt.expected, getPollingInterval())
		})
	}
}

func TestInitFeatureFlags(t *testing.T) {
	// Create a temporary flags.yaml file
	dir := t.TempDir()
	flagsFile := filepath.Join(dir, "flags.yaml")
	if err := os.WriteFile(flagsFile, []byte("test-flag:\n  variations:\n    true_var: true\n    false_var: false\n  defaultRule:\n    variation: true_var\n"), 0644); err != nil {
		t.Fatalf("Failed to create temp flags file: %v", err)
	}

	t.Run("Successful initialization", func(t *testing.T) {
		os.Setenv("FLAGS_CONFIG_PATH", flagsFile)
		defer os.Unsetenv("FLAGS_CONFIG_PATH")

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("InitFeatureFlags panicked: %v", r)
			}
		}()
		InitFeatureFlags()
	})

	t.Run("No GITHUB_TOKEN", func(t *testing.T) {
		os.Setenv("FLAGS_CONFIG_PATH", flagsFile)
		defer os.Unsetenv("FLAGS_CONFIG_PATH")
		os.Unsetenv("GITHUB_TOKEN")
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("InitFeatureFlags panicked: %v", r)
			}
		}()
		InitFeatureFlags()
	})
}
