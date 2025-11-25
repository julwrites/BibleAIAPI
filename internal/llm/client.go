package llm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"bible-api-service/internal/llm/deepseek"
	"bible-api-service/internal/llm/gemini"
	"bible-api-service/internal/llm/openai"
	"bible-api-service/internal/llm/openapicustom"
	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/secrets"
)

// FallbackClient is a client that tries a list of providers in order until one succeeds.
type FallbackClient struct {
	clients []provider.LLMClient
}

// NewFallbackClient creates a new FallbackClient with the providers specified in the LLM_PROVIDERS environment variable or secret.
func NewFallbackClient(ctx context.Context, secretsClient secrets.Client) (*FallbackClient, error) {
	providerNames, err := secrets.Get(ctx, secretsClient, "LLM_PROVIDERS")
	if err != nil {
		log.Printf("LLM_PROVIDERS secret or environment variable not set, defaulting to 'deepseek': %v", err)
		providerNames = "deepseek"
	}

	providers := strings.Split(providerNames, ",")
	clients := make([]provider.LLMClient, 0, len(providers))

	var configErrors []string

	for _, p := range providers {
		var client provider.LLMClient
		var err error

		switch p {
		case "openai":
			client, err = openai.NewClient(ctx, secretsClient)
		case "openai-custom":
			client, err = openapicustom.NewClient(ctx, secretsClient)
		case "deepseek":
			client, err = deepseek.NewClient(ctx, secretsClient)
		case "gemini":
			client, err = gemini.NewClient(ctx, secretsClient)
		default:
			// Optionally log a warning for unsupported providers
			continue
		}

		if err == nil && client != nil {
			clients = append(clients, client)
		} else if err != nil {
			configErrors = append(configErrors, fmt.Sprintf("%s: %v", p, err))
		} else {
			configErrors = append(configErrors, fmt.Sprintf("%s: failed to initialize client (unknown error)", p))
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
		lastErr = err
	}

	return "", fmt.Errorf("all LLM providers failed: %w", lastErr)
}
