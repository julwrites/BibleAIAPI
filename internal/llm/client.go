package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"bible-api-service/internal/llm/deepseek"
	"bible-api-service/internal/llm/gemini"
	"bible-api-service/internal/llm/openai"
	"bible-api-service/internal/llm/openapicustom"
	"bible-api-service/internal/llm/openrouter"
	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/secrets"
)

// FallbackClient is a client that tries a list of providers in order until one succeeds.
type FallbackClient struct {
	clients []provider.LLMClient
}

// parseLLMConfig parses the LLM configuration from environment variable or secret.
// It returns a map of provider name to model name.
func parseLLMConfig(ctx context.Context, secretsClient secrets.Client) (map[string]string, error) {
	configJSON, err := secrets.Get(ctx, secretsClient, "LLM_CONFIG")
	if err != nil {
		// Fall back to LLM_PROVIDERS for backward compatibility
		providerNames, err := secrets.Get(ctx, secretsClient, "LLM_PROVIDERS")
		if err != nil {
			log.Printf("LLM_CONFIG and LLM_PROVIDERS not set, defaulting to {\"deepseek\":\"deepseek-chat\"}: %v", err)
			return map[string]string{"deepseek": "deepseek-chat"}, nil
		}
		log.Printf("WARNING: LLM_PROVIDERS is deprecated, use LLM_CONFIG JSON instead")
		// Convert comma-separated list to map with empty model names
		providers := strings.Split(providerNames, ",")
		config := make(map[string]string, len(providers))
		for _, p := range providers {
			config[p] = "" // Empty model name will use provider default
		}
		return config, nil
	}

	var config map[string]string
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("invalid LLM_CONFIG JSON: %w", err)
	}
	return config, nil
}

// NewFallbackClient creates a new FallbackClient with the providers specified in the LLM_CONFIG JSON or LLM_PROVIDERS environment variable/secret.
func NewFallbackClient(ctx context.Context, secretsClient secrets.Client) (*FallbackClient, error) {
	providerConfig, err := parseLLMConfig(ctx, secretsClient)
	if err != nil {
		return nil, err
	}

	// Get sorted provider names for deterministic order
	providerNames := make([]string, 0, len(providerConfig))
	for name := range providerConfig {
		providerNames = append(providerNames, name)
	}
	sort.Strings(providerNames)

	clients := make([]provider.LLMClient, 0, len(providerConfig))
	var configErrors []string

	for _, providerName := range providerNames {
		modelName := providerConfig[providerName]
		var client provider.LLMClient
		var err error

		switch providerName {
		case "openai":
			client, err = openai.NewClient(ctx, secretsClient, modelName)
		case "openai-custom":
			client, err = openapicustom.NewClient(ctx, secretsClient, modelName)
		case "deepseek":
			client, err = deepseek.NewClient(ctx, secretsClient, modelName)
		case "gemini":
			client, err = gemini.NewClient(ctx, secretsClient, modelName)
		case "openrouter":
			client, err = openrouter.NewClient(ctx, secretsClient, modelName)
		default:
			log.Printf("WARNING: unsupported provider '%s' in LLM_CONFIG, skipping", providerName)
			continue
		}

		if err == nil && client != nil {
			clients = append(clients, client)
		} else if err != nil {
			log.Printf("Failed to initialize provider '%s': %v", providerName, err)
			configErrors = append(configErrors, fmt.Sprintf("%s: %v", providerName, err))
		} else {
			log.Printf("Failed to initialize provider '%s': unknown error", providerName)
			configErrors = append(configErrors, fmt.Sprintf("%s: failed to initialize client (unknown error)", providerName))
		}
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("no valid LLM clients could be created. Errors: %s", strings.Join(configErrors, "; "))
	}

	return &FallbackClient{clients: clients}, nil
}

// Query tries each client in order until one succeeds.
func (c *FallbackClient) Query(ctx context.Context, prompt string, schema string) (string, error) {
	var lastErr error

	for _, client := range c.clients {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Minute)

		result, err := client.Query(ctxWithTimeout, prompt, schema)
		cancel()
		if err == nil {
			return result, nil
		}
		log.Printf("Provider failed: %v", err)
		lastErr = err
	}

	return "", fmt.Errorf("all LLM providers failed: %w", lastErr)
}
