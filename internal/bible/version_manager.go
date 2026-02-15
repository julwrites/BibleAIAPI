package bible

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

// VersionManager manages Bible versions and their provider mappings.
type VersionManager struct {
	versions []Version
	byCode   map[string]Version
	mu       sync.RWMutex
}

// NewVersionManager creates a new VersionManager by loading versions from the config file.
func NewVersionManager(path string) (*VersionManager, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read versions config: %w", err)
	}

	var versions []Version
	if err := yaml.Unmarshal(data, &versions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal versions config: %w", err)
	}

	vm := &VersionManager{
		versions: versions,
		byCode:   make(map[string]Version),
	}

	for _, v := range versions {
		vm.byCode[strings.ToUpper(v.Code)] = v
	}

	return vm, nil
}

// GetAll returns all available versions.
func (vm *VersionManager) GetAll() []Version {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	// Return a copy to avoid mutation?
	// For now, slice of structs is copy-ish (structs are copied if accessed by value, but slice backing array is shared)
	// Given we only read, it's fine.
	return vm.versions
}

// GetProviderCode returns the provider-specific code for a given unified version code.
func (vm *VersionManager) GetProviderCode(unifiedCode, provider string) (string, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if unifiedCode == "" {
		return "", nil // Or default?
	}

	v, ok := vm.byCode[strings.ToUpper(unifiedCode)]
	if !ok {
		// If version is not found in our config, assume it's valid and pass it through?
		// Or strictly enforce?
		// Let's pass it through for flexibility, maybe log a warning?
		return unifiedCode, nil
	}

	if code, ok := v.Providers[provider]; ok && code != "" {
		return code, nil
	}

	// Fallback to unified code if provider mapping is missing
	return unifiedCode, nil
}

// SelectProvider resolves the best provider and provider-specific code for a unified version code.
// It iterates through the preferredProviders list and selects the first one that supports the version.
// If preferredProviders is nil or empty, it defaults to ["biblegateway", "biblehub", "biblenow"].
func (vm *VersionManager) SelectProvider(unifiedCode string, preferredProviders []string) (string, string, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if len(preferredProviders) == 0 {
		preferredProviders = []string{"biblegateway", "biblehub", "biblenow"}
	}

	v, ok := vm.byCode[strings.ToUpper(unifiedCode)]
	if !ok {
		return "", "", fmt.Errorf("version not found: %s", unifiedCode)
	}

	for _, provider := range preferredProviders {
		if code, ok := v.Providers[provider]; ok && code != "" {
			return provider, code, nil
		}
	}

	return "", "", fmt.Errorf("no suitable provider found for version: %s", unifiedCode)
}


// GetPrioritizedProviders returns a list of providers that support the given version,
// prioritized by the preferredProviders list (or default order).
func (vm *VersionManager) GetPrioritizedProviders(unifiedCode string, preferredProviders []string) ([]ProviderConfig, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if len(preferredProviders) == 0 {
		preferredProviders = []string{"biblegateway", "biblehub", "biblenow", "biblecom"}
	}

	v, ok := vm.byCode[strings.ToUpper(unifiedCode)]
	if !ok {
		return nil, fmt.Errorf("version not found: %s", unifiedCode)
	}

	var configs []ProviderConfig
	for _, provider := range preferredProviders {
		if code, ok := v.Providers[provider]; ok && code != "" {
			configs = append(configs, ProviderConfig{
				Name:        provider,
				VersionCode: code,
			})
		}
	}

	if len(configs) == 0 {
		// Fallback: if no providers explicitly listed, maybe try to match any available?
		// But for now, if it's in the map it should be found.
		return nil, fmt.Errorf("no providers available for version: %s", unifiedCode)
	}

	return configs, nil
}
