package llm

import (
	"context"
	"fmt"
	"os"
)

var (
	NewOpenAIClientFunc = NewOpenAIClient
	NewGeminiClientFunc = NewGeminiClient
)

type LLMClient interface {
	Query(ctx context.Context, prompt string, schema string) (string, error)
}

func GetClient() (LLMClient, error) {
	llmProvider := os.Getenv("LLM_PROVIDER")
	if llmProvider == "" {
		llmProvider = "openai" // Default to openai
	}

	switch llmProvider {
	case "openai":
		return NewOpenAIClientFunc()
	case "gemini":
		return NewGeminiClientFunc()
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", llmProvider)
	}
}
