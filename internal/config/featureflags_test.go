package config

import (
	"os"
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
	t.Run("Successful initialization", func(t *testing.T) {
		// We can't easily mock the gofeatureflag.Init function,
		// so we'll just call it and check for panics.
		// In a real-world scenario, you might use a more sophisticated
		// approach, like a wrapper around the gofeatureflag package.
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("InitFeatureFlags panicked: %v", r)
			}
		}()
		InitFeatureFlags()
	})

	t.Run("No GITHUB_TOKEN", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("InitFeatureFlags panicked: %v", r)
			}
		}()
		InitFeatureFlags()
	})
}
