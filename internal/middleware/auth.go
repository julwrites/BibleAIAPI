package middleware

import (
	"bible-api-service/internal/secrets"
	"bible-api-service/internal/util"
	"log"
	"net/http"
	"os"
)

type AuthMiddleware struct {
	secretsClient secrets.Client
}

func NewAuthMiddleware(secretsClient secrets.Client) *AuthMiddleware {
	return &AuthMiddleware{secretsClient: secretsClient}
}

func (m *AuthMiddleware) APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := m.secretsClient.GetSecret(r.Context(), "API_KEY")
		if err != nil {
			log.Printf("could not get API key from secret manager: %v", err)
			// If the secret is not found, bypass the check in a non-production environment.
			// This is useful for local development.
			if os.Getenv("ENV") != "production" {
				next.ServeHTTP(w, r)
				return
			}
			util.JSONError(w, http.StatusInternalServerError, "Could not verify API key")
			return
		}

		clientKey := r.Header.Get("X-API-KEY")
		if clientKey != apiKey {
			util.JSONError(w, http.StatusUnauthorized, "Invalid API Key")
			return
		}

		next.ServeHTTP(w, r)
	})
}
