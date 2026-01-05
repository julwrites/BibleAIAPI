package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"bible-api-service/internal/biblegateway"
	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/util"
)

// BibleGatewayClient defines the interface for the Bible Gateway client.
type BibleGatewayClient interface {
	GetVerse(book, chapter, verse, version string) (string, error)
	SearchWords(query, version string) ([]biblegateway.SearchResult, error)
}

// GetLLMClient defines the function signature for getting an LLM client.
type GetLLMClient func() (provider.LLMClient, error)

// ChatService orchestrates fetching Bible verses, processing them, and interacting with an LLM.
type ChatService struct {
	BibleGatewayClient BibleGatewayClient
	GetLLMClient       GetLLMClient
}

// NewChatService creates a new ChatService.
func NewChatService(bgClient BibleGatewayClient, getLLMClient GetLLMClient) *ChatService {
	return &ChatService{
		BibleGatewayClient: bgClient,
		GetLLMClient:       getLLMClient,
	}
}

// Request represents the input for the chat service.
type Request struct {
	VerseRefs []string `json:"verse_refs"`
	Words     []string `json:"words"`
	Version   string   `json:"version"`
	Prompt    string   `json:"prompt"`
	Schema    string   `json:"schema"`
}

// Response represents the structured output from the LLM.
type Response map[string]interface{}

// Process handles the chat request.
func (s *ChatService) Process(ctx context.Context, req Request) (Response, error) {
	// 1. Retrieve verses from biblegateway
	var verseTexts []string
	for _, verseRef := range req.VerseRefs {
		book, chapter, verseNum, err := util.ParseVerseReference(verseRef)
		if err != nil {
			return nil, fmt.Errorf("invalid verse reference format (%s): %w", verseRef, err)
		}

		verseHTML, err := s.BibleGatewayClient.GetVerse(book, chapter, verseNum, req.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to get verse %s: %w", verseRef, err)
		}

		// 2. Keep the verse HTML content to preserve structure/poetry
		verseTexts = append(verseTexts, fmt.Sprintf("%s: %s", verseRef, verseHTML))
	}

	// 3. Search for words and add to context
	var searchResults []string
	for _, word := range req.Words {
		results, err := s.BibleGatewayClient.SearchWords(word, req.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to search word %s: %w", word, err)
		}
		for _, result := range results {
			searchResults = append(searchResults, fmt.Sprintf("%s: %s", result.Verse, result.Text))
		}
	}

	// 4. Add the text content to the chat context for llm
	var promptBuilder strings.Builder
	promptBuilder.WriteString(req.Prompt)

	if len(verseTexts) > 0 {
		promptBuilder.WriteString("\n\nBible Verses:\n")
		promptBuilder.WriteString(strings.Join(verseTexts, "\n\n"))
	}

	if len(searchResults) > 0 {
		promptBuilder.WriteString("\n\nRelevant Search Results:\n")
		promptBuilder.WriteString(strings.Join(searchResults, "\n\n"))
	}

	// Append instruction to return semantic HTML
	promptBuilder.WriteString("\n\nPlease format your response using semantic HTML.")

	llmPrompt := promptBuilder.String()

	// 5. Refer to the system prompt specified by the request, and send this
	llmClient, err := s.GetLLMClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get llm client: %w", err)
	}

	// 6. Require the llm response to be structured output
	llmResponse, err := llmClient.Query(ctx, llmPrompt, req.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to query llm: %w", err)
	}

	// 7. Return the structured output.
	var result Response
	if err := json.Unmarshal([]byte(llmResponse), &result); err != nil {
		return nil, fmt.Errorf("failed to parse llm response: %w", err)
	}

	return result, nil
}
