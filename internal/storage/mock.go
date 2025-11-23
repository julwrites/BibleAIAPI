package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type MockClient struct {
	mu   sync.Mutex
	Keys map[string]*APIKey
}

func NewMockClient() *MockClient {
	return &MockClient{
		Keys: make(map[string]*APIKey),
	}
}

func (m *MockClient) Close() error {
	return nil
}

func (m *MockClient) CreateAPIKey(ctx context.Context, clientName string, dailyLimit int) (*APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("mock-key-%d", len(m.Keys)+1)
	apiKey := &APIKey{
		Key:           key,
		ClientName:    clientName,
		CreatedAt:     time.Now(),
		Active:        true,
		DailyLimit:    dailyLimit,
		LastUsageDate: time.Now().Format("2006-01-02"),
		RequestCount:  0,
	}
	m.Keys[key] = apiKey
	return apiKey, nil
}

func (m *MockClient) GetAPIKey(ctx context.Context, key string) (*APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if k, ok := m.Keys[key]; ok {
		// Return a copy to mimic DB behavior
		copyKey := *k
		return &copyKey, nil
	}
	return nil, nil
}

func (m *MockClient) IncrementUsage(ctx context.Context, key string) (int, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	apiKey, ok := m.Keys[key]
	if !ok {
		return 0, false, fmt.Errorf("key not found")
	}

	if !apiKey.Active {
		return apiKey.RequestCount, true, nil
	}

	today := time.Now().Format("2006-01-02")
	if apiKey.LastUsageDate != today {
		apiKey.LastUsageDate = today
		apiKey.RequestCount = 0
	}

	if apiKey.RequestCount >= apiKey.DailyLimit {
		return apiKey.RequestCount, true, nil
	}

	apiKey.RequestCount++
	m.Keys[key] = apiKey // Update map

	return apiKey.RequestCount, false, nil
}
