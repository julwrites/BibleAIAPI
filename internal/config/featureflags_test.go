package config

import (
	"os"
	"testing"

	gofeatureflag "github.com/thomaspoignant/go-feature-flag"
)

func TestInitFeatureFlags(t *testing.T) {
	t.Run("Successful initialization", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "test")
		defer os.Unsetenv("GITHUB_TOKEN")

		// We can't easily mock the gofeatureflag.Init function,
		// so we'll just call it and check for panics.
		// In a real-world scenario, you might use a more sophisticated
		// approach, like a wrapper around the gofeatureflag package.
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("InitFeatureFlags panicked: %v", r)
			}
		}()

		// Create a dummy flags.yaml file
		if err := os.MkdirAll("configs", 0755); err != nil {
			t.Fatalf("could not create configs directory: %v", err)
		}
		defer os.RemoveAll("configs")

		file, err := os.Create("configs/flags.yaml")
		if err != nil {
			t.Fatalf("could not create dummy flags.yaml: %v", err)
		}
		defer file.Close()

		_, err = file.WriteString(`
test-flag:
  variations:
    true: true
    false: false
  defaultRule:
    variation: true
`)
		if err != nil {
			t.Fatalf("could not write to dummy flags.yaml: %v", err)
		}

		InitFeatureFlags()
		gofeatureflag.Close()
	})
}
