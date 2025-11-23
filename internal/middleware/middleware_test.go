package middleware

import (
	"bible-api-service/internal/secrets"
	"bible-api-service/internal/storage"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockSecretsClient struct {
	getSecretFunc func(ctx context.Context, name string) (string, error)
}

func (m *mockSecretsClient) GetSecret(ctx context.Context, name string) (string, error) {
	return m.getSecretFunc(ctx, name)
}

var _ secrets.Client = &mockSecretsClient{}

func TestAPIKeyAuth(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("no api key configured and no db key", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "", errors.New("secret not found")
			},
		}
		storageClient := storage.NewMockClient()

		authMiddleware := NewAuthMiddleware(secretsClient, storageClient)
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)

		// If legacy key fails, and no header provided?
		// Code checks header first. If empty -> 401.
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})

	t.Run("bypass local dev if secret fails and key provided matches legacy logic?", func(t *testing.T) {
		// Existing logic: If secret fetch fails, BYPASS check.
		// "Log and allow".
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "", errors.New("secret not found")
			},
		}
		storageClient := storage.NewMockClient()

		authMiddleware := NewAuthMiddleware(secretsClient, storageClient)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-KEY", "anything") // Must provide some key to pass check "if clientKey != apiKey" (which is bypassed here)

		// Actually, in the new logic:
		// 1. Check Firestore. Returns err/nil.
		// 2. Check Legacy Secret. Returns error.
		// 3. Log warning and ServeHTTP.

		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("valid legacy api key", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "legacykey", nil
			},
		}
		storageClient := storage.NewMockClient() // Empty DB

		authMiddleware := NewAuthMiddleware(secretsClient, storageClient)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-KEY", "legacykey")
		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("valid db api key", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "legacykey", nil
			},
		}
		storageClient := storage.NewMockClient()
		// Inject a key directly
		key, _ := storageClient.CreateAPIKey(context.Background(), "testclient", 100)

		authMiddleware := NewAuthMiddleware(secretsClient, storageClient)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-KEY", key.Key)

		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}

		// Verify usage incremented
		keyAfter, _ := storageClient.GetAPIKey(context.Background(), key.Key)
		if keyAfter.RequestCount != 1 {
			t.Errorf("expected count 1, got %d", keyAfter.RequestCount)
		}
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "legacykey", nil
			},
		}
		storageClient := storage.NewMockClient()
		key, _ := storageClient.CreateAPIKey(context.Background(), "testclient", 1) // Limit 1

		authMiddleware := NewAuthMiddleware(secretsClient, storageClient)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-KEY", key.Key)

		// 1st request -> OK
		rr1 := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr1, req)
		if rr1.Code != http.StatusOK {
			t.Errorf("expected 1st request OK, got %d", rr1.Code)
		}

		// 2nd request -> 429
		rr2 := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr2, req)
		if rr2.Code != http.StatusTooManyRequests {
			t.Errorf("expected 429 Too Many Requests, got %d", rr2.Code)
		}
	})

	t.Run("invalid api key", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "legacykey", nil
			},
		}
		storageClient := storage.NewMockClient()

		authMiddleware := NewAuthMiddleware(secretsClient, storageClient)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-KEY", "wrongkey")
		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})
}

func TestLogging(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	Logging(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}
}
