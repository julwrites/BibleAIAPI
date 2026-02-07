package mocks

import (
	"context"
	"time"
)

// MockLLMClient is a mock implementation of provider.LLMClient
type MockLLMClient struct {
	Response      string
	Err           error
	Delay         time.Duration
	QueryCalled   bool
	LastPrompt    string
	LastSchema    string
}

func (m *MockLLMClient) Query(ctx context.Context, prompt string, schema string) (string, error) {
	m.QueryCalled = true
	m.LastPrompt = prompt
	m.LastSchema = schema

	if m.Delay > 0 {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(m.Delay):
		}
	}

	if m.Err != nil {
		return "", m.Err
	}
	return m.Response, nil
}
