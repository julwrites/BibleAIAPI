package middleware

import (
	"bible-api-service/internal/secrets"
	"bible-api-service/internal/util"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type AuthMiddleware struct {
	secretsClient secrets.Client
}

type contextKey string

const ClientIDKey contextKey = "ClientID"

func NewAuthMiddleware(secretsClient secrets.Client) *AuthMiddleware {
	return &AuthMiddleware{
		secretsClient: secretsClient,
	}
}

func (m *AuthMiddleware) APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientKey := r.Header.Get("X-API-KEY")
		if clientKey == "" {
			util.JSONError(w, http.StatusUnauthorized, "Missing API Key")
			return
		}

		ctx := r.Context()

		// Check API Keys against secret
		// Secret is expected to be a JSON object: {"client_id": "api_key", ...}
		secretVal, err := m.secretsClient.GetSecret(ctx, "API_KEYS")
		if err != nil {
			log.Printf("could not get API_KEYS from secret manager: %v", err)
			log.Println("Bypassing auth because API_KEYS secret could not be retrieved (Local Dev Mode)")
			next.ServeHTTP(w, r)
			return
		}

		var apiKeys map[string]string
		if err := json.Unmarshal([]byte(secretVal), &apiKeys); err != nil {
			log.Printf("failed to parse API_KEYS secret as JSON: %v", err)
			util.JSONError(w, http.StatusInternalServerError, "Internal Authentication Configuration Error")
			return
		}

		for clientID, key := range apiKeys {
			if clientKey == key {
				log.Printf("Authenticated request from client: %s", clientID)
				ctx = context.WithValue(ctx, ClientIDKey, clientID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		util.JSONError(w, http.StatusUnauthorized, "Invalid API Key")
	})
}
