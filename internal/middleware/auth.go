package middleware

import (
	"bible-api-service/internal/secrets"
	"bible-api-service/internal/storage"
	"bible-api-service/internal/util"
	"log"
	"net/http"
)

type AuthMiddleware struct {
	secretsClient secrets.Client
	storageClient storage.Client
}

func NewAuthMiddleware(secretsClient secrets.Client, storageClient storage.Client) *AuthMiddleware {
	return &AuthMiddleware{
		secretsClient: secretsClient,
		storageClient: storageClient,
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

		// 1. Check Firestore for the key
		apiKey, err := m.storageClient.GetAPIKey(ctx, clientKey)
		if err != nil {
			log.Printf("Error looking up API key in storage: %v", err)
		}

		if apiKey != nil {
			if !apiKey.Active {
				util.JSONError(w, http.StatusUnauthorized, "API Key is inactive")
				return
			}

			currentUsage, limitExceeded, err := m.storageClient.IncrementUsage(ctx, clientKey)
			if err != nil {
				log.Printf("Error incrementing usage for key %s: %v", clientKey, err)
				util.JSONError(w, http.StatusInternalServerError, "Internal Server Error")
				return
			}

			if limitExceeded {
				log.Printf("Rate limit exceeded for client %s (Usage: %d/%d)", apiKey.ClientName, currentUsage, apiKey.DailyLimit)
				util.JSONError(w, http.StatusTooManyRequests, "Rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		// 2. Fallback: Check Legacy API Key
		legacyKey, err := m.secretsClient.GetSecret(ctx, "API_KEY")
		if err != nil {
			log.Printf("could not get legacy API key from secret manager: %v", err)
			log.Println("Bypassing auth because API_KEY secret could not be retrieved (Local Dev Mode)")
			next.ServeHTTP(w, r)
			return
		}

		if clientKey == legacyKey {
			next.ServeHTTP(w, r)
			return
		}

		util.JSONError(w, http.StatusUnauthorized, "Invalid API Key")
	})
}
