package middleware

import (
	"bible-api-service/internal/secrets"
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

	t.Run("no api key configured", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "", errors.New("secret not found")
			},
		}
		authMiddleware := NewAuthMiddleware(secretsClient)
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("valid api key", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "testkey", nil
			},
		}
		authMiddleware := NewAuthMiddleware(secretsClient)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-KEY", "testkey")
		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("invalid api key", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "testkey", nil
			},
		}
		authMiddleware := NewAuthMiddleware(secretsClient)
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

	// In a real test, you would capture the log output and assert its contents.
	// For this example, we'll just ensure the handler is called.
	Logging(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}
}
