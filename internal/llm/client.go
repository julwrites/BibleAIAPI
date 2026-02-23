package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
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
	clients    []provider.LLMClient
	clientsMap map[string]provider.LLMClient
}

var (
	cachedLLMConfig    map[string]string
	cachedLLMOrder     []string
	cachedLLMConfigErr error
	parseLLMConfigOnce sync.Once
)

// parseLLMConfig parses the LLM configuration from environment variable or secret.
// It returns a map of provider name to model name and a slice of provider names in the order they appear.
func parseLLMConfig(ctx context.Context, secretsClient secrets.Client) (map[string]string, []string, error) {
	parseLLMConfigOnce.Do(func() {
		configJSON, err := secrets.Get(ctx, secretsClient, "LLM_CONFIG")
		if err != nil {
			// Fall back to LLM_PROVIDERS for backward compatibility
			providerNames, err := secrets.Get(ctx, secretsClient, "LLM_PROVIDERS")
			if err != nil {
				log.Printf("LLM_CONFIG and LLM_PROVIDERS not set, defaulting to {\"deepseek\":\"deepseek-chat\"}: %v", err)
				cachedLLMConfig = map[string]string{"deepseek": "deepseek-chat"}
				cachedLLMOrder = []string{"deepseek"}
				cachedLLMConfigErr = nil
				return
			}
			log.Printf("WARNING: LLM_PROVIDERS is deprecated, use LLM_CONFIG JSON instead")
			// Convert comma-separated list to map with empty model names
			providers := strings.Split(providerNames, ",")
			config := make(map[string]string, len(providers))
			for _, p := range providers {
				config[p] = "" // Empty model name will use provider default
			}
			cachedLLMConfig = config
			cachedLLMOrder = providers
			cachedLLMConfigErr = nil
			return
		}

		// Parse JSON while preserving key order
		decoder := json.NewDecoder(strings.NewReader(configJSON))
		decoder.UseNumber()

		// Read opening brace
		token, err := decoder.Token()
		if err != nil {
			cachedLLMConfigErr = fmt.Errorf("invalid LLM_CONFIG JSON: %w", err)
			return
		}
		if delim, ok := token.(json.Delim); !ok || delim != '{' {
			cachedLLMConfigErr = fmt.Errorf("invalid LLM_CONFIG JSON: expected object")
			return
		}

		config := make(map[string]string)
		var order []string

		for decoder.More() {
			// Read key
			token, err := decoder.Token()
			if err != nil {
				cachedLLMConfigErr = fmt.Errorf("invalid LLM_CONFIG JSON: %w", err)
				return
			}
			key, ok := token.(string)
			if !ok {
				cachedLLMConfigErr = fmt.Errorf("invalid LLM_CONFIG JSON: key must be string")
				return
			}

			// Read value
			token, err = decoder.Token()
			if err != nil {
				cachedLLMConfigErr = fmt.Errorf("invalid LLM_CONFIG JSON: %w", err)
				return
			}
			value, ok := token.(string)
			if !ok {
				cachedLLMConfigErr = fmt.Errorf("invalid LLM_CONFIG JSON: value must be string for key %q", key)
				return
			}

			config[key] = value
			order = append(order, key)
		}

		// Read closing brace
		token, err = decoder.Token()
		if err != nil {
			cachedLLMConfigErr = fmt.Errorf("invalid LLM_CONFIG JSON: %w", err)
			return
		}
		if delim, ok := token.(json.Delim); !ok || delim != '}' {
			cachedLLMConfigErr = fmt.Errorf("invalid LLM_CONFIG JSON: malformed object")
			return
		}

		cachedLLMConfig = config
		cachedLLMOrder = order
		cachedLLMConfigErr = nil
	})

	return cachedLLMConfig, cachedLLMOrder, cachedLLMConfigErr
}

// NewFallbackClient creates a new FallbackClient with the providers specified in the LLM_CONFIG JSON or LLM_PROVIDERS environment variable/secret.
func NewFallbackClient(ctx context.Context, secretsClient secrets.Client) (*FallbackClient, error) {
	providerConfig, providerOrder, err := parseLLMConfig(ctx, secretsClient)
	if err != nil {
		return nil, err
	}

	clientsMap := make(map[string]provider.LLMClient)
	log.Printf("LLM provider order: %v", providerOrder)

	clients := make([]provider.LLMClient, 0, len(providerConfig))
	var configErrors []string

	for _, providerName := range providerOrder {
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
			if modelName != "" {
				log.Printf("Successfully initialized provider '%s' with model '%s'", providerName, modelName)
			} else {
				log.Printf("Successfully initialized provider '%s' with default model", providerName)
			}
			clients = append(clients, client)
			clientsMap[client.Name()] = client
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

	return &FallbackClient{clients: clients, clientsMap: clientsMap}, nil
}

// NewFallbackClientWithProviders creates a new FallbackClient with the given providers.
// This is primarily used for testing.
func NewFallbackClientWithProviders(clients []provider.LLMClient) *FallbackClient {
	clientsMap := make(map[string]provider.LLMClient)
	for _, client := range clients {
		clientsMap[client.Name()] = client
	}
	return &FallbackClient{clients: clients, clientsMap: clientsMap}
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
		providerType := reflect.TypeOf(client).String()
		log.Printf("Attempting LLM provider: %s", providerType)

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

// resetLLMConfig resets the cached LLM configuration. This is used for testing purposes.
func resetLLMConfig() {
	cachedLLMConfig = nil
	cachedLLMOrder = nil
	cachedLLMConfigErr = nil
	parseLLMConfigOnce = sync.Once{}
}
