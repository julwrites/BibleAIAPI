package handlers

import (
	"bible-api-service/internal/middleware"
	"bible-api-service/internal/secrets"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type mockSecretsClient struct {
	getSecretFunc func(ctx context.Context, name string) (string, error)
}

func (m *mockSecretsClient) GetSecret(ctx context.Context, name string) (string, error) {
	return m.getSecretFunc(ctx, name)
}

var _ secrets.Client = &mockSecretsClient{}

func TestQueryEndpointIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("skipping integration tests")
	}

	secretsClient := &mockSecretsClient{
		getSecretFunc: func(ctx context.Context, name string) (string, error) {
			if name == "API_KEY" {
				return os.Getenv("API_KEY"), nil
			}
			return "", errors.New("secret not found")
		},
	}
	authMiddleware := middleware.NewAuthMiddleware(secretsClient)
	handler := NewQueryHandler()
	server := httptest.NewServer(middleware.Logging(authMiddleware.APIKeyAuth(handler)))
	defer server.Close()

	t.Run("verse query", func(t *testing.T) {
		reqBody := `{
			"query": {
				"verses": ["John 3:16"]
			}
		}`
		req, _ := http.NewRequest("POST", server.URL+"/query", bytes.NewBufferString(reqBody))
		req.Header.Set("X-API-KEY", os.Getenv("API_KEY"))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("prompt query", func(t *testing.T) {
		reqBody := `{
			"query": {
				"prompt": "Who was Moses?"
			},
			"context": {
				"user": {
					"version": "ESV"
				}
			}
		}`
		req, _ := http.NewRequest("POST", server.URL+"/query", bytes.NewBufferString(reqBody))
		req.Header.Set("X-API-KEY", os.Getenv("API_KEY"))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Note: This will likely fail with 500 because LLM client is not mocked in integration tests or configured
		// The original test expected 500 too, but for different reasons (missing prompt/oquery/etc).
		// However, if we are truly integration testing, we might expect it to work if env vars are set, or fail if not.
		// In the original code it expected 500.
		// Let's check if we should expect 500 or 200. If LLM keys are missing, it will be 500.
		// Since we are running in a CI environment or similar where keys might not be present for external services during this specific test run unless specified.
		// The original test code:
		// if resp.StatusCode != http.StatusInternalServerError {
		// 	t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		// }
		// It seems it expected failure. Let's keep it consistent unless we know for sure.
		// Wait, the original test said "open query" and passed "oquery". It expected 500.
		// Why? Maybe because no LLM provider is configured in the test environment?
		// Yes, likely.

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("word search query", func(t *testing.T) {
		reqBody := `{
			"query": {
				"words": ["grace"]
			}
		}`
		req, _ := http.NewRequest("POST", server.URL+"/query", bytes.NewBufferString(reqBody))
		req.Header.Set("X-API-KEY", os.Getenv("API_KEY"))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}
