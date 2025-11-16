package provider

import "context"

// LLMClient is the interface that all LLM clients must implement.
type LLMClient interface {
	Query(ctx context.Context, prompt string, schema string) (string, error)
}
