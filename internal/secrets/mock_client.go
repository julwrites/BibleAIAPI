package secrets

import (
	"context"
	"fmt"
	"os"
)

// MockClient is a mock implementation of the secrets.Client interface.
type MockClient struct{}

// GetSecret retrieves a secret from an environment variable.
func (c *MockClient) GetSecret(ctx context.Context, name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("secret %s not found in environment", name)
	}
	return value, nil
}
