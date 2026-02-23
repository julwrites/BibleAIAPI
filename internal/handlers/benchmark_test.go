package handlers

import (
	"bible-api-service/internal/bible"
	"bible-api-service/internal/llm"
	"bible-api-service/internal/secrets"
	"context"
	"os"
	"testing"
)

// BenchmarkGetLLMClient benchmarks the cost of retrieving the LLM client.
// This ensures that the client initialization is cached and not repeated per call.
func BenchmarkGetLLMClient(b *testing.B) {
	// Setup env for test to ensure NewFallbackClient succeeds
	os.Setenv("LLM_PROVIDERS", "openai")
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("LLM_PROVIDERS")
	defer os.Unsetenv("OPENAI_API_KEY")

	secretsClient := &secrets.EnvClient{}

	// Baseline: Cost of creating a new client every time (simulating old behavior)
	b.Run("NewClientPerCall", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := llm.NewFallbackClient(context.Background(), secretsClient)
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
	})

	// Benchmark the actual NewQueryHandler implementation (Cached)
	b.Run("CachedClient", func(b *testing.B) {
		// Create handler (expensive init, done once)
		handler := NewQueryHandler(secretsClient, &bible.VersionManager{})

		b.ResetTimer() // Reset timer to exclude setup time

		for i := 0; i < b.N; i++ {
			if _, err := handler.GetLLMClient(); err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
	})
}
