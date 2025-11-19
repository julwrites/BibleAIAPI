package handlers

// QueryRequest represents the request body for the /query endpoint.
type QueryRequest struct {
	Query struct {
		Verses     []string `json:"verses"`
		Words      []string `json:"words"`
		OQuery     string   `json:"oquery"`
		ChatPrompt string   `json:"chat_prompt"`
		ChatSchema string   `json:"chat_schema"`
	} `json:"query"`
	Context struct {
		Instruction string   `json:"instruction"`
		PQuery      []string `json:"pquery"`
		Verses      []string `json:"verses"`
		Words       []string `json:"words"`
		User        struct {
			Version string `json:"version"`
		} `json:"user"`
	} `json:"context"`
}
