package llm

import (
	"context"
	"fmt"
	"os"
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
		return NewOpenAIClient()
	case "gemini":
		return NewGeminiClient()
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", llmProvider)
	}
}
