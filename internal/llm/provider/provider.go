package provider

import "context"

// LLMClient is the interface that all LLM clients must implement.
type LLMClient interface {
	// Query sends a prompt to the LLM and returns the response, the provider name, and an error.
	Query(ctx context.Context, prompt string, schema string) (string, string, error)

	// Stream sends a prompt to the LLM and returns a channel of response chunks, the provider name, and an error.
	Stream(ctx context.Context, prompt string) (<-chan string, string, error)

	// Name returns the name of the provider.
	Name() string
}
