package handlers

// QueryRequest represents the request body for the /query endpoint.
type QueryRequest struct {
	Query struct {
		Verses []string `json:"verses,omitempty"`
		Words  []string `json:"words,omitempty"`
		Prompt string   `json:"prompt,omitempty"`
	} `json:"query"`
	Context struct {
		History []string `json:"history,omitempty"`
		Schema  string   `json:"schema,omitempty"`
		Verses  []string `json:"verses,omitempty"`
		Words   []string `json:"words,omitempty"`
		User    struct {
			Version string `json:"version"`
		} `json:"user"`
	} `json:"context,omitempty"`
}
