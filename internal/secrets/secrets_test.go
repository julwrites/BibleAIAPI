package secrets

import (
	"context"
	"errors"
	"testing"
)

type MockClient struct {
	GetSecretFunc func(ctx context.Context, name string) (string, error)
}

func (m *MockClient) GetSecret(ctx context.Context, name string) (string, error) {
	if m.GetSecretFunc != nil {
		return m.GetSecretFunc(ctx, name)
	}
	return "", nil
}

func TestGet(t *testing.T) {
	t.Run("Secret found", func(t *testing.T) {
		mock := &MockClient{
			GetSecretFunc: func(ctx context.Context, name string) (string, error) {
				if name == "test-key" {
					return "test-value", nil
				}
				return "", errors.New("not found")
			},
		}

		val, err := Get(context.Background(), mock, "test-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "test-value" {
			t.Errorf("expected 'test-value', got '%s'", val)
		}
	})

	t.Run("Secret not found", func(t *testing.T) {
		mock := &MockClient{
			GetSecretFunc: func(ctx context.Context, name string) (string, error) {
				return "", errors.New("not found")
			},
		}

		_, err := Get(context.Background(), mock, "test-key")
		if err == nil {
			t.Error("expected error for missing secret")
		}
	})
}
