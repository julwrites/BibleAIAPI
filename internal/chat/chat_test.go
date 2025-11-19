package chat

import (
	"context"
	"errors"
	"testing"

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

	mockBgClient.On("GetVerse", "John", "3", "16", "NIV").Return("<h1>John 3:16</h1><p>For God so loved the world...</p>", nil)
	mockLLMClient.On("Query", mock.Anything, "Explain this verse.\n\nBible Verses:\nJohn 3:16For God so loved the world...", req.Schema).Return(`{"explanation": "It means God loves everyone."}`, nil)

	resp, err := chatService.Process(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "It means God loves everyone.", resp["explanation"])

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

	mockBgClient.On("GetVerse", "1 John", "3", "16", "NIV").Return("<h1>1 John 3:16</h1><p>This is how we know what love is...</p>", nil)
	mockLLMClient.On("Query", mock.Anything, "Explain this verse.\n\nBible Verses:\n1 John 3:16This is how we know what love is...", req.Schema).Return(`{"explanation": "It is about sacrificial love."}`, nil)

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
