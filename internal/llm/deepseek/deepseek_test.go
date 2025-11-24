package deepseek

import (
	"context"
	"errors"
	"os"
	"testing"

	"bible-api-service/internal/secrets"

	"github.com/google/go-cmp/cmp"
	"github.com/tmc/langchaingo/llms"
)

type mockLLM struct {
	generateContentFunc func(context.Context, []llms.MessageContent, ...llms.CallOption) (*llms.ContentResponse, error)
}

func (m *mockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, messages, options...)
	}
	return nil, errors.New("generateContentFunc not implemented")
}

func (m *mockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "", errors.New("not implemented")
}

func TestNewClient(t *testing.T) {
	secretsClient := &secrets.EnvClient{}

	t.Run("No API key", func(t *testing.T) {
		os.Unsetenv("DEEPSEEK_API_KEY")
		_, err := NewClient(context.Background(), secretsClient)
		if err == nil {
			t.Error("expected error when DEEPSEEK_API_KEY is not set")
		}
	})

	t.Run("API key set", func(t *testing.T) {
		os.Setenv("DEEPSEEK_API_KEY", "test-key")
		client, err := NewClient(context.Background(), secretsClient)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if client == nil {
			t.Error("expected client to be initialized")
		}
	})
}

func TestDeepseekClient_Query(t *testing.T) {
	tests := []struct {
		name          string
		prompt        string
		schema        string
		mockResponse  *llms.ContentResponse
		mockError     error
		expectedValue string
		expectedError error
	}{
		{
			name:   "Successful query",
			prompt: "test prompt",
			schema: "{\"name\": \"test_tool\", \"description\": \"A test tool\", \"parameters\": {\"type\": \"object\", \"properties\": {}}}",
			mockResponse: &llms.ContentResponse{
				Choices: []*llms.ContentChoice{
					{
						ToolCalls: []llms.ToolCall{
							{
								FunctionCall: &llms.FunctionCall{
									Arguments: "{\"key\": \"value\"}",
								},
							},
						},
					},
				},
			},
			mockError:     nil,
			expectedValue: "{\"key\": \"value\"}",
			expectedError: nil,
		},
		{
			name:          "No tool call in response",
			prompt:        "test prompt",
			schema:        "{\"name\": \"test_tool\", \"description\": \"A test tool\", \"parameters\": {\"type\": \"object\", \"properties\": {}}}",
			mockResponse:  &llms.ContentResponse{Choices: []*llms.ContentChoice{}},
			mockError:     nil,
			expectedValue: "",
			expectedError: errors.New("no tool call found in LLM response"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLLM{
				generateContentFunc: func(context.Context, []llms.MessageContent, ...llms.CallOption) (*llms.ContentResponse, error) {
					return tt.mockResponse, tt.mockError
				},
			}
			client := NewDeepseek(mock)

			value, err := client.Query(context.Background(), tt.prompt, tt.schema)

			if value != tt.expectedValue {
				t.Errorf("unexpected value: got %q, want %q", value, tt.expectedValue)
			}

			if !cmp.Equal(err, tt.expectedError, cmp.Comparer(func(x, y error) bool {
				if x == nil || y == nil {
					return x == y
				}
				return x.Error() == y.Error()
			})) {
				t.Errorf("unexpected error: got %v, want %v", err, tt.expectedError)
			}
		})
	}
}
