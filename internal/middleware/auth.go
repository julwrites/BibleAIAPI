package middleware

import (
	"bible-api-service/internal/secrets"
	"bible-api-service/internal/util"
	"context"
	"log"
	"net/http"
)

type AuthMiddleware struct {
	secretsClient secrets.Client
}

func NewAuthMiddleware(secretsClient secrets.Client) *AuthMiddleware {
	return &AuthMiddleware{secretsClient: secretsClient}
}

func (m *AuthMiddleware) APIKeyAuth(next http.Handler) http.Handler {
	apiKey, err := m.secretsClient.GetSecret(context.Background(), "API_KEY")
	if err != nil {
		log.Printf("could not get API key from secret manager: %v", err)
		// If the secret is not found, bypass the check.
		// This is useful for local development.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientKey := r.Header.Get("X-API-KEY")
		if clientKey != apiKey {
			util.JSONError(w, http.StatusUnauthorized, "Invalid API Key")
			return
		}

		next.ServeHTTP(w, r)
	})
}
