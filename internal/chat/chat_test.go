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

// MockProvider is a mock type for bible.Provider
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) GetVerse(book, chapter, verse, version string) (string, error) {
	args := m.Called(book, chapter, verse, version)
	return args.String(0), args.Error(1)
}

func (m *MockProvider) SearchWords(query, version string) ([]bible.SearchResult, error) {
	args := m.Called(query, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]bible.SearchResult), args.Error(1)
}

func (m *MockProvider) GetVersions() ([]bible.ProviderVersion, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]bible.ProviderVersion), args.Error(1)
}

// MockBibleProviderRegistry is a mock type for the BibleProviderRegistry interface
type MockBibleProviderRegistry struct {
	mock.Mock
}

func (m *MockBibleProviderRegistry) GetProvider(name string) (bible.Provider, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(bible.Provider), args.Error(1)
}

// MockLLMClient is a mock type for the LLMClient interface
type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) Query(ctx context.Context, prompt, schema string) (string, string, error) {
	args := m.Called(ctx, prompt, schema)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockLLMClient) Stream(ctx context.Context, prompt string) (<-chan string, string, error) {
	args := m.Called(ctx, prompt)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(<-chan string), args.String(1), args.Error(2)
}

func (m *MockLLMClient) Name() string {
	args := m.Called()
	return args.String(0)
}

func TestChatService_Process_Success(t *testing.T) {
	mockRegistry := new(MockBibleProviderRegistry)
	mockProvider := new(MockProvider)
	mockLLMClient := new(MockLLMClient)

	mockGetLLMClient := func() (provider.LLMClient, error) {
		return mockLLMClient, nil
	}

	chatService := NewChatService(mockRegistry, mockGetLLMClient)

	req := Request{
		VerseRefs: []string{"John 3:16"},
		Version:   "NIV",
		Provider:  "biblegateway",
		Prompt:    "Explain this verse.",
		Schema:    `{"type": "object", "properties": {"explanation": {"type": "string"}}}`,
	}

	mockRegistry.On("GetProvider", "biblegateway").Return(mockProvider, nil)

	verseHTML := "<h1>John 3:16</h1><p>For God so loved the world...</p>"
	mockProvider.On("GetVerse", "John", "3", "16", "NIV").Return(verseHTML, nil)

	// Expect the prompt to contain the original HTML and the new instruction
	expectedPromptPart := "John 3:16: <h1>John 3:16</h1><p>For God so loved the world...</p>"
	expectedInstruction := "Please format your response using semantic HTML."

	mockLLMClient.On("Query", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return strings.Contains(prompt, expectedPromptPart) && strings.Contains(prompt, expectedInstruction)
	}), req.Schema).Return(`{"explanation": "It means God loves everyone."}`, "mock-provider", nil)

	result, err := chatService.Process(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsStream)
	assert.Equal(t, "It means God loves everyone.", result.Data["explanation"])
	assert.Equal(t, "mock-provider", result.Meta["ai_provider"])

	mockRegistry.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}

func TestChatService_Process_VersesAndWords(t *testing.T) {
	mockRegistry := new(MockBibleProviderRegistry)
	mockProvider := new(MockProvider)
	mockLLMClient := new(MockLLMClient)

	mockGetLLMClient := func() (provider.LLMClient, error) {
		return mockLLMClient, nil
	}

	chatService := NewChatService(mockRegistry, mockGetLLMClient)

	req := Request{
		VerseRefs: []string{"1 Corinthians 15:10", "Genesis 5:1"},
		Words:     []string{"Grace"},
		Version:   "ESV",
		Provider:  "biblegateway",
		Prompt:    "Which of these verses are relevant to these themes?",
		Schema:    `{"type": "object", "properties": {"response": {"type": "string"}}}`,
	}

	mockRegistry.On("GetProvider", "biblegateway").Return(mockProvider, nil)

	// Mock GetVerse calls
	mockProvider.On("GetVerse", "1 Corinthians", "15", "10", "ESV").Return("<p>But by the grace of God I am what I am...</p>", nil)
	mockProvider.On("GetVerse", "Genesis", "5", "1", "ESV").Return("<p>This is the book of the generations of Adam...</p>", nil)

	// Mock SearchWords calls
	searchResults := []bible.SearchResult{
		{Verse: "Ephesians 2:8", Text: "For by grace you have been saved..."},
	}
	mockProvider.On("SearchWords", "Grace", "ESV").Return(searchResults, nil)

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
	}), req.Schema).Return(`{"response": "Both are relevant."}`, "mock-provider", nil)

	result, err := chatService.Process(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsStream)
	assert.Equal(t, "Both are relevant.", result.Data["response"])

	mockRegistry.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}

func TestChatService_Process_WithWords(t *testing.T) {
	mockRegistry := new(MockBibleProviderRegistry)
	mockProvider := new(MockProvider)
	mockLLMClient := new(MockLLMClient)

	mockGetLLMClient := func() (provider.LLMClient, error) {
		return mockLLMClient, nil
	}

	chatService := NewChatService(mockRegistry, mockGetLLMClient)

	req := Request{
		Words:    []string{"Grace"},
		Version:  "NIV",
		Provider: "biblegateway",
		Prompt:   "Summarize these search results.",
		Schema:   `{"type": "object", "properties": {"summary": {"type": "string"}}}`,
	}

	mockRegistry.On("GetProvider", "biblegateway").Return(mockProvider, nil)

	searchResults := []bible.SearchResult{
		{Verse: "Ephesians 2:8", Text: "For it is by grace you have been saved..."},
	}

	mockProvider.On("SearchWords", "Grace", "NIV").Return(searchResults, nil)

	expectedPromptPart := "Summarize these search results.\n\nRelevant Search Results:\nEphesians 2:8: For it is by grace you have been saved..."
	expectedInstruction := "Please format your response using semantic HTML."

	mockLLMClient.On("Query", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return strings.Contains(prompt, expectedPromptPart) && strings.Contains(prompt, expectedInstruction)
	}), req.Schema).Return(`{"summary": "Grace saves."}`, "mock-provider", nil)

	result, err := chatService.Process(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsStream)
	assert.Equal(t, "Grace saves.", result.Data["summary"])

	mockRegistry.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}

func TestChatService_Process_BookWithSpace(t *testing.T) {
	mockRegistry := new(MockBibleProviderRegistry)
	mockProvider := new(MockProvider)
	mockLLMClient := new(MockLLMClient)

	mockGetLLMClient := func() (provider.LLMClient, error) {
		return mockLLMClient, nil
	}

	chatService := NewChatService(mockRegistry, mockGetLLMClient)

	req := Request{
		VerseRefs: []string{"1 John 3:16"},
		Version:   "NIV",
		Provider:  "biblegateway",
		Prompt:    "Explain this verse.",
		Schema:    `{"type": "object", "properties": {"explanation": {"type": "string"}}}`,
	}

	mockRegistry.On("GetProvider", "biblegateway").Return(mockProvider, nil)

	verseHTML := "<h1>1 John 3:16</h1><p>This is how we know what love is...</p>"
	mockProvider.On("GetVerse", "1 John", "3", "16", "NIV").Return(verseHTML, nil)

	expectedPromptPart := "1 John 3:16: <h1>1 John 3:16</h1><p>This is how we know what love is...</p>"
	expectedInstruction := "Please format your response using semantic HTML."

	mockLLMClient.On("Query", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return strings.Contains(prompt, expectedPromptPart) && strings.Contains(prompt, expectedInstruction)
	}), req.Schema).Return(`{"explanation": "It is about sacrificial love."}`, "mock-provider", nil)

	result, err := chatService.Process(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsStream)
	assert.Equal(t, "It is about sacrificial love.", result.Data["explanation"])

	mockRegistry.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}

func TestChatService_Process_BibleGatewayError(t *testing.T) {
	mockRegistry := new(MockBibleProviderRegistry)
	mockProvider := new(MockProvider)
	mockGetLLMClient := func() (provider.LLMClient, error) {
		return nil, nil
	}

	chatService := NewChatService(mockRegistry, mockGetLLMClient)

	req := Request{
		VerseRefs: []string{"Invalid 1:1"},
		Version:   "NIV",
		Provider:  "biblegateway",
		Prompt:    "Explain this verse.",
	}

	mockRegistry.On("GetProvider", "biblegateway").Return(mockProvider, nil)

	mockProvider.On("GetVerse", "Invalid", "1", "1", "NIV").Return("", errors.New("verse not found"))

	_, err := chatService.Process(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verse not found")

	mockRegistry.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
}

func TestChatService_Process_LLMError(t *testing.T) {
	mockRegistry := new(MockBibleProviderRegistry)
	mockProvider := new(MockProvider)
	mockLLMClient := new(MockLLMClient)
	mockGetLLMClient := func() (provider.LLMClient, error) {
		return mockLLMClient, nil
	}

	chatService := NewChatService(mockRegistry, mockGetLLMClient)

	req := Request{
		VerseRefs: []string{"John 3:16"},
		Version:   "NIV",
		Provider:  "biblegateway",
		Prompt:    "Explain this verse.",
		Schema:    `{"type": "object", "properties": {"explanation": {"type": "string"}}}`,
	}

	mockRegistry.On("GetProvider", "biblegateway").Return(mockProvider, nil)
	mockProvider.On("GetVerse", "John", "3", "16", "NIV").Return("<p>For God so loved the world...</p>", nil)
	mockLLMClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return("", "", errors.New("LLM failed"))

	_, err := chatService.Process(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM failed")

	mockRegistry.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}

func TestChatService_Process_Streaming(t *testing.T) {
	mockRegistry := new(MockBibleProviderRegistry)
	mockProvider := new(MockProvider)
	mockLLMClient := new(MockLLMClient)

	mockGetLLMClient := func() (provider.LLMClient, error) {
		return mockLLMClient, nil
	}

	chatService := NewChatService(mockRegistry, mockGetLLMClient)

	req := Request{
		VerseRefs:  []string{"John 3:16"},
		Version:    "NIV",
		Provider:   "biblegateway",
		Prompt:     "Explain this verse.",
		Stream:     true,
		AIProvider: "openai",
	}

	mockRegistry.On("GetProvider", "biblegateway").Return(mockProvider, nil)

	verseHTML := "<h1>John 3:16</h1><p>For God so loved the world...</p>"
	mockProvider.On("GetVerse", "John", "3", "16", "NIV").Return(verseHTML, nil)

	streamChan := make(chan string, 2)
	streamChan <- "God loves "
	streamChan <- "everyone."
	close(streamChan)

	mockLLMClient.On("Stream", mock.MatchedBy(func(ctx context.Context) bool {
		val, ok := ctx.Value(provider.PreferredProviderKey).(string)
		return ok && val == "openai"
	}), mock.Anything).Return((<-chan string)(streamChan), "openai", nil)

	result, err := chatService.Process(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsStream)
	assert.Equal(t, "openai", result.Meta["ai_provider"])

	var content string
	for chunk := range result.Stream {
		content += chunk
	}
	assert.Equal(t, "God loves everyone.", content)

	mockRegistry.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}
