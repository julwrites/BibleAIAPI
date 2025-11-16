package secrets

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	mock := &mockSecretManagerClient{
		closeFunc: func() error { return nil },
	}
	client := New(mock, "test-project")
	if client == nil {
		t.Error("expected client to not be nil")
	}
}

func TestNewClient(t *testing.T) {
	// This test verifies that the NewClient function falls back to the
	// environment variable client when it fails to create a Google Secret
	// Manager client. This is the expected behavior in a local test
	// environment where credentials are not available.
	client, err := NewClient(context.Background(), "test-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := client.(*EnvClient); !ok {
		t.Errorf("expected an env client, but got %T", client)
	}
}
