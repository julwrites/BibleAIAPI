package storage

import (
	"context"
	"time"
)

// APIKey represents an API key document.
type APIKey struct {
	Key           string    `firestore:"key" json:"key"`
	ClientName    string    `firestore:"client_name" json:"client_name"`
	CreatedAt     time.Time `firestore:"created_at" json:"created_at"`
	Active        bool      `firestore:"active" json:"active"`
	DailyLimit    int       `firestore:"daily_limit" json:"daily_limit"`
	LastUsageDate string    `firestore:"last_usage_date" json:"last_usage_date"` // YYYY-MM-DD
	RequestCount  int       `firestore:"request_count" json:"request_count"`
}

// Client defines the interface for storage operations.
type Client interface {
	// CreateAPIKey creates a new API key.
	CreateAPIKey(ctx context.Context, clientName string, dailyLimit int) (*APIKey, error)

	// GetAPIKey retrieves an API key by its string value.
	// Returns nil, nil if not found.
	GetAPIKey(ctx context.Context, key string) (*APIKey, error)

	// IncrementUsage increments the usage counter for the key.
	// It handles resetting the counter if the date has changed.
	// Returns the new count and whether the limit has been exceeded.
	IncrementUsage(ctx context.Context, key string) (int, bool, error)

	// Close closes the client connection.
	Close() error
}
