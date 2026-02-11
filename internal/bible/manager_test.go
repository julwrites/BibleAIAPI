package bible

import (
	"testing"
)

type MockProvider struct {
	GetVerseFunc    func(book, chapter, verse, version string) (string, error)
	SearchWordsFunc func(query, version string) ([]SearchResult, error)
	GetVersionsFunc func() ([]ProviderVersion, error)
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

func (m *MockProvider) GetVersions() ([]ProviderVersion, error) {
	if m.GetVersionsFunc != nil {
		return m.GetVersionsFunc()
	}
	return nil, nil
}

func TestProviderManager(t *testing.T) {
	t.Run("GetVerse success", func(t *testing.T) {
		mock := &MockProvider{
			GetVerseFunc: func(book, chapter, verse, version string) (string, error) {
				return "Jesus wept", nil
			},
		}
		pm := NewProviderManager(mock)
		verse, err := pm.GetVerse("John", "11", "35", "ESV")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if verse != "Jesus wept" {
			t.Errorf("expected 'Jesus wept', got '%s'", verse)
		}
	})

	t.Run("GetVerse no primary", func(t *testing.T) {
		pm := NewProviderManager(nil)
		_, err := pm.GetVerse("John", "11", "35", "ESV")
		if err == nil {
			t.Error("expected error when no primary provider is set")
		}
	})

	t.Run("SearchWords success", func(t *testing.T) {
		mock := &MockProvider{
			SearchWordsFunc: func(query, version string) ([]SearchResult, error) {
				return []SearchResult{{Text: "Jesus wept"}}, nil
			},
		}
		pm := NewProviderManager(mock)
		results, err := pm.SearchWords("wept", "ESV")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 1 || results[0].Text != "Jesus wept" {
			t.Errorf("unexpected results: %v", results)
		}
	})

	t.Run("SearchWords no primary", func(t *testing.T) {
		pm := NewProviderManager(nil)
		_, err := pm.SearchWords("wept", "ESV")
		if err == nil {
			t.Error("expected error when no primary provider is set")
		}
	})

	t.Run("GetProvider success", func(t *testing.T) {
		mock := &MockProvider{}
		pm := NewProviderManager(mock)
		pm.RegisterProvider("test", mock)

		p, err := pm.GetProvider("test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p != mock {
			t.Error("expected mock provider")
		}
	})

	t.Run("GetProvider not found", func(t *testing.T) {
		pm := NewProviderManager(nil)
		_, err := pm.GetProvider("unknown")
		if err == nil {
			t.Error("expected error for unknown provider")
		}
	})

	t.Run("RegisterProvider lazy init", func(t *testing.T) {
		// Create manager with nil map (simulated by manual struct construction if needed,
        // but NewProviderManager initializes it. To test lazy init, we'd need to bypass constructor or
        // set it to nil manually).
        // Since `providers` field is unexported, we can't set it to nil easily from outside package
        // unless we are in same package. This test file IS in package bible.
		pm := &ProviderManager{primary: nil} // providers map is nil
		mock := &MockProvider{}

		// Should not panic and should initialize map
		pm.RegisterProvider("test", mock)

		p, err := pm.GetProvider("test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p != mock {
			t.Error("expected mock provider")
		}
	})
}
