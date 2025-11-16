package secrets

import (
	"context"
	"fmt"
	"os"
)

// EnvClient is a client that retrieves secrets from environment variables.
type EnvClient struct{}

// GetSecret retrieves a secret from an environment variable.
func (c *EnvClient) GetSecret(_ context.Context, name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("secret %q not found in environment", name)
	}
	return value, nil
}
