package handlers

import (
	"bible-api-service/internal/biblegateway"
	"bible-api-service/internal/chat"
	"bible-api-service/internal/llm/provider"
	"context"
)

// BibleGatewayClient defines the interface for the Bible Gateway client.
type BibleGatewayClient interface {
	GetVerse(book, chapter, verse, version string) (string, error)
	SearchWords(query, version string) ([]biblegateway.SearchResult, error)
}

// LLMClient defines the interface for the LLM client.
type LLMClient provider.LLMClient

// GetLLMClient defines the function signature for getting an LLM client.
type GetLLMClient func() (provider.LLMClient, error)

// ChatService defines the interface for the chat service.
type ChatService interface {
	Process(ctx context.Context, req chat.Request) (chat.Response, error)
}
