package secrets

import (
	"context"
	"os"
	"testing"
)

func TestEnvClient_GetSecret(t *testing.T) {
	t.Run("Secret found", func(t *testing.T) {
		os.Setenv("TEST_SECRET", "test-value")
		defer os.Unsetenv("TEST_SECRET")
		client := &EnvClient{}
		secret, err := client.GetSecret(context.Background(), "TEST_SECRET")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if secret != "test-value" {
			t.Errorf("unexpected secret value: got %q, want %q", secret, "test-value")
		}
	})

	t.Run("Secret not found", func(t *testing.T) {
		client := &EnvClient{}
		_, err := client.GetSecret(context.Background(), "NON_EXISTENT_SECRET")
		if err == nil {
			t.Error("expected an error but got nil")
		}
	})
}
