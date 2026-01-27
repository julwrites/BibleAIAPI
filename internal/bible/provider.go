package bible

// SearchResult represents a search result from a Bible provider.
type SearchResult struct {
	Verse string `json:"verse"`
	Text  string `json:"text"`
	URL   string `json:"url"`
}

// ProviderVersion represents a Bible version provided by a scraping source.
type ProviderVersion struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Code     string `json:"code"`
	Language string `json:"language"`
}

// Provider defines the interface for a Bible data provider.
type Provider interface {
	// GetVerse fetches a single Bible verse or passage by reference.
	// Returns the content as a string (often HTML) or an error.
	GetVerse(book, chapter, verse, version string) (string, error)

	// SearchWords searches for a word or phrase and returns a list of relevant verses.
	SearchWords(query, version string) ([]SearchResult, error)

	// GetVersions fetches the list of available Bible versions.
	GetVersions() ([]ProviderVersion, error)
}
