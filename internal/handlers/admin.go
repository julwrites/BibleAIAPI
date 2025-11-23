package handlers

import (
	"bible-api-service/internal/secrets"
	"bible-api-service/internal/storage"
	"bible-api-service/internal/util"
	"encoding/json"
	"log"
	"net/http"
)

type AdminHandler struct {
	StorageClient storage.Client
	SecretsClient secrets.Client
}

func NewAdminHandler(storageClient storage.Client, secretsClient secrets.Client) *AdminHandler {
	return &AdminHandler{
		StorageClient: storageClient,
		SecretsClient: secretsClient,
	}
}

type CreateKeyRequest struct {
	Password   string `json:"password"`
	ClientName string `json:"client_name"`
	DailyLimit int    `json:"daily_limit"`
}

func (h *AdminHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ClientName == "" {
		util.JSONError(w, http.StatusBadRequest, "Client name is required")
		return
	}
	if req.Password == "" {
		util.JSONError(w, http.StatusUnauthorized, "Password is required")
		return
	}

	// Verify Password
	adminPassword, err := h.SecretsClient.GetSecret(r.Context(), "ADMIN_PASSWORD")
	if err != nil {
		log.Printf("Failed to retrieve ADMIN_PASSWORD: %v", err)
		util.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if req.Password != adminPassword {
		// Log the attempt (optional security audit)
		log.Printf("Failed admin login attempt for client '%s'", req.ClientName)
		util.JSONError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	// Default limit if not set
	limit := req.DailyLimit
	if limit <= 0 {
		limit = 1000 // Default to 1000 requests per day
	}

	apiKey, err := h.StorageClient.CreateAPIKey(r.Context(), req.ClientName, limit)
	if err != nil {
		log.Printf("Failed to create API key: %v", err)
		util.JSONError(w, http.StatusInternalServerError, "Failed to create API key")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiKey)
}
