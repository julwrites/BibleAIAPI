package client

// QueryRequest represents the request body for the /query endpoint.
type QueryRequest struct {
	Query   Query   `json:"query"`
	Context Context `json:"context"`
}

type Query struct {
	Verses     []string `json:"verses,omitempty"`
	Words      []string `json:"words,omitempty"`
	OQuery     string   `json:"oquery,omitempty"`
	ChatPrompt string   `json:"chat_prompt,omitempty"`
	ChatSchema string   `json:"chat_schema,omitempty"`
}

type Context struct {
	Instruction string   `json:"instruction,omitempty"`
	PQuery      []string `json:"pquery,omitempty"`
	Verses      []string `json:"verses,omitempty"`
	Words       []string `json:"words,omitempty"`
	User        User     `json:"user,omitempty"`
}

type User struct {
	Version string `json:"version,omitempty"`
}

// Response types

type VerseResponse struct {
	Verse string `json:"verse"`
}

type WordSearchResponse []SearchResult

type SearchResult struct {
	Verse string `json:"verse"`
	URL   string `json:"url"`
}

type OQueryResponse struct {
	Text       string         `json:"text"`
	References []SearchResult `json:"references"`
}

// ErrorResponse represents an error from the API.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
