package llm

import (
	"context"
	"errors"
	"testing"
)

type mockSecretsClient struct {
	getSecretFunc func(ctx context.Context, name string) (string, error)
}

func (m *mockSecretsClient) GetSecret(ctx context.Context, name string) (string, error) {
	return m.getSecretFunc(ctx, name)
}

func TestParseLLMConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("Valid JSON Config", func(t *testing.T) {
		resetLLMConfig()
		jsonConfig := `{"openai":"gpt-4","deepseek":"deepseek-chat"}`
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "LLM_CONFIG" {
					return jsonConfig, nil
				}
				return "", errors.New("not found")
			},
		}

		config, order, err := parseLLMConfig(ctx, mockSecrets)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(order) != 2 {
			t.Errorf("expected 2 providers, got %d", len(order))
		}
		if order[0] != "openai" || order[1] != "deepseek" {
			t.Errorf("unexpected order: %v", order)
		}
		if config["openai"] != "gpt-4" {
			t.Errorf("expected openai model gpt-4, got %s", config["openai"])
		}
	})

	t.Run("Legacy LLM_PROVIDERS", func(t *testing.T) {
		resetLLMConfig()
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "LLM_PROVIDERS" {
					return "openai,deepseek", nil
				}
				return "", errors.New("not found")
			},
		}

		config, order, err := parseLLMConfig(ctx, mockSecrets)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(order) != 2 {
			t.Errorf("expected 2 providers, got %d", len(order))
		}
		if config["openai"] != "" {
			t.Errorf("expected empty model for legacy config, got %s", config["openai"])
		}
	})

	t.Run("Default Config", func(t *testing.T) {
		resetLLMConfig()
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "", errors.New("not found")
			},
		}

		config, order, err := parseLLMConfig(ctx, mockSecrets)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(order) != 1 || order[0] != "deepseek" {
			t.Errorf("expected default deepseek provider, got %v", order)
		}
		if config["deepseek"] != "deepseek-chat" {
			t.Errorf("expected default model deepseek-chat, got %s", config["deepseek"])
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		resetLLMConfig()
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "LLM_CONFIG" {
					return "{invalid-json", nil
				}
				return "", errors.New("not found")
			},
		}

		_, _, err := parseLLMConfig(ctx, mockSecrets)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("Invalid JSON Type (Array)", func(t *testing.T) {
		resetLLMConfig()
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "LLM_CONFIG" {
					return "[]", nil
				}
				return "", errors.New("not found")
			},
		}

		_, _, err := parseLLMConfig(ctx, mockSecrets)
		if err == nil {
			t.Error("expected error for JSON array")
		}
	})

	t.Run("Invalid JSON Value Type", func(t *testing.T) {
		resetLLMConfig()
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "LLM_CONFIG" {
					return `{"openai": 123}`, nil
				}
				return "", errors.New("not found")
			},
		}

		_, _, err := parseLLMConfig(ctx, mockSecrets)
		if err == nil {
			t.Error("expected error for non-string value")
		}
	})

	t.Run("Empty JSON Object", func(t *testing.T) {
		resetLLMConfig()
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "LLM_CONFIG" {
					return "{}", nil
				}
				return "", errors.New("not found")
			},
		}

		config, order, err := parseLLMConfig(ctx, mockSecrets)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(config) != 0 || len(order) != 0 {
			t.Errorf("expected empty config and order, got %v, %v", config, order)
		}
	})

	t.Run("Invalid JSON Key Type", func(t *testing.T) {
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "LLM_CONFIG" {
					// json.Decoder.Token() will return json.Delim or number if key is not string
					// Note: standard JSON doesn't allow non-string keys, but we test the decoder's handling
					return `{123: "value"}`, nil
				}
				return "", errors.New("not found")
			},
		}

		_, _, err := parseLLMConfig(ctx, mockSecrets)
		if err == nil {
			t.Error("expected error for non-string key")
		}
	})

	t.Run("Malformed JSON After Valid Part", func(t *testing.T) {
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "LLM_CONFIG" {
					return `{"openai": "gpt-4", "deepseek"}`, nil
				}
				return "", errors.New("not found")
			},
		}

		_, _, err := parseLLMConfig(ctx, mockSecrets)
		if err == nil {
			t.Error("expected error for malformed JSON")
		}
	})

	t.Run("Not a JSON Object (String)", func(t *testing.T) {
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "LLM_CONFIG" {
					return `"just a string"`, nil
				}
				return "", errors.New("not found")
			},
		}

		_, _, err := parseLLMConfig(ctx, mockSecrets)
		if err == nil {
			t.Error("expected error for non-object JSON")
		}
	})

	t.Run("Duplicate Keys in JSON", func(t *testing.T) {
		resetLLMConfig()
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "LLM_CONFIG" {
					return `{"openai": "gpt-3.5", "openai": "gpt-4"}`, nil
				}
				return "", errors.New("not found")
			},
		}

		config, order, err := parseLLMConfig(ctx, mockSecrets)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if config["openai"] != "gpt-4" {
			t.Errorf("expected openai to be gpt-4, got %s", config["openai"])
		}
		if len(order) != 2 {
			// Depending on parser implementation.
			// json.Decoder token logic appends each key it sees to order.
			// "openai", "gpt-3.5", "openai", "gpt-4".
			// So order will be ["openai", "openai"].
			if len(order) != 2 || order[0] != "openai" || order[1] != "openai" {
				t.Errorf("unexpected order with duplicates: %v", order)
			}
		}
	})

	t.Run("Legacy LLM_PROVIDERS with Spaces and Extra Commas", func(t *testing.T) {
		resetLLMConfig()
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "LLM_PROVIDERS" {
					return " openai , , deepseek ", nil
				}
				return "", errors.New("not found")
			},
		}

		config, order, err := parseLLMConfig(ctx, mockSecrets)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// " openai ", "", " deepseek " (depending on Split behavior with consecutive delimiters)
		// strings.Split(" openai , , deepseek ", ",") -> [" openai ", " ", " deepseek "]

		if len(order) != 3 {
			t.Errorf("expected 3 elements, got %d: %v", len(order), order)
		}
		if _, ok := config[" openai "]; !ok {
			t.Errorf("expected key ' openai ' to exist")
		}
		if _, ok := config[" "]; !ok {
			t.Errorf("expected key ' ' to exist")
		}
		if _, ok := config[" deepseek "]; !ok {
			t.Errorf("expected key ' deepseek ' to exist")
		}
	})
}

func TestFallbackClient_Name(t *testing.T) {
	client := &FallbackClient{}
	if client.Name() != "fallback" {
		t.Errorf("expected name 'fallback', got '%s'", client.Name())
	}
}

func TestNewFallbackClientWithProviders(t *testing.T) {
	// Simple test to cover the helper function
	client := NewFallbackClientWithProviders(nil)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Name() != "fallback" {
		t.Errorf("expected name 'fallback', got '%s'", client.Name())
	}
}
