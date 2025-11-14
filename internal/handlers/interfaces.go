package handlers

import (
	"bible-api-service/internal/biblegateway"
	"bible-api-service/internal/llm"
	"context"
)

// BibleGatewayClient defines the interface for the Bible Gateway client.
type BibleGatewayClient interface {
	GetVerse(book, chapter, verse, version string) (string, error)
	SearchWords(query, version string) ([]biblegateway.SearchResult, error)
}

// LLMClient defines the interface for the LLM client.
type LLMClient interface {
	Query(ctx context.Context, query, schema string) (string, error)
}

// GetLLMClient defines the function signature for getting an LLM client.
type GetLLMClient func() (llm.LLMClient, error)
