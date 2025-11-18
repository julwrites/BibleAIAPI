package config

import (
	"os"
	"testing"
)

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
