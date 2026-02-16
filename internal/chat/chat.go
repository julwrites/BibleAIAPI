package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"bible-api-service/internal/bible"
	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/util"

	"github.com/xeipuuv/gojsonschema"
)

const (
	promptHeaderHistory       = "Previous Conversation Context:\n"
	promptHistoryItemFormat   = "- %s\n"
	promptHeaderVerses        = "\n\nBible Verses:\n"
	promptHeaderSearchResults = "\n\nRelevant Search Results:\n"
	promptInstructionHTML     = "\n\nPlease format your response using semantic HTML."
	promptItemFormat          = "%s: %s"
	promptSectionSeparator    = "\n\n"
)

// BibleProviderRegistry defines the interface for retrieving Bible providers.
type BibleProviderRegistry interface {
	GetProvider(name string) (bible.Provider, error)
}

// GetLLMClient defines the function signature for getting an LLM client.
type GetLLMClient func() (provider.LLMClient, error)

// ChatService orchestrates fetching Bible verses, processing them, and interacting with an LLM.
type ChatService struct {
	BibleProviderRegistry BibleProviderRegistry
	GetLLMClient          GetLLMClient
}

// NewChatService creates a new ChatService.
func NewChatService(registry BibleProviderRegistry, getLLMClient GetLLMClient) *ChatService {
	return &ChatService{
		BibleProviderRegistry: registry,
		GetLLMClient:          getLLMClient,
	}
}

// Request represents the input for the chat service.
type Request struct {
	VerseRefs  []string `json:"verse_refs"`
	Words      []string `json:"words"`
	Version    string   `json:"version"`
	Provider   string   `json:"provider"`
	Prompt     string   `json:"prompt"`
	Schema     string   `json:"schema"`
	AIProvider string   `json:"ai_provider"`
	Stream     bool     `json:"stream"`
	History    []string `json:"history"`
}

// Response represents the structured output from the LLM.
type Response map[string]interface{}

// Result represents the outcome of a chat process, supporting both blocking and streaming.
type Result struct {
	Data     Response               // For blocking response
	Stream   <-chan string          // For streaming response
	Meta     map[string]interface{} // Metadata (e.g. provider name)
	IsStream bool                   // Flag to indicate if it's a stream
}

// Process handles the chat request.
func (s *ChatService) Process(ctx context.Context, req Request) (*Result, error) {
	// Get the provider
	bibleProvider, err := s.BibleProviderRegistry.GetProvider(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider '%s': %w", req.Provider, err)
	}

	// 1. Retrieve verses
	var verseTexts []string
	for _, verseRef := range req.VerseRefs {
		book, chapter, verseNum, err := util.ParseVerseReference(verseRef)
		if err != nil {
			return nil, fmt.Errorf("invalid verse reference format (%s): %w", verseRef, err)
		}

		verseHTML, err := bibleProvider.GetVerse(book, chapter, verseNum, req.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to get verse %s: %w", verseRef, err)
		}

		// 2. Keep the verse HTML content to preserve structure/poetry
		verseTexts = append(verseTexts, fmt.Sprintf(promptItemFormat, verseRef, verseHTML))
	}

	// 3. Search for words and add to context
	var searchResults []string
	for _, word := range req.Words {
		results, err := bibleProvider.SearchWords(word, req.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to search word %s: %w", word, err)
		}
		for _, result := range results {
			searchResults = append(searchResults, fmt.Sprintf(promptItemFormat, result.Verse, result.Text))
		}
	}

	// 4. Add the text content to the chat context for llm
	var promptBuilder strings.Builder

	if historyStr := formatHistory(req.History); historyStr != "" {
		promptBuilder.WriteString(historyStr)
		promptBuilder.WriteString(promptSectionSeparator)
	}

	promptBuilder.WriteString(req.Prompt)

	if len(verseTexts) > 0 {
		promptBuilder.WriteString(promptHeaderVerses)
		promptBuilder.WriteString(strings.Join(verseTexts, promptSectionSeparator))
	}

	if len(searchResults) > 0 {
		promptBuilder.WriteString(promptHeaderSearchResults)
		promptBuilder.WriteString(strings.Join(searchResults, promptSectionSeparator))
	}

	// Append instruction to return semantic HTML
	promptBuilder.WriteString(promptInstructionHTML)

	llmPrompt := promptBuilder.String()

	// 5. Refer to the system prompt specified by the request, and send this
	llmClient, err := s.GetLLMClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get llm client: %w", err)
	}

	// Inject preferred provider if specified
	if req.AIProvider != "" {
		ctx = context.WithValue(ctx, provider.PreferredProviderKey, req.AIProvider)
	}

	if req.Stream {
		ch, providerName, err := llmClient.Stream(ctx, llmPrompt)
		if err != nil {
			return nil, fmt.Errorf("failed to stream llm: %w", err)
		}
		return &Result{
			Stream:   ch,
			IsStream: true,
			Meta:     map[string]interface{}{"ai_provider": providerName},
		}, nil
	} else {
		// 6. Require the llm response to be structured output
		llmResponse, providerName, err := llmClient.Query(ctx, llmPrompt, req.Schema)
		if err != nil {
			return nil, fmt.Errorf("failed to query llm: %w", err)
		}

		// Validate the response against the schema
		if req.Schema != "" {
			schemaLoader := gojsonschema.NewStringLoader(req.Schema)
			documentLoader := gojsonschema.NewStringLoader(llmResponse)

			result, err := gojsonschema.Validate(schemaLoader, documentLoader)
			if err != nil {
				return nil, fmt.Errorf("failed to validate response against schema: %w", err)
			}

			if !result.Valid() {
				var errs []string
				for _, desc := range result.Errors() {
					errs = append(errs, desc.String())
				}
				return nil, fmt.Errorf("response validation failed: %s", strings.Join(errs, "; "))
			}
		}

		// 7. Return the structured output.
		var result Response
		if err := json.Unmarshal([]byte(llmResponse), &result); err != nil {
			return nil, fmt.Errorf("failed to parse llm response: %w", err)
		}

		return &Result{
			Data:     result,
			IsStream: false,
			Meta:     map[string]interface{}{"ai_provider": providerName},
		}, nil
	}
}

// formatHistory formats the history list into a string, limiting to the last N entries.
func formatHistory(history []string) string {
	const maxHistory = 6
	if len(history) == 0 {
		return ""
	}

	start := 0
	if len(history) > maxHistory {
		start = len(history) - maxHistory
	}

	recentHistory := history[start:]
	var sb strings.Builder
	sb.WriteString(promptHeaderHistory)
	for _, msg := range recentHistory {
		sb.WriteString(fmt.Sprintf(promptHistoryItemFormat, msg))
	}
	return sb.String()
}
