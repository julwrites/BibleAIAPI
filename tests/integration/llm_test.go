//go:build integration_llm

package integration

import (
	"bible-api-service/internal/llm"
	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/secrets"
	"bible-api-service/tests/mocks"
	"context"
	"os"
	"testing"
)

func TestLLMProvidersIntegration(t *testing.T) {
	// For integration tests, we assume secrets are available via environment variables or a local secrets client.
	// We'll use the real secrets client logic which usually defaults to env vars if GCP is not configured or fails.
	ctx := context.Background()
	secretsClient, err := secrets.NewClient(ctx, "") // Empty project ID implies local/env usually
	if err != nil {
		t.Fatalf("Failed to create secrets client: %v", err)
	}

	// Check if any LLM configuration is present
	hasConfig := os.Getenv("LLM_CONFIG") != "" ||
		os.Getenv("OPENAI_API_KEY") != "" ||
		os.Getenv("GEMINI_API_KEY") != "" ||
		os.Getenv("DEEPSEEK_API_KEY") != ""

	var client provider.LLMClient

	if !hasConfig {
		t.Log("No LLM configuration found (LLM_CONFIG, OPENAI_API_KEY, GEMINI_API_KEY, DEEPSEEK_API_KEY). Running with MOCK client.")

		mockClient := &mocks.MockLLMClient{
			Response: "Yes",
		}
		// Wrap in FallbackClient to test the structure, even if it's just one mock
		client = llm.NewFallbackClientWithProviders([]provider.LLMClient{mockClient})
	} else {
		t.Log("LLM configuration found. Running integration test with REAL providers.")

		// Create the real client
		realClient, err := llm.NewFallbackClient(ctx, secretsClient)
		if err != nil {
			t.Fatalf("Failed to create LLM client: %v", err)
		}
		client = realClient
	}

	response, _, err := client.Query(ctx, "Hello, can you hear me? Reply with 'Yes'.", "")
	if err != nil {
		t.Fatalf("LLM Query failed: %v", err)
	}

	if response == "" {
		t.Error("LLM returned empty response")
	}
	t.Logf("LLM Response: %s", response)
}
