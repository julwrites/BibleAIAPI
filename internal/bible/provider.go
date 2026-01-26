package bible

// SearchResult represents a search result from a Bible provider.
type SearchResult struct {
	Verse string `json:"verse"`
	Text  string `json:"text"`
	URL   string `json:"url"`
}

// Provider defines the interface for a Bible data provider.
type Provider interface {
	// GetVerse fetches a single Bible verse or passage by reference.
	// Returns the content as a string (often HTML) or an error.
	GetVerse(book, chapter, verse, version string) (string, error)

	// SearchWords searches for a word or phrase and returns a list of relevant verses.
	SearchWords(query, version string) ([]SearchResult, error)
}
