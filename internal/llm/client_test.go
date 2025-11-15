package llm

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

type mockLLM struct {
	generateContentFunc func(context.Context, []llms.MessageContent, ...llms.CallOption) (*llms.ContentResponse, error)
}

func (m *mockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return m.generateContentFunc(ctx, messages, options...)
}

func (m *mockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "", nil
}

func TestGetClient(t *testing.T) {
	originalNewOpenAIClientFunc := NewOpenAIClientFunc
	originalNewGeminiClientFunc := NewGeminiClientFunc
	defer func() {
		NewOpenAIClientFunc = originalNewOpenAIClientFunc
		NewGeminiClientFunc = originalNewGeminiClientFunc
	}()

	NewOpenAIClientFunc = func() (*OpenAIClient, error) {
		return &OpenAIClient{llm: &mockLLM{}}, nil
	}
	NewGeminiClientFunc = func() (*GeminiClient, error) {
		return &GeminiClient{llm: &mockLLM{}}, nil
	}

	tests := []struct {
		name          string
		provider      string
		expectedType  string
		expectedError string
	}{
		{
			name:          "OpenAI provider",
			provider:      "openai",
			expectedType:  "*llm.OpenAIClient",
			expectedError: "",
		},
		{
			name:          "Gemini provider",
			provider:      "gemini",
			expectedType:  "*llm.GeminiClient",
			expectedError: "",
		},
		{
			name:          "Default to OpenAI",
			provider:      "",
			expectedType:  "*llm.OpenAIClient",
			expectedError: "",
		},
		{
			name:          "Unsupported provider",
			provider:      "unsupported",
			expectedType:  "",
			expectedError: "unsupported LLM provider: unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LLM_PROVIDER", tt.provider)
			defer os.Unsetenv("LLM_PROVIDER")

			client, err := GetClient()

			if err != nil && err.Error() != tt.expectedError {
				t.Errorf("unexpected error: got %v, want %v", err, tt.expectedError)
			}

			if err == nil {
				if clientType := fmt.Sprintf("%T", client); clientType != tt.expectedType {
					t.Errorf("unexpected client type: got %s, want %s", clientType, tt.expectedType)
				}
			}
		})
	}
}
