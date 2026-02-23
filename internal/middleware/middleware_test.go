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
	// Handler that echos the ClientID from context into a header for verification
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if val := r.Context().Value(ClientIDKey); val != nil {
			if clientID, ok := val.(string); ok {
				w.Header().Set("X-Authenticated-Client", clientID)
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	t.Run("no api key configured (missing header)", func(t *testing.T) {
		// Even if secret exists, missing header -> 401
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return `{"client": "key"}`, nil
			},
		}

		authMiddleware := NewAuthMiddleware(secretsClient)
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})

	t.Run("bypass local dev if secret fetch fails", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "", errors.New("secret not found")
			},
		}

		authMiddleware := NewAuthMiddleware(secretsClient)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-KEY", "anything")
		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("valid api key for client A", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return `{"clientA": "keyA", "clientB": "keyB"}`, nil
			},
		}

		authMiddleware := NewAuthMiddleware(secretsClient)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-KEY", "keyA")
		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}
		if client := rr.Header().Get("X-Authenticated-Client"); client != "clientA" {
			t.Errorf("expected clientA, got %s", client)
		}
	})

	t.Run("valid api key for client B", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return `{"clientA": "keyA", "clientB": "keyB"}`, nil
			},
		}

		authMiddleware := NewAuthMiddleware(secretsClient)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-KEY", "keyB")
		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}
		if client := rr.Header().Get("X-Authenticated-Client"); client != "clientB" {
			t.Errorf("expected clientB, got %s", client)
		}
	})

	t.Run("invalid api key", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return `{"clientA": "keyA"}`, nil
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

	t.Run("malformed secret json", func(t *testing.T) {
		secretsClient := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return `not json`, nil
			},
		}

		authMiddleware := NewAuthMiddleware(secretsClient)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-KEY", "anything")
		rr := httptest.NewRecorder()
		authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)

		// Should return 500 because it's a server config error
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
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
