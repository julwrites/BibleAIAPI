package chat

import (
	"context"
	"errors"
	"strings"
	"testing"

	"bible-api-service/internal/bible"
	"bible-api-service/internal/llm/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBibleGatewayClient is a mock type for the BibleGatewayClient interface
type MockBibleGatewayClient struct {
	mock.Mock
}

func (m *MockBibleGatewayClient) GetVerse(book, chapter, verse, version string) (string, error) {
	args := m.Called(book, chapter, verse, version)
	return args.String(0), args.Error(1)
}

func (m *MockBibleGatewayClient) SearchWords(query, version string) ([]bible.SearchResult, error) {
	args := m.Called(query, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]bible.SearchResult), args.Error(1)
}

// MockLLMClient is a mock type for the LLMClient interface
type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) Query(ctx context.Context, prompt, schema string) (string, error) {
	args := m.Called(ctx, prompt, schema)
	return args.String(0), args.Error(1)
}

func TestChatService_Process_Success(t *testing.T) {
	mockBgClient := new(MockBibleGatewayClient)
	mockLLMClient := new(MockLLMClient)

	mockGetLLMClient := func() (provider.LLMClient, error) {
		return mockLLMClient, nil
	}

	chatService := NewChatService(mockBgClient, mockGetLLMClient)

	req := Request{
		VerseRefs: []string{"John 3:16"},
		Version:   "NIV",
		Prompt:    "Explain this verse.",
		Schema:    `{"type": "object", "properties": {"explanation": {"type": "string"}}}`,
	}

	verseHTML := "<h1>John 3:16</h1><p>For God so loved the world...</p>"
	mockBgClient.On("GetVerse", "John", "3", "16", "NIV").Return(verseHTML, nil)

	// Expect the prompt to contain the original HTML and the new instruction
	expectedPromptPart := "John 3:16: <h1>John 3:16</h1><p>For God so loved the world...</p>"
	expectedInstruction := "Please format your response using semantic HTML."

	mockLLMClient.On("Query", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return strings.Contains(prompt, expectedPromptPart) && strings.Contains(prompt, expectedInstruction)
	}), req.Schema).Return(`{"explanation": "It means God loves everyone."}`, nil)

	resp, err := chatService.Process(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "It means God loves everyone.", resp["explanation"])

	mockBgClient.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}

func TestChatService_Process_VersesAndWords(t *testing.T) {
	mockBgClient := new(MockBibleGatewayClient)
	mockLLMClient := new(MockLLMClient)

	mockGetLLMClient := func() (provider.LLMClient, error) {
		return mockLLMClient, nil
	}

	chatService := NewChatService(mockBgClient, mockGetLLMClient)

	req := Request{
		VerseRefs: []string{"1 Corinthians 15:10", "Genesis 5:1"},
		Words:     []string{"Grace"},
		Version:   "ESV",
		Prompt:    "Which of these verses are relevant to these themes?",
		Schema:    `{"type": "object", "properties": {"response": {"type": "string"}}}`,
	}

	// Mock GetVerse calls
	mockBgClient.On("GetVerse", "1 Corinthians", "15", "10", "ESV").Return("<p>But by the grace of God I am what I am...</p>", nil)
	mockBgClient.On("GetVerse", "Genesis", "5", "1", "ESV").Return("<p>This is the book of the generations of Adam...</p>", nil)

	// Mock SearchWords calls
	searchResults := []bible.SearchResult{
		{Verse: "Ephesians 2:8", Text: "For by grace you have been saved..."},
	}
	mockBgClient.On("SearchWords", "Grace", "ESV").Return(searchResults, nil)

	// Mock LLM Query
	// The prompt should contain both verses (with HTML) and search results
	expectedPromptPart1 := "Bible Verses:\n1 Corinthians 15:10: <p>But by the grace of God I am what I am...</p>\n\nGenesis 5:1: <p>This is the book of the generations of Adam...</p>"
	expectedPromptPart2 := "Relevant Search Results:\nEphesians 2:8: For by grace you have been saved..."
	expectedInstruction := "Please format your response using semantic HTML."

	mockLLMClient.On("Query", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return strings.Contains(prompt, req.Prompt) &&
			strings.Contains(prompt, expectedPromptPart1) &&
			strings.Contains(prompt, expectedPromptPart2) &&
			strings.Contains(prompt, expectedInstruction)
	}), req.Schema).Return(`{"response": "Both are relevant."}`, nil)

	resp, err := chatService.Process(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Both are relevant.", resp["response"])

	mockBgClient.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}

func TestChatService_Process_WithWords(t *testing.T) {
	mockBgClient := new(MockBibleGatewayClient)
	mockLLMClient := new(MockLLMClient)

	mockGetLLMClient := func() (provider.LLMClient, error) {
		return mockLLMClient, nil
	}

	chatService := NewChatService(mockBgClient, mockGetLLMClient)

	req := Request{
		Words:   []string{"Grace"},
		Version: "NIV",
		Prompt:  "Summarize these search results.",
		Schema:  `{"type": "object", "properties": {"summary": {"type": "string"}}}`,
	}

	searchResults := []bible.SearchResult{
		{Verse: "Ephesians 2:8", Text: "For it is by grace you have been saved..."},
	}

	mockBgClient.On("SearchWords", "Grace", "NIV").Return(searchResults, nil)

	expectedPromptPart := "Summarize these search results.\n\nRelevant Search Results:\nEphesians 2:8: For it is by grace you have been saved..."
	expectedInstruction := "Please format your response using semantic HTML."

	mockLLMClient.On("Query", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return strings.Contains(prompt, expectedPromptPart) && strings.Contains(prompt, expectedInstruction)
	}), req.Schema).Return(`{"summary": "Grace saves."}`, nil)

	resp, err := chatService.Process(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Grace saves.", resp["summary"])

	mockBgClient.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}

func TestChatService_Process_BookWithSpace(t *testing.T) {
	mockBgClient := new(MockBibleGatewayClient)
	mockLLMClient := new(MockLLMClient)

	mockGetLLMClient := func() (provider.LLMClient, error) {
		return mockLLMClient, nil
	}

	chatService := NewChatService(mockBgClient, mockGetLLMClient)

	req := Request{
		VerseRefs: []string{"1 John 3:16"},
		Version:   "NIV",
		Prompt:    "Explain this verse.",
		Schema:    `{"type": "object", "properties": {"explanation": {"type": "string"}}}`,
	}

	verseHTML := "<h1>1 John 3:16</h1><p>This is how we know what love is...</p>"
	mockBgClient.On("GetVerse", "1 John", "3", "16", "NIV").Return(verseHTML, nil)

	expectedPromptPart := "1 John 3:16: <h1>1 John 3:16</h1><p>This is how we know what love is...</p>"
	expectedInstruction := "Please format your response using semantic HTML."

	mockLLMClient.On("Query", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return strings.Contains(prompt, expectedPromptPart) && strings.Contains(prompt, expectedInstruction)
	}), req.Schema).Return(`{"explanation": "It is about sacrificial love."}`, nil)

	resp, err := chatService.Process(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "It is about sacrificial love.", resp["explanation"])

	mockBgClient.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}

func TestChatService_Process_BibleGatewayError(t *testing.T) {
	mockBgClient := new(MockBibleGatewayClient)
	mockGetLLMClient := func() (provider.LLMClient, error) {
		return nil, nil
	}

	chatService := NewChatService(mockBgClient, mockGetLLMClient)

	req := Request{
		VerseRefs: []string{"Invalid 1:1"},
		Version:   "NIV",
		Prompt:    "Explain this verse.",
	}

	mockBgClient.On("GetVerse", "Invalid", "1", "1", "NIV").Return("", errors.New("verse not found"))

	_, err := chatService.Process(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verse not found")

	mockBgClient.AssertExpectations(t)
}

func TestChatService_Process_LLMError(t *testing.T) {
	mockBgClient := new(MockBibleGatewayClient)
	mockLLMClient := new(MockLLMClient)
	mockGetLLMClient := func() (provider.LLMClient, error) {
		return mockLLMClient, nil
	}

	chatService := NewChatService(mockBgClient, mockGetLLMClient)

	req := Request{
		VerseRefs: []string{"John 3:16"},
		Version:   "NIV",
		Prompt:    "Explain this verse.",
		Schema:    `{"type": "object", "properties": {"explanation": {"type": "string"}}}`,
	}

	mockBgClient.On("GetVerse", "John", "3", "16", "NIV").Return("<p>For God so loved the world...</p>", nil)
	mockLLMClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("LLM failed"))

	_, err := chatService.Process(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM failed")

	mockBgClient.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}
