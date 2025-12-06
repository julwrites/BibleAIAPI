package middleware

import (
	"bible-api-service/internal/secrets"
	"bible-api-service/internal/util"
	"log"
	"net/http"
)

type AuthMiddleware struct {
	secretsClient secrets.Client
}

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

		// Check API Key against secret
		apiKey, err := m.secretsClient.GetSecret(ctx, "API_KEY")
		if err != nil {
			log.Printf("could not get API key from secret manager: %v", err)
			log.Println("Bypassing auth because API_KEY secret could not be retrieved (Local Dev Mode)")
			next.ServeHTTP(w, r)
			return
		}

		if clientKey == apiKey {
			next.ServeHTTP(w, r)
			return
		}

		util.JSONError(w, http.StatusUnauthorized, "Invalid API Key")
	})
}
