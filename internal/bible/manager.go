package bible

import (
	"fmt"
)

// ProviderManager manages multiple Bible providers.
type ProviderManager struct {
	primary Provider
}

// NewProviderManager creates a new ProviderManager with the given primary provider.
func NewProviderManager(primary Provider) *ProviderManager {
	return &ProviderManager{
		primary: primary,
	}
}

// GetVerse fetches a verse using the primary provider.
// In the future, this could implement fallback logic.
func (m *ProviderManager) GetVerse(book, chapter, verse, version string) (string, error) {
	if m.primary == nil {
		return "", fmt.Errorf("no primary provider configured")
	}
	return m.primary.GetVerse(book, chapter, verse, version)
}

// SearchWords searches for words using the primary provider.
// In the future, this could implement fallback logic.
func (m *ProviderManager) SearchWords(query, version string) ([]SearchResult, error) {
	if m.primary == nil {
		return nil, fmt.Errorf("no primary provider configured")
	}
	return m.primary.SearchWords(query, version)
}
