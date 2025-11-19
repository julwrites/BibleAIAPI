package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"bible-api-service/internal/llm/provider"
	"github.com/PuerkitoBio/goquery"
)

// BibleGatewayClient defines the interface for the Bible Gateway client.
type BibleGatewayClient interface {
	GetVerse(book, chapter, verse, version string) (string, error)
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
		lastSpaceIndex := strings.LastIndex(verseRef, " ")
		if lastSpaceIndex == -1 {
			return nil, fmt.Errorf("invalid verse reference format: %s", verseRef)
		}

		book := verseRef[:lastSpaceIndex]
		chapterAndVerseStr := verseRef[lastSpaceIndex+1:]

		chapterAndVerse := strings.Split(chapterAndVerseStr, ":")
		if len(chapterAndVerse) < 2 {
			return nil, fmt.Errorf("invalid verse reference format: %s", verseRef)
		}
		chapter := chapterAndVerse[0]
		verseNum := chapterAndVerse[1]

		verseHTML, err := s.BibleGatewayClient.GetVerse(book, chapter, verseNum, req.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to get verse %s: %w", verseRef, err)
		}

		// 2. Remove all tags, keeping only the text content
		plainText, err := stripHTML(verseHTML)
		if err != nil {
			return nil, fmt.Errorf("failed to strip html from verse %s: %w", verseRef, err)
		}
		verseTexts = append(verseTexts, plainText)
	}

	// 3. Add the text content to the chat context for llm
	fullVerseText := strings.Join(verseTexts, "\n\n")
	llmPrompt := fmt.Sprintf("%s\n\nBible Verses:\n%s", req.Prompt, fullVerseText)

	// 4. Refer to the system prompt specified by the request, and send this
	llmClient, err := s.GetLLMClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get llm client: %w", err)
	}

	// 5. Require the llm response to be structured output
	llmResponse, err := llmClient.Query(ctx, llmPrompt, req.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to query llm: %w", err)
	}

	// 6. Return the structured output.
	var result Response
	if err := json.Unmarshal([]byte(llmResponse), &result); err != nil {
		return nil, fmt.Errorf("failed to parse llm response: %w", err)
	}

	return result, nil
}

// stripHTML removes HTML tags from a string, returning only the text content.
func stripHTML(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}
	return doc.Text(), nil
}
