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

// BibleProviderRegistry defines the interface for retrieving Bible providers.
type BibleProviderRegistry interface {
	GetProvider(name string) (bible.Provider, error)
}

// GetLLMClient defines the function signature for getting an LLM client.
type GetLLMClient func() (provider.LLMClient, error)

// Service defines the interface for the chat service.
type Service interface {
	Process(ctx context.Context, req Request) (*Result, error)
}

// ChatService orchestrates fetching Bible verses, processing them, and interacting with an LLM.
type ChatService struct {
	BibleProviderRegistry BibleProviderRegistry
	GetLLMClient          GetLLMClient
}

// NewChatService creates a new ChatService.
func NewChatService(registry BibleProviderRegistry, getLLMClient GetLLMClient) Service {
	return &ChatService{
		BibleProviderRegistry: registry,
		GetLLMClient:          getLLMClient,
	}
}

// Request represents the input for the chat service.
type Request struct {
	VerseRefs            []string               `json:"verse_refs"`
	Words                []string               `json:"words"`
	Version              string                 `json:"version"`  // Deprecated: use PrioritizedProviders
	Provider             string                 `json:"provider"` // Deprecated: use PrioritizedProviders
	PrioritizedProviders []bible.ProviderConfig `json:"prioritized_providers"`
	Prompt               string                 `json:"prompt"`
	Schema               string                 `json:"schema"`
	AIProvider           string                 `json:"ai_provider"`
	Stream               bool                   `json:"stream"`
	History              []string               `json:"history"`
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
	// Backwards compatibility: if PrioritizedProviders is empty, use single provider/version
	if len(req.PrioritizedProviders) == 0 && req.Provider != "" {
		req.PrioritizedProviders = []bible.ProviderConfig{
			{Name: req.Provider, VersionCode: req.Version},
		}
	}

	if len(req.PrioritizedProviders) == 0 {
		return nil, fmt.Errorf("no bible providers configured")
	}

	// 1. Retrieve verses
	var verseTexts []string
	if len(req.VerseRefs) > 0 {
		var lastErr error
		success := false
		for _, providerConfig := range req.PrioritizedProviders {
			bibleProvider, err := s.BibleProviderRegistry.GetProvider(providerConfig.Name)
			if err != nil {
				lastErr = fmt.Errorf("provider %s not found: %w", providerConfig.Name, err)
				continue
			}

			currentVerseTexts := []string{}
			permErr := false
			for _, verseRef := range req.VerseRefs {
				book, chapter, verseNum, err := util.ParseVerseReference(verseRef)
				if err != nil {
					// Invalid reference is a permanent error, don't retry other providers
					return nil, fmt.Errorf("invalid verse reference format (%s): %w", verseRef, err)
				}

				verseHTML, err := bibleProvider.GetVerse(book, chapter, verseNum, providerConfig.VersionCode)
				if err != nil {
					// Provider failed for this verse
					lastErr = err
					permErr = true
					break // Try next provider for ALL verses (atomic per request? or per verse?)
					// Assumption: if a provider fails for one verse, it's likely flaky for others.
					// We'll retry the whole batch with next provider.
				}
				currentVerseTexts = append(currentVerseTexts, fmt.Sprintf("%s: %s", verseRef, verseHTML))
			}

			if !permErr {
				verseTexts = currentVerseTexts
				success = true
				break
			}
		}
		if !success {
			return nil, fmt.Errorf("failed to retrieve verses from any provider: %w", lastErr)
		}
	}

	// 3. Search for words and add to context
	var searchResults []string
	if len(req.Words) > 0 {
		var lastErr error
		success := false
		for _, providerConfig := range req.PrioritizedProviders {
			bibleProvider, err := s.BibleProviderRegistry.GetProvider(providerConfig.Name)
			if err != nil {
				continue
			}

			currentSearchResults := []string{}
			permErr := false
			for _, word := range req.Words {
				results, err := bibleProvider.SearchWords(word, providerConfig.VersionCode)
				if err != nil {
					lastErr = err
					permErr = true
					break
				}
				for _, result := range results {
					currentSearchResults = append(currentSearchResults, fmt.Sprintf("%s: %s", result.Verse, result.Text))
				}
			}

			if !permErr {
				searchResults = currentSearchResults
				success = true
				break
			}
		}
		if !success {
			return nil, fmt.Errorf("failed to search words from any provider: %w", lastErr)
		}
	}

	// 4. Add the text content to the chat context for llm
	var promptBuilder strings.Builder

	if historyStr := formatHistory(req.History); historyStr != "" {
		promptBuilder.WriteString(historyStr)
		promptBuilder.WriteString("\n\n")
	}

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
	sb.WriteString("Previous Conversation Context:\n")
	for _, msg := range recentHistory {
		sb.WriteString(fmt.Sprintf("- %s\n", msg))
	}
	return sb.String()
}
