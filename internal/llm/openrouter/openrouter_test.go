package openrouter

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockSecretsClient struct {
	getSecretFunc func(ctx context.Context, name string) (string, error)
}

func (m *mockSecretsClient) GetSecret(ctx context.Context, name string) (string, error) {
	return m.getSecretFunc(ctx, name)
}

func TestNewClient(t *testing.T) {
	ctx := context.Background()

	t.Run("No API key", func(t *testing.T) {
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "", errors.New("not found")
			},
		}

		client, err := NewClient(ctx, mockSecrets, "")
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("API key set", func(t *testing.T) {
		mockSecrets := &mockSecretsClient{
			getSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "OPENROUTER_API_KEY" {
					return "test-key", nil
				}
				return "", nil
			},
		}

		client, err := NewClient(ctx, mockSecrets, "")
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "openrouter", client.Name())
	})
}
