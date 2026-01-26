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

	// Use a fixed key for testing
	testKey := "integration-test-key"
	apiKeysJSON := `{"integration-test": "` + testKey + `"}`

	secretsClient := &mockSecretsClient{
		getSecretFunc: func(ctx context.Context, name string) (string, error) {
			if name == "API_KEYS" {
				return apiKeysJSON, nil
			}
			return "", errors.New("secret not found")
		},
	}

	authMiddleware := middleware.NewAuthMiddleware(secretsClient)

	providers := []string{"biblegateway", "biblehub"}

	for _, provider := range providers {
		t.Run("provider="+provider, func(t *testing.T) {
			os.Setenv("BIBLE_PROVIDER", provider)
			defer os.Unsetenv("BIBLE_PROVIDER")

			handler := NewQueryHandler(secretsClient)
			server := httptest.NewServer(middleware.Logging(authMiddleware.APIKeyAuth(handler)))
			defer server.Close()

			t.Run("verse query", func(t *testing.T) {
				reqBody := `{
					"query": {
						"verses": ["John 3:16"]
					}
				}`
				req, _ := http.NewRequest("POST", server.URL+"/query", bytes.NewBufferString(reqBody))
				req.Header.Set("X-API-KEY", testKey)

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
				req.Header.Set("X-API-KEY", testKey)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatalf("failed to make request: %v", err)
				}
				defer resp.Body.Close()

				// Expecting 500 because we don't have real LLM credentials in integration test environment usually
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
				req.Header.Set("X-API-KEY", testKey)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatalf("failed to make request: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
				}
			})
		})
	}
}
