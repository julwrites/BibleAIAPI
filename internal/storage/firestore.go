package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FirestoreClient struct {
	client *firestore.Client
}

// NewFirestoreClient creates a new Firestore client.
func NewFirestoreClient(ctx context.Context, projectID string) (Client, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}
	return &FirestoreClient{client: client}, nil
}

func (c *FirestoreClient) Close() error {
	return c.client.Close()
}

func (c *FirestoreClient) CreateAPIKey(ctx context.Context, clientName string, dailyLimit int) (*APIKey, error) {
	key, err := generateRandomKey()
	if err != nil {
		return nil, err
	}

	apiKey := &APIKey{
		Key:           key,
		ClientName:    clientName,
		CreatedAt:     time.Now(),
		Active:        true,
		DailyLimit:    dailyLimit,
		LastUsageDate: time.Now().Format("2006-01-02"),
		RequestCount:  0,
	}

	_, err = c.client.Collection("api_keys").Doc(key).Create(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create api key doc: %w", err)
	}

	return apiKey, nil
}

func (c *FirestoreClient) GetAPIKey(ctx context.Context, key string) (*APIKey, error) {
	doc, err := c.client.Collection("api_keys").Doc(key).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var apiKey APIKey
	if err := doc.DataTo(&apiKey); err != nil {
		return nil, fmt.Errorf("failed to parse api key data: %w", err)
	}

	return &apiKey, nil
}

func (c *FirestoreClient) IncrementUsage(ctx context.Context, key string) (int, bool, error) {
	var currentCount int
	var limitExceeded bool

	err := c.client.RunTransaction(ctx, func(ctx context.Context, t *firestore.Transaction) error {
		docRef := c.client.Collection("api_keys").Doc(key)
		doc, err := t.Get(docRef)
		if err != nil {
			return err
		}

		var apiKey APIKey
		if err := doc.DataTo(&apiKey); err != nil {
			return err
		}

		if !apiKey.Active {
			limitExceeded = true
			return nil
		}

		today := time.Now().Format("2006-01-02")
		if apiKey.LastUsageDate != today {
			apiKey.LastUsageDate = today
			apiKey.RequestCount = 0
		}

		if apiKey.RequestCount >= apiKey.DailyLimit {
			limitExceeded = true
			currentCount = apiKey.RequestCount
			return nil // Do not update
		}

		apiKey.RequestCount++
		currentCount = apiKey.RequestCount
		limitExceeded = false

		return t.Set(docRef, apiKey)
	})

	if err != nil {
		return 0, false, err
	}

	return currentCount, limitExceeded, nil
}

func generateRandomKey() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
