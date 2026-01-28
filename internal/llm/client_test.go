package llm

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/secrets"
)

// mockLLMClient is a mock implementation of the LLMClient interface for testing.
type mockLLMClient struct {
	queryFunc func(ctx context.Context, prompt string, schema string) (string, string, error)
	streamFunc func(ctx context.Context, prompt string) (<-chan string, string, error)
	nameFunc func() string
}

func (m *mockLLMClient) Query(ctx context.Context, prompt string, schema string) (string, string, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, prompt, schema)
	}
	return "", "", errors.New("queryFunc not implemented")
}

func (m *mockLLMClient) Stream(ctx context.Context, prompt string) (<-chan string, string, error) {
	if m.streamFunc != nil {
		return m.streamFunc(ctx, prompt)
	}
	return nil, "", errors.New("streamFunc not implemented")
}

func (m *mockLLMClient) Name() string {
	if m.nameFunc != nil {
		return m.nameFunc()
	}
	return "mock"
}

func TestNewFallbackClient(t *testing.T) {
	secretsClient := &secrets.EnvClient{}

	t.Run("LLM_PROVIDERS not set", func(t *testing.T) {
		os.Unsetenv("LLM_PROVIDERS")
		os.Setenv("DEEPSEEK_API_KEY", "dummy-key")
		defer os.Unsetenv("DEEPSEEK_API_KEY")

		client, err := NewFallbackClient(context.Background(), secretsClient)
		if err != nil {
			t.Errorf("unexpected error when LLM_PROVIDERS is not set: %v", err)
		}
		if client == nil {
			t.Error("expected client to be initialized with default provider")
		}
		if len(client.clients) != 1 {
			t.Errorf("expected 1 client (default), got %d", len(client.clients))
		}
		if len(client.clientsMap) != 1 {
			t.Errorf("expected 1 client in map, got %d", len(client.clientsMap))
		}
	})

	t.Run("Unsupported provider", func(t *testing.T) {
		os.Setenv("LLM_PROVIDERS", "unsupported")
		defer os.Unsetenv("LLM_PROVIDERS")
		_, err := NewFallbackClient(context.Background(), secretsClient)
		if err == nil {
			t.Error("expected error for unsupported provider")
		}
		if err.Error() != "no valid LLM clients could be created. Errors: " {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("All providers fail initialization", func(t *testing.T) {
		os.Setenv("LLM_PROVIDERS", "openai")
		// Not setting OPENAI_API_KEY should cause failure
		defer os.Unsetenv("LLM_PROVIDERS")

		_, err := NewFallbackClient(context.Background(), secretsClient)
		if err == nil {
			t.Error("expected error when all providers fail")
		}
		expectedPart := "openai: OPENAI_API_KEY secret or environment variable not set"
		if err != nil && !strings.Contains(err.Error(), expectedPart) {
			t.Errorf("expected error message to contain %q, got %q", expectedPart, err.Error())
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
		if len(client.clientsMap) != 1 {
			t.Errorf("expected 1 client in map, got %d", len(client.clientsMap))
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
		if len(client.clientsMap) != 1 {
			t.Errorf("expected 1 client in map, got %d", len(client.clientsMap))
		}
	})
}

func TestFallbackClient_Query_Preference(t *testing.T) {
	client1 := &mockLLMClient{
		nameFunc: func() string { return "client1" },
		queryFunc: func(ctx context.Context, prompt, schema string) (string, string, error) {
			return "response1", "client1", nil
		},
	}
	client2 := &mockLLMClient{
		nameFunc: func() string { return "client2" },
		queryFunc: func(ctx context.Context, prompt, schema string) (string, string, error) {
			return "response2", "client2", nil
		},
	}

	clients := []provider.LLMClient{client1, client2}
	clientsMap := map[string]provider.LLMClient{
		"client1": client1,
		"client2": client2,
	}

	fc := &FallbackClient{clients: clients, clientsMap: clientsMap}

	// Case 1: Prefer client2
	ctx := context.WithValue(context.Background(), provider.PreferredProviderKey, "client2")
	_, name, err := fc.Query(ctx, "prompt", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "client2" {
		t.Errorf("expected client2, got %s", name)
	}

	// Case 2: Prefer client1
	ctx = context.WithValue(context.Background(), provider.PreferredProviderKey, "client1")
	_, name, err = fc.Query(ctx, "prompt", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "client1" {
		t.Errorf("expected client1, got %s", name)
	}

	// Case 3: Prefer non-existent client (fallback to order)
	ctx = context.WithValue(context.Background(), provider.PreferredProviderKey, "client3")
	_, name, err = fc.Query(ctx, "prompt", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "client1" { // Default order starts with client1
		t.Errorf("expected client1, got %s", name)
	}

	// Case 4: Prefer client2, but client2 fails (fallback to others)
	client2Fail := &mockLLMClient{
		nameFunc: func() string { return "client2" },
		queryFunc: func(ctx context.Context, prompt, schema string) (string, string, error) {
			return "", "", errors.New("fail")
		},
	}

	clientsFail := []provider.LLMClient{client1, client2Fail}
	clientsMapFail := map[string]provider.LLMClient{"client1": client1, "client2": client2Fail}
	fcFail := &FallbackClient{clients: clientsFail, clientsMap: clientsMapFail}

	ctx = context.WithValue(context.Background(), provider.PreferredProviderKey, "client2")
	_, name, err = fcFail.Query(ctx, "prompt", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "client1" {
		t.Errorf("expected fallback to client1, got %s", name)
	}
}

func TestFallbackClient_Query(t *testing.T) {
	tests := []struct {
		name           string
		clients        []provider.LLMClient
		prompt         string
		schema         string
		expectedResult string
		expectedName   string
		expectedError  error
	}{
		{
			name: "First client succeeds",
			clients: []provider.LLMClient{
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, string, error) {
					return "success from client 1", "mock1", nil
				}},
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, string, error) {
					return "", "", errors.New("client 2 should not be called")
				}},
			},
			expectedResult: "success from client 1",
			expectedName:   "mock1",
			expectedError:  nil,
		},
		{
			name: "Fallback to second client",
			clients: []provider.LLMClient{
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, string, error) {
					return "", "", errors.New("client 1 fails")
				}},
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, string, error) {
					return "success from client 2", "mock2", nil
				}},
			},
			expectedResult: "success from client 2",
			expectedName:   "mock2",
			expectedError:  nil,
		},
		{
			name: "All clients fail",
			clients: []provider.LLMClient{
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, string, error) {
					return "", "", errors.New("client 1 fails")
				}},
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, string, error) {
					return "", "", errors.New("client 2 fails")
				}},
			},
			expectedResult: "",
			expectedName:   "",
			expectedError:  errors.New("all LLM providers failed: client 2 fails"),
		},
		{
			name: "First client times out",
			clients: []provider.LLMClient{
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, string, error) {
					time.Sleep(10 * time.Millisecond)
					return "", "", errors.New("should have timed out")
				}},
				&mockLLMClient{queryFunc: func(ctx context.Context, prompt, schema string) (string, string, error) {
					return "success from client 2", "mock2", nil
				}},
			},
			expectedResult: "success from client 2",
			expectedName:   "mock2",
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fallbackClient := &FallbackClient{clients: tt.clients}
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			result, name, err := fallbackClient.Query(ctx, tt.prompt, tt.schema)
			if result != tt.expectedResult {
				t.Errorf("unexpected result: got %q, want %q", result, tt.expectedResult)
			}
			if name != tt.expectedName {
				t.Errorf("unexpected name: got %q, want %q", name, tt.expectedName)
			}

			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) || (err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error()) {
				t.Errorf("unexpected error: got %v, want %v", err, tt.expectedError)
			}
		})
	}
}

func TestFallbackClient_Stream_Preference(t *testing.T) {
	client1 := &mockLLMClient{
		nameFunc: func() string { return "client1" },
		streamFunc: func(ctx context.Context, prompt string) (<-chan string, string, error) {
			ch := make(chan string, 1)
			ch <- "response1"
			close(ch)
			return ch, "client1", nil
		},
	}
	client2 := &mockLLMClient{
		nameFunc: func() string { return "client2" },
		streamFunc: func(ctx context.Context, prompt string) (<-chan string, string, error) {
			ch := make(chan string, 1)
			ch <- "response2"
			close(ch)
			return ch, "client2", nil
		},
	}

	clients := []provider.LLMClient{client1, client2}
	clientsMap := map[string]provider.LLMClient{
		"client1": client1,
		"client2": client2,
	}

	fc := &FallbackClient{clients: clients, clientsMap: clientsMap}

	// Case 1: Prefer client2
	ctx := context.WithValue(context.Background(), provider.PreferredProviderKey, "client2")
	_, name, err := fc.Stream(ctx, "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "client2" {
		t.Errorf("expected client2, got %s", name)
	}

	// Case 2: Prefer client1
	ctx = context.WithValue(context.Background(), provider.PreferredProviderKey, "client1")
	_, name, err = fc.Stream(ctx, "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "client1" {
		t.Errorf("expected client1, got %s", name)
	}

	// Case 3: Prefer non-existent client (fallback to order)
	ctx = context.WithValue(context.Background(), provider.PreferredProviderKey, "client3")
	_, name, err = fc.Stream(ctx, "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "client1" { // Default order starts with client1
		t.Errorf("expected client1, got %s", name)
	}

	// Case 4: Prefer client2, but client2 fails (fallback to others)
	client2Fail := &mockLLMClient{
		nameFunc: func() string { return "client2" },
		streamFunc: func(ctx context.Context, prompt string) (<-chan string, string, error) {
			return nil, "", errors.New("fail")
		},
	}

	clientsFail := []provider.LLMClient{client1, client2Fail}
	clientsMapFail := map[string]provider.LLMClient{"client1": client1, "client2": client2Fail}
	fcFail := &FallbackClient{clients: clientsFail, clientsMap: clientsMapFail}

	ctx = context.WithValue(context.Background(), provider.PreferredProviderKey, "client2")
	_, name, err = fcFail.Stream(ctx, "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "client1" {
		t.Errorf("expected fallback to client1, got %s", name)
	}
}
