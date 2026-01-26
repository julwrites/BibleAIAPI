package bible

import (
	"errors"
	"testing"
)

// MockProvider is a mock implementation of the Provider interface.
type MockProvider struct {
	GetVerseFunc    func(book, chapter, verse, version string) (string, error)
	SearchWordsFunc func(query, version string) ([]SearchResult, error)
}

func (m *MockProvider) GetVerse(book, chapter, verse, version string) (string, error) {
	if m.GetVerseFunc != nil {
		return m.GetVerseFunc(book, chapter, verse, version)
	}
	return "", nil
}

func (m *MockProvider) SearchWords(query, version string) ([]SearchResult, error) {
	if m.SearchWordsFunc != nil {
		return m.SearchWordsFunc(query, version)
	}
	return nil, nil
}

func TestProviderManager_GetVerse(t *testing.T) {
	mockProvider := &MockProvider{
		GetVerseFunc: func(book, chapter, verse, version string) (string, error) {
			if book == "John" && chapter == "3" && verse == "16" {
				return "For God so loved the world", nil
			}
			return "", errors.New("not found")
		},
	}

	manager := NewProviderManager(mockProvider)

	// Test success
	verse, err := manager.GetVerse("John", "3", "16", "ESV")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verse != "For God so loved the world" {
		t.Errorf("expected verse to be 'For God so loved the world', got '%s'", verse)
	}

	// Test error
	_, err = manager.GetVerse("Genesis", "1", "1", "ESV")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProviderManager_SearchWords(t *testing.T) {
	mockProvider := &MockProvider{
		SearchWordsFunc: func(query, version string) ([]SearchResult, error) {
			if query == "love" {
				return []SearchResult{{Verse: "John 3:16", Text: "For God so loved..."}}, nil
			}
			return nil, errors.New("search failed")
		},
	}

	manager := NewProviderManager(mockProvider)

	// Test success
	results, err := manager.SearchWords("love", "ESV")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Verse != "John 3:16" {
		t.Errorf("expected verse 'John 3:16', got '%s'", results[0].Verse)
	}

	// Test error
	_, err = manager.SearchWords("hate", "ESV")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProviderManager_NoProvider(t *testing.T) {
	manager := NewProviderManager(nil)

	_, err := manager.GetVerse("John", "3", "16", "ESV")
	if err == nil {
		t.Error("expected error when no provider is set")
	}

	_, err = manager.SearchWords("love", "ESV")
	if err == nil {
		t.Error("expected error when no provider is set")
	}
}
