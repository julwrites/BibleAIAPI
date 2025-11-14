package middleware

import (
	"bible-api-service/internal/util"
	"net/http"
	"os"
)

func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("API_KEY")
		if apiKey == "" {
			// If no API key is configured on the server, bypass the check.
			// This is useful for local development.
			next.ServeHTTP(w, r)
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
