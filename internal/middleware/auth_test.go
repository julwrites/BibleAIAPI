package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"bible-api-service/internal/secrets"
)

func TestAPIKeyAuth(t *testing.T) {
	tests := []struct {
		name           string
		apiKeyInSecret string
		apiKeyInHeader string
		secretError    error
		env            string
		expectedStatus int
	}{
		{
			name:           "Valid API Key",
			apiKeyInSecret: "test-key",
			apiKeyInHeader: "test-key",
			secretError:    nil,
			env:            "production",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid API Key",
			apiKeyInSecret: "test-key",
			apiKeyInHeader: "wrong-key",
			secretError:    nil,
			env:            "production",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing API Key",
			apiKeyInSecret: "test-key",
			apiKeyInHeader: "",
			secretError:    nil,
			env:            "production",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Secret Manager Error in Prod",
			apiKeyInSecret: "",
			apiKeyInHeader: "any-key",
			secretError:    errors.New("secret not found"),
			env:            "production",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Secret Manager Error in Dev",
			apiKeyInSecret: "",
			apiKeyInHeader: "any-key",
			secretError:    errors.New("secret not found"),
			env:            "development",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ENV", tt.env)
			defer os.Unsetenv("ENV")

			mockSecretsClient := &secrets.MockClient{
				GetSecretFunc: func(ctx context.Context, name string) (string, error) {
					if tt.secretError != nil {
						return "", tt.secretError
					}
					return tt.apiKeyInSecret, nil
				},
			}

			authMiddleware := NewAuthMiddleware(mockSecretsClient)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/", nil)
			if tt.apiKeyInHeader != "" {
				req.Header.Set("X-API-KEY", tt.apiKeyInHeader)
			}
			rr := httptest.NewRecorder()

			authMiddleware.APIKeyAuth(handler).ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
