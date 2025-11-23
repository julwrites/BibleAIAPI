package llm

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/secrets"
)

// mockLLMClient is a mock implementation of the LLMClient interface for testing.
type mockLLMClient struct {
	queryFunc func(ctx context.Context, prompt string, schema string) (string, error)
}

func (m *mockLLMClient) Query(ctx context.Context, prompt string, schema string) (string, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, prompt, schema)
	}
	return "", errors.New("queryFunc not implemented")
}

func TestNewFallbackClient(t *testing.T) {
	secretsClient := &secrets.EnvClient{}

	t.Run("LLM_PROVIDERS not set", func(t *testing.T) {
		os.Unsetenv("LLM_PROVIDERS")
		_, err := NewFallbackClient(context.Background(), secretsClient)
		if err == nil {
			t.Error("expected error when LLM_PROVIDERS is not set")
		}
	})

	t.Run("Unsupported provider", func(t *testing.T) {
		os.Setenv("LLM_PROVIDERS", "unsupported")
		defer os.Unsetenv("LLM_PROVIDERS")
		_, err := NewFallbackClient(context.Background(), secretsClient)
		if err == nil {
			t.Error("expected error for unsupported provider")
		}
	})

	t.Run("Valid provider", func(t *testing.T) {
		os.Setenv("LLM_PROVIDERS", "openai")
		os.Setenv("OPENAI_API_KEY", "test-key")
		defer os.Unsetenv("LLM_PROVIDERS")
		defer os.Unsetenv("OPENAI_API_KEY")
		client, err := NewFallbackClient(context.Background(), secretsClient)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if client == nil {
			t.Error("expected client to be initialized")
		}
		if len(client.clients) != 1 {
			t.Errorf("expected 1 client, got %d", len(client.clients))
		}
	})

	t.Run("Mixed providers", func(t *testing.T) {
		os.Setenv("LLM_PROVIDERS", "openai,unsupported")
		os.Setenv("OPENAI_API_KEY", "test-key")
		defer os.Unsetenv("LLM_PROVIDERS")
		defer os.Unsetenv("OPENAI_API_KEY")
		client, err := NewFallbackClient(context.Background(), secretsClient)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if client == nil {
			t.Error("expected client to be initialized")
		}
		if len(client.clients) != 1 {
			t.Errorf("expected 1 client, got %d", len(client.clients))
		}
	})
}

func TestFallbackClient_Query(t *testing.T) {
	tests := []struct {
		name           string
		clients        []provider.LLMClient
		prompt         string
		schema         string
		expectedResult string
		expectedError  error
	}{
		{
			name: "First client succeeds",
			clients: []provider.LLMClient{
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, error) {
					return "success from client 1", nil
				}},
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, error) {
					return "", errors.New("client 2 should not be called")
				}},
			},
			expectedResult: "success from client 1",
			expectedError:  nil,
		},
		{
			name: "Fallback to second client",
			clients: []provider.LLMClient{
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, error) {
					return "", errors.New("client 1 fails")
				}},
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, error) {
					return "success from client 2", nil
				}},
			},
			expectedResult: "success from client 2",
			expectedError:  nil,
		},
		{
			name: "All clients fail",
			clients: []provider.LLMClient{
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, error) {
					return "", errors.New("client 1 fails")
				}},
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, error) {
					return "", errors.New("client 2 fails")
				}},
			},
			expectedResult: "",
			expectedError:  errors.New("all LLM providers failed: client 2 fails"),
		},
		{
			name: "First client times out",
			clients: []provider.LLMClient{
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, error) {
					time.Sleep(10 * time.Millisecond)
					return "", errors.New("should have timed out")
				}},
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, error) {
					return "success from client 2", nil
				}},
			},
			expectedResult: "success from client 2",
			expectedError:  nil,
		},
		{
			name: "Fallback to deepseek",
			clients: []provider.LLMClient{
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, error) {
					return "", errors.New("openai-custom fails")
				}},
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, error) {
					return "success from deepseek", nil
				}},
			},
			expectedResult: "success from deepseek",
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fallbackClient := &FallbackClient{clients: tt.clients}
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			result, err := fallbackClient.Query(ctx, tt.prompt, tt.schema)
			if result != tt.expectedResult {
				t.Errorf("unexpected result: got %q, want %q", result, tt.expectedResult)
			}

			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) || (err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error()) {
				t.Errorf("unexpected error: got %v, want %v", err, tt.expectedError)
			}
		})
	}
}
