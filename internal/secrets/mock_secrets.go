package secrets

import "context"

// MockClient is a mock implementation of the secrets client for testing.
type MockClient struct {
	GetSecretFunc func(ctx context.Context, name string) (string, error)
}

// GetSecret calls the mock's GetSecretFunc.
func (m *MockClient) GetSecret(ctx context.Context, name string) (string, error) {
	return m.GetSecretFunc(ctx, name)
}
