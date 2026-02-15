package mocks

import (
	"context"
	"time"
)

// MockLLMClient is a mock implementation of provider.LLMClient
type MockLLMClient struct {
	Response     string
	Err          error
	Delay        time.Duration
	QueryCalled  bool
	LastPrompt   string
	LastSchema   string
	ProviderName string
}

func (m *MockLLMClient) Name() string {
	if m.ProviderName == "" {
		return "mock"
	}
	return m.ProviderName
}

func (m *MockLLMClient) Query(ctx context.Context, prompt string, schema string) (string, string, error) {
	m.QueryCalled = true
	m.LastPrompt = prompt
	m.LastSchema = schema

	if m.Delay > 0 {
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		case <-time.After(m.Delay):
		}
	}

	if m.Err != nil {
		return "", m.Name(), m.Err
	}
	return m.Response, m.Name(), nil
}

func (m *MockLLMClient) Stream(ctx context.Context, prompt string) (<-chan string, string, error) {
	// Not implemented for now, return error if called
	return nil, m.Name(), nil
}
