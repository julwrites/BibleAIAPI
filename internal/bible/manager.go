package bible

import (
	"fmt"
)

// DefaultProviderName is the name of the default provider (Bible Gateway).
const DefaultProviderName = "biblegateway"

// ProviderManager manages multiple Bible providers.
type ProviderManager struct {
	primary   Provider
	providers map[string]Provider
}

// NewProviderManager creates a new ProviderManager with the given primary provider.
func NewProviderManager(primary Provider) *ProviderManager {
	return &ProviderManager{
		primary:   primary,
		providers: make(map[string]Provider),
	}
}

// RegisterProvider registers a provider with a given name.
func (m *ProviderManager) RegisterProvider(name string, p Provider) {
	if m.providers == nil {
		m.providers = make(map[string]Provider)
	}
	m.providers[name] = p
}

// GetProvider retrieves a provider by name.
func (m *ProviderManager) GetProvider(name string) (Provider, error) {
	if p, ok := m.providers[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("provider not found: %s", name)
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
