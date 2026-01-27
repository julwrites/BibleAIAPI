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
