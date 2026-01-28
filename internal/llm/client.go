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
	clients    []provider.LLMClient
	clientsMap map[string]provider.LLMClient
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
	clientsMap := make(map[string]provider.LLMClient)

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
			clientsMap[client.Name()] = client
		} else if err != nil {
			log.Printf("Failed to initialize provider '%s': %v", p, err)
			configErrors = append(configErrors, fmt.Sprintf("%s: %v", p, err))
		} else {
			log.Printf("Failed to initialize provider '%s': unknown error", p)
			configErrors = append(configErrors, fmt.Sprintf("%s: failed to initialize client (unknown error)", p))
		}
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("no valid LLM clients could be created. Errors: %s", strings.Join(configErrors, "; "))
	}

	return &FallbackClient{clients: clients, clientsMap: clientsMap}, nil
}

// Query tries each client in order until one succeeds.
func (c *FallbackClient) Query(ctx context.Context, prompt string, schema string) (string, string, error) {
	var lastErr error

	preferredName, _ := ctx.Value(provider.PreferredProviderKey).(string)
	triedPreferred := false

	if preferredName != "" {
		if client, ok := c.clientsMap[preferredName]; ok {
			ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Minute)
			result, providerName, err := client.Query(ctxWithTimeout, prompt, schema)
			cancel()
			if err == nil {
				return result, providerName, nil
			}
			log.Printf("Preferred provider %s failed: %v", client.Name(), err)
			lastErr = err
			triedPreferred = true
		}
	}

	for _, client := range c.clients {
		if triedPreferred && client.Name() == preferredName {
			continue
		}
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Minute)

		result, providerName, err := client.Query(ctxWithTimeout, prompt, schema)
		cancel()
		if err == nil {
			return result, providerName, nil
		}
		log.Printf("Provider %s failed: %v", client.Name(), err)
		lastErr = err
	}

	return "", "", fmt.Errorf("all LLM providers failed: %w", lastErr)
}

// Stream tries each client in order until one succeeds.
func (c *FallbackClient) Stream(ctx context.Context, prompt string) (<-chan string, string, error) {
	var lastErr error

	preferredName, _ := ctx.Value(provider.PreferredProviderKey).(string)
	triedPreferred := false

	if preferredName != "" {
		if client, ok := c.clientsMap[preferredName]; ok {
			ch, providerName, err := client.Stream(ctx, prompt)
			if err == nil {
				return ch, providerName, nil
			}
			log.Printf("Preferred provider %s stream failed: %v", client.Name(), err)
			lastErr = err
			triedPreferred = true
		}
	}

	for _, client := range c.clients {
		if triedPreferred && client.Name() == preferredName {
			continue
		}
		ch, providerName, err := client.Stream(ctx, prompt)
		if err == nil {
			return ch, providerName, nil
		}
		log.Printf("Provider %s stream failed: %v", client.Name(), err)
		lastErr = err
	}

	return nil, "", fmt.Errorf("all LLM providers failed to stream: %w", lastErr)
}

// Name returns the name of the client.
func (c *FallbackClient) Name() string {
	return "fallback"
}
