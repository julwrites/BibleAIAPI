package mocks

import "bible-api-service/internal/bible"

// MockBibleClient is a mock implementation of the BibleProvider
type MockBibleClient struct {
	VerseResponse  string
	VerseError     error
	SearchResults  []bible.SearchResult
	SearchError    error
	GetVerseCalled bool
	SearchCalled   bool
}

func (m *MockBibleClient) GetVerse(book, chapter, verse, version string) (string, error) {
	m.GetVerseCalled = true
	if m.VerseError != nil {
		return "", m.VerseError
	}
	return m.VerseResponse, nil
}

func (m *MockBibleClient) SearchWords(query, version string) ([]bible.SearchResult, error) {
	m.SearchCalled = true
	if m.SearchError != nil {
		return nil, m.SearchError
	}
	return m.SearchResults, nil
}

func (m *MockBibleClient) GetVersions() ([]bible.ProviderVersion, error) {
	return []bible.ProviderVersion{}, nil
}
